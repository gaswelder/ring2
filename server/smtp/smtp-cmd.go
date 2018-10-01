package smtp

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gaswelder/ring2/scanner"
)

/*
 * Command functions register
 */
type cmdFunc func(s *session, cmd *Command)

var commands = make(map[string]cmdFunc)
var smtpExts = make(map[string]cmdFunc)

func defineCmd(name string, f cmdFunc) {
	commands[name] = f
}
func smtpExt(name string, f cmdFunc) {
	smtpExts[name] = f
}

/*
 * Call the corresponding function to process a command
 */
func processCmd(s *session, cmd *Command) bool {
	f, ok := commands[cmd.Name]
	if !ok {
		f, ok = smtpExts[cmd.Name]
	}
	if !ok {
		return false
	}
	f(s, cmd)
	return true
}

func init() {

	/*
	 * HELO <host>
	 */
	defineCmd("HELO", func(s *session, cmd *Command) {
		if cmd.Arg == "" {
			s.Send(501, "Argument expected")
			return
		}
		s.senderHost = cmd.Arg
		s.Send(250, "Go ahead, %s", cmd.Arg)
	})

	/*
	 * EHLO <host>
	 */
	defineCmd("EHLO", func(s *session, cmd *Command) {
		if cmd.Arg == "" {
			s.Send(501, "Argument expected")
			return
		}
		s.senderHost = cmd.Arg

		// Send greeting and a list of supported extensions
		w := s.BeginBatch(250)
		w.Send("Hello, %s", cmd.Arg)
		for name := range smtpExts {
			w.Send("%s", name)
		}
		w.End()
	})

	/*
	 * RSET - reset everything
	 */
	defineCmd("RSET", func(s *session, cmd *Command) {
		s.draft = nil
		s.Send(250, "OK")
	})

	/*
	 * MAIL FROM:<path>[ <params>]
	 */
	defineCmd("MAIL", func(s *session, cmd *Command) {

		if s.senderHost == "" {
			s.Send(503, "HELO expected")
			return
		}

		p := scanner.New(cmd.Arg)
		if !p.SkipStri("FROM:") {
			s.Send(501, "The format is: MAIL FROM:<reverse-path>[ <params>]")
			return
		}

		// Read the <path> part
		rpath, err := ParsePath(p)
		if err != nil {
			s.Send(501, "Malformed reverse-path")
			return
		}

		// If <params> part follows, read it,
		// but don't do anything with it
		if p.More() && p.Next() == ' ' {
			log.Println("MAIL params: " + p.Rest())
		}

		s.draft = NewDraft(rpath)
		s.Send(250, "OK")
	})

	/*
	 * RCPT TO:<path>
	 */
	defineCmd("RCPT", func(s *session, cmd *Command) {
		if s.draft == nil {
			s.Send(BadSequenceOfCommands, "Not in mail mode")
			return
		}

		p := scanner.New(cmd.Arg)
		if !p.SkipStri("TO:") {
			s.Send(ParameterSyntaxError, "The format is: RCPT TO:<forward-path>")
			return
		}

		path, err := ParsePath(p)
		if err != nil {
			s.Send(ParameterSyntaxError, "Malformed forward-path")
			return
		}

		if len(path.Hosts) > 0 {
			s.Send(551, "This server does not relay")
			return
		}

		mailboxes, err := s.lookup(path.Addr.Name)
		if err != nil {
			s.Send(550, err.Error())
			return
		}
		s.recipients = append(s.recipients, mailboxes...)

		s.Send(250, "OK")
		s.draft.Recipients = append(s.draft.Recipients, path)
	})

	/*
	 * Data
	 */
	defineCmd("DATA", func(s *session, cmd *Command) {

		if s.draft == nil {
			s.Send(503, "Not in mail mode")
			return
		}

		if len(s.draft.Recipients) == 0 {
			s.Send(503, "No recipients specified")
			return
		}

		s.Send(354, "Start mail input, terminate with a dot line (.)")

		/*
		 * Read the message
		 */
		text := ""
		for {
			line, err := s.ReadLine()
			if err != nil {
				log.Println(err)
				return
			}

			if line == ".\r\n" {
				break
			}

			/*
			 * If the line starts with a dot and there are other
			 * characters, remove the dot.
			 */
			if line[0] == '.' {
				line = line[1:]
			}

			text += line
		}

		/*
		 * Insert a stamp at the beginning of the message
		 * Example: Received: from GHI.ARPA by JKL.ARPA ; 27 Oct 81 15:27:39 PST
		 */
		hostname, err := os.Hostname()
		if err != nil {
			log.Printf("failed to get hostname: %s\n", err.Error())
			hostname = "localhost"
		}

		rpathLine := fmt.Sprintf("Return-Path: %s\r\n", s.draft.Sender.Format())
		receivedLine := fmt.Sprintf("Received: from %s by %s ; %s\r\n",
			s.senderHost, hostname, time.Now().Format(time.RFC822))

		text = rpathLine + receivedLine + text

		for _, mailbox := range s.recipients {
			err := mailbox.Add(text)
			if err != nil {
				s.Send(554, "Couldn't send to %s: %s", mailbox.Name(), err.Error())
				return
			}
		}
		s.Send(250, "OK")
	})

	defineCmd("VRFY", obsolete)

	smtpExt("HELP", func(s *session, cmd *Command) {
		s.Send(214, helpfulMessage())
	})

	// AUTH <type> <arg>
	smtpExt("AUTH", func(s *session, cmd *Command) {
		// Naively split the auth arguments by a single space.
		parts := strings.Split(cmd.Arg, " ")

		// Here we are prepared to deal only with the "PLAIN <...>"" case.
		if len(parts) != 2 || parts[0] != "PLAIN" {
			s.Send(ParameterNotImplemented, "Only PLAIN <...> is supported")
			return
		}

		// If already authorized, reject
		if s.auth {
			s.Send(BadSequenceOfCommands, "Already authorized")
			return
		}

		user, password, smtpErr := plainAuth(parts[1])
		if smtpErr != nil {
			s.Send(smtpErr.code, smtpErr.message)
			return
		}

		if s.authorize(user, password) != nil {
			s.Send(AuthInvalid, "Authentication credentials invalid")
			return
		}

		s.auth = true
		s.Send(AuthOK, "Authentication succeeded")
	})
}

type smtpError struct {
	code    int
	message string
}

func plainAuth(arg string) (login, pass string, serr *smtpError) {
	// AGdhcwAxMjM= -> \0user\0pass
	data, err := base64.StdEncoding.DecodeString(arg)
	if err != nil {
		return "", "", &smtpError{ParameterSyntaxError, err.Error()}
	}

	parts := strings.Split(string(data), "\x00")
	if len(parts) != 3 {
		return "", "", &smtpError{ParameterSyntaxError, "Could not parse the auth string"}
	}

	login = parts[1]
	pass = parts[2]
	return login, pass, nil
}

/*
 * Command function for obsolete commands
 */
func obsolete(s *session, cmd *Command) {
	s.Send(502, "Obsolete command")
}
