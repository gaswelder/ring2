package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gaswelder/ring2/scanner"
	"github.com/gaswelder/ring2/server/smtp"
)

/*
 * Command functions register
 */
type cmdFunc func(s *session, cmd *command)

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
func processCmd(s *session, cmd *command) bool {
	f, ok := commands[cmd.name]
	if !ok {
		f, ok = smtpExts[cmd.name]
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
	defineCmd("HELO", func(s *session, cmd *command) {
		if cmd.arg == "" {
			s.send(501, "Argument expected")
			return
		}
		s.senderHost = cmd.arg
		s.send(250, "Go ahead, %s", cmd.arg)
	})

	/*
	 * EHLO <host>
	 */
	defineCmd("EHLO", func(s *session, cmd *command) {
		if cmd.arg == "" {
			s.send(501, "Argument expected")
			return
		}
		s.senderHost = cmd.arg

		// Send greeting and a list of supported extensions
		w := s.begin(250)
		w.Send("Hello, %s", cmd.arg)
		for name := range smtpExts {
			w.Send("%s", name)
		}
		w.End()
	})

	/*
	 * RSET - reset everything
	 */
	defineCmd("RSET", func(s *session, cmd *command) {
		s.draft = nil
		s.send(250, "OK")
	})

	/*
	 * MAIL FROM:<path>[ <params>]
	 */
	defineCmd("MAIL", func(s *session, cmd *command) {

		if s.senderHost == "" {
			s.send(503, "HELO expected")
			return
		}

		p := scanner.New(cmd.arg)
		if !p.SkipStri("FROM:") {
			s.send(501, "The format is: MAIL FROM:<reverse-path>[ <params>]")
			return
		}

		// Read the <path> part
		rpath, err := parsePath(p)
		if err != nil {
			s.send(501, "Malformed reverse-path")
			return
		}

		// If <params> part follows, read it,
		// but don't do anything with it
		if p.More() && p.Next() == ' ' {
			log.Println("MAIL params: " + p.Rest())
		}

		s.draft = newDraft(rpath)
		s.send(250, "OK")
	})

	/*
	 * RCPT TO:<path>
	 */
	defineCmd("RCPT", func(s *session, cmd *command) {

		if s.draft == nil {
			s.send(503, "Not in mail mode")
			return
		}

		p := scanner.New(cmd.arg)
		if !p.SkipStri("TO:") {
			s.send(501, "The format is: RCPT TO:<forward-path>")
			return
		}

		path, err := parsePath(p)
		if err != nil {
			s.send(501, "Malformed forward-path")
			return
		}
		code, str := checkPath(path, s.server.config)
		s.send(code, str)
		if code >= 300 || code < 200 {
			return
		}
		s.draft.recipients = append(s.draft.recipients, path)
	})

	/*
	 * Data
	 */
	defineCmd("DATA", func(s *session, cmd *command) {

		if s.draft == nil {
			s.send(503, "Not in mail mode")
			return
		}

		if len(s.draft.recipients) == 0 {
			s.send(503, "No recipients specified")
			return
		}

		s.send(354, "Start mail input, terminate with a dot line (.)")

		/*
		 * Insert a stamp at the beginning of the message
		 * Example: Received: from GHI.ARPA by JKL.ARPA ; 27 Oct 81 15:27:39 PST
		 */
		text := fmt.Sprintf("Received: from %s by %s ; %s\r\n",
			s.senderHost, s.server.config.Hostname, formatDate())

		/*
		 * Read the message
		 */
		for {
			line, err := s.r.ReadString('\n')
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

		if processMessage(s.draft, text, s.server.config) {
			s.send(250, "OK")
		} else {
			s.send(554, "Transaction failed")
		}
	})

	defineCmd("VRFY", obsolete)

	smtpExt("HELP", func(s *session, cmd *command) {
		s.send(214, helpfulMessage())
	})

	// AUTH <type> <arg>
	smtpExt("AUTH", func(s *session, cmd *command) {
		// Naively split the auth arguments by a single space.
		parts := strings.Split(cmd.arg, " ")

		// Here we are prepared to deal only with the "PLAIN <...>"" case.
		if len(parts) != 2 || parts[0] != "PLAIN" {
			s.send(smtp.ParameterNotImplemented, "Only PLAIN <...> is supported")
			return
		}

		// If already authorized, reject
		if s.user != nil {
			s.send(smtp.BadSequenceOfCommands, "Already authorized")
			return
		}

		user, smtpErr := plainAuth(parts[1], s.server)
		if smtpErr != nil {
			s.send(smtpErr.code, smtpErr.message)
			return
		}

		if user == nil {
			s.send(smtp.AuthInvalid, "Authentication credentials invalid")
			return
		}

		s.user = user
		s.send(smtp.AuthOK, "Authentication succeeded")
	})
}

type smtpError struct {
	code    int
	message string
}

func plainAuth(arg string, server *Server) (*UserRec, *smtpError) {
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
	return server.findUser(login, pass), nil
}

/*
 * Command function for obsolete commands
 */
func obsolete(s *session, cmd *command) {
	s.send(502, "Obsolete command")
}

func splitAddress(addr string) (name string, host string, err error) {
	pos := strings.Index(addr, "@")
	if pos < 0 {
		err = errors.New("Invalid email address")
	}
	name = addr[:pos]
	host = addr[pos+1:]
	return
}

func checkPath(p *path, config *Config) (int, string) {
	if len(p.hosts) > 0 {
		return 551, "This server does not relay"
	}

	name, host, err := splitAddress(p.address)
	if err != nil {
		return 501, err.Error()
	}
	if !strings.EqualFold(host, config.Hostname) {
		return 550, "Not a local address"
	}

	_, ok := config.Lists[name]
	if ok {
		return 250, "OK"
	}

	_, ok = config.Users[name]
	if ok {
		return 250, "OK"
	}

	return 550, "Unknown Recipient"
}

func processMessage(m *mail, text string, config *Config) bool {

	ok := 0
	rpath := m.sender
	var err error

	for _, fpath := range m.recipients {
		var name string
		name, _, err = splitAddress(fpath.address)
		if err == nil {
			err = dispatchMail(text, name, rpath, config)
		}
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

func dispatchMail(text string, name string, rpath *path, config *Config) error {
	log.Printf("Dispatch: %s\n", name)
	/*
	 * If it a user?
	 */
	user, ok := config.Users[name]
	if ok {
		return storeMessage(text, rpath, user, config)
	}

	/*
	 * A list?
	 */
	list, _ := config.Lists[name]
	if list != nil {
		ok := false
		for _, user := range list {
			err := dispatchMail(text, user.Name, rpath, config)
			if err == nil {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("Dispatch to list '%s' failed", name)
		}
		return nil
	}

	/*
	 * What then?
	 */
	return fmt.Errorf("Unhandled recipient: %s", name)
}

/*
 * Store a message locally
 */
func storeMessage(text string, rpath *path, u *UserRec, config *Config) error {
	box, err := u.mailbox(config)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("Return-Path: %s\r\n", formatPath(rpath))
	err = box.Add(line + text)
	return err
}

func sendBounce(fpath, rpath *path, config *Config) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Date: %s\r\n", formatDate())
	fmt.Fprintf(&b, "From: ring2@%s\r\n", config.Hostname)
	fmt.Fprintf(&b, "To: %s\r\n", rpath.address)
	fmt.Fprintf(&b, "Subject: mail delivery failure\r\n")
	fmt.Fprintf(&b, "\r\n")
	fmt.Fprintf(&b, "Sorry, your mail could not be delivered to %s", fpath.address)
	/*
	 * Specify null as reverse-path to prevent loops
	 */
	return sendMail(b.String(), rpath, nil, config)
}

/*
 * Send a message using SMTP protocol
 */
func sendMail(text string, fpath, rpath *path, config *Config) error {

	if len(fpath.hosts) == 0 {
		return errors.New("Empty forward-path")
	}

	conn, err := net.Dial("tcp", fpath.hosts[0])
	if err != nil {
		return err
	}

	w := newTpClient(conn)
	w.Expect(250)

	w.WriteLine("HELO %s", config.Hostname)
	w.Expect(250)

	if rpath != nil {
		w.WriteLine("MAIL FROM:" + formatPath(rpath))
	} else {
		w.WriteLine("MAIL FROM:<>")
	}
	w.Expect(250)

	w.WriteLine("RCPT TO:" + formatPath(fpath))
	w.Expect(250)

	w.WriteLine("DATA")
	w.Expect(354)

	lines := strings.Split(text, "\r\n")
	for _, line := range lines {
		/*
		 * If the first character of a line is a dot,
		 * insert one additional dot.
		 */
		if len(line) > 0 && line[0] == '.' {
			w.Write(".")
		}
		w.WriteLine(line)
	}

	w.WriteLine(".")
	w.Expect(250)

	w.WriteLine("QUIT")
	w.Expect(221)

	conn.Close()

	return w.Err()
}
