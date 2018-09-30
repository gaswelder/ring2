package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/gaswelder/ring2/scanner"
	"github.com/gaswelder/ring2/server/smtp"
)

/*
 * Command functions register
 */
type cmdFunc func(s *session, cmd *smtp.Command)

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
func processCmd(s *session, cmd *smtp.Command) bool {
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
	defineCmd("HELO", func(s *session, cmd *smtp.Command) {
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
	defineCmd("EHLO", func(s *session, cmd *smtp.Command) {
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
	defineCmd("RSET", func(s *session, cmd *smtp.Command) {
		s.draft = nil
		s.Send(250, "OK")
	})

	/*
	 * MAIL FROM:<path>[ <params>]
	 */
	defineCmd("MAIL", func(s *session, cmd *smtp.Command) {

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
		rpath, err := smtp.ParsePath(p)
		if err != nil {
			s.Send(501, "Malformed reverse-path")
			return
		}

		// If <params> part follows, read it,
		// but don't do anything with it
		if p.More() && p.Next() == ' ' {
			log.Println("MAIL params: " + p.Rest())
		}

		s.draft = smtp.NewDraft(rpath)
		s.Send(250, "OK")
	})

	/*
	 * RCPT TO:<path>
	 */
	defineCmd("RCPT", func(s *session, cmd *smtp.Command) {
		if s.draft == nil {
			s.Send(smtp.BadSequenceOfCommands, "Not in mail mode")
			return
		}

		p := scanner.New(cmd.Arg)
		if !p.SkipStri("TO:") {
			s.Send(smtp.ParameterSyntaxError, "The format is: RCPT TO:<forward-path>")
			return
		}

		path, err := smtp.ParsePath(p)
		if err != nil {
			s.Send(smtp.ParameterSyntaxError, "Malformed forward-path")
			return
		}

		if len(path.Hosts) > 0 {
			s.Send(551, "This server does not relay")
			return
		}

		if !strings.EqualFold(path.Addr.Host, s.config.Hostname) {
			s.Send(550, "Not a local address")
			return
		}

		_, ok1 := s.config.Lists[path.Addr.Name]
		_, ok2 := s.config.Users[path.Addr.Name]
		if !ok1 && !ok2 {
			s.Send(550, "Unknown Recipient: "+path.Addr.Format())
			return
		}
		s.Send(250, "OK")
		s.draft.Recipients = append(s.draft.Recipients, path)
	})

	/*
	 * Data
	 */
	defineCmd("DATA", func(s *session, cmd *smtp.Command) {

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
		 * Insert a stamp at the beginning of the message
		 * Example: Received: from GHI.ARPA by JKL.ARPA ; 27 Oct 81 15:27:39 PST
		 */
		text := fmt.Sprintf("Received: from %s by %s ; %s\r\n",
			s.senderHost, s.config.Hostname, formatDate())

		/*
		 * Read the message
		 */
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

		if processMessage(s.draft, text, s.config) {
			s.Send(250, "OK")
		} else {
			s.Send(554, "Transaction failed")
		}
	})

	defineCmd("VRFY", obsolete)

	smtpExt("HELP", func(s *session, cmd *smtp.Command) {
		s.Send(214, helpfulMessage())
	})

	// AUTH <type> <arg>
	smtpExt("AUTH", func(s *session, cmd *smtp.Command) {
		// Naively split the auth arguments by a single space.
		parts := strings.Split(cmd.Arg, " ")

		// Here we are prepared to deal only with the "PLAIN <...>"" case.
		if len(parts) != 2 || parts[0] != "PLAIN" {
			s.Send(smtp.ParameterNotImplemented, "Only PLAIN <...> is supported")
			return
		}

		// If already authorized, reject
		if s.user != nil {
			s.Send(smtp.BadSequenceOfCommands, "Already authorized")
			return
		}

		user, smtpErr := plainAuth(parts[1], s.config)
		if smtpErr != nil {
			s.Send(smtpErr.code, smtpErr.message)
			return
		}

		if user == nil {
			s.Send(smtp.AuthInvalid, "Authentication credentials invalid")
			return
		}

		s.user = user
		s.Send(smtp.AuthOK, "Authentication succeeded")
	})
}

type smtpError struct {
	code    int
	message string
}

func plainAuth(arg string, config *Config) (*UserRec, *smtpError) {
	// AGdhcwAxMjM= -> \0user\0pass
	data, err := base64.StdEncoding.DecodeString(arg)
	if err != nil {
		return nil, &smtpError{smtp.ParameterSyntaxError, err.Error()}
	}

	parts := strings.Split(string(data), "\x00")
	if len(parts) != 3 {
		return nil, &smtpError{smtp.ParameterSyntaxError, "Could not parse the auth string"}
	}

	login := parts[1]
	pass := parts[2]
	return config.findUser(login, pass), nil
}

/*
 * Command function for obsolete commands
 */
func obsolete(s *session, cmd *smtp.Command) {
	s.Send(502, "Obsolete command")
}

func processMessage(m *smtp.Mail, text string, config *Config) bool {

	ok := 0
	rpath := m.Sender
	var err error

	for _, fpath := range m.Recipients {
		err = dispatchMail(text, fpath.Addr.Name, rpath, config)

		/*
		 * If processing failed, send failure notification
		 * using the reverse-path.
		 */
		if err != nil {
			log.Println(err)

			if rpath == nil {
				log.Println("Processing failed and reverse path is null")
				continue
			}

			err = sendBounce(fpath, rpath, config)
		}

		if err != nil {
			log.Printf("Could not send failure notification: %e\n", err)
			continue
		}

		ok++
	}
	return ok > 0
}

func dispatchMail(text string, name string, rpath *smtp.Path, config *Config) error {
	log.Printf("Dispatch: %s\n", name)

	// If the destination address is a list, recurse into sending
	// to each of the participants.
	list, _ := config.Lists[name]
	if list != nil {
		for _, user := range list {
			err := dispatchMail(text, user.Name, rpath, config)
			if err != nil {
				log.Printf("List %s: dispatch to %s failed: %s", name, user.Name, err.Error())
			}
		}
		return nil
	}

	// For a user destination work as usual.
	user, ok := config.Users[name]
	if ok {
		box, err := config.mailbox(user)
		if err != nil {
			return err
		}
		line := fmt.Sprintf("Return-Path: %s\r\n", rpath.Format())
		return box.Add(line + text)
	}

	/*
	 * What then?
	 */
	return fmt.Errorf("Unhandled recipient: %s", name)
}
