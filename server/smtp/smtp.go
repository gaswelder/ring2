package smtp

import (
	"io"
	"log"
	"os"

	"github.com/gaswelder/ring2/server/mailbox"
)

type cmdFunc func(s *session, cmd *Command)

var commands = map[string]cmdFunc{
	"HELO": cmdHelo,
	"EHLO": cmdEhlo,
	"RSET": cmdRset,
	"MAIL": cmdMail,
	"RCPT": cmdRcpt,
	"DATA": cmdData,
}

// Extensions are registered separately because they are listed
// by the EHLO command.
var smtpExts = map[string]cmdFunc{
	"HELP": cmdHelp,
	"AUTH": cmdAuth,
}

const AuthOK = 235
const ParameterSyntaxError = 501
const BadSequenceOfCommands = 503
const ParameterNotImplemented = 504
const AuthInvalid = 535

type AuthFunc func(name, password string) error
type MailboxLookupFunc func(name string) ([]*mailbox.Mailbox, error)

type session struct {
	*ReadWriter
	senderHost string
	draft      *Mail
	auth       bool
	authorize  AuthFunc
	lookup     MailboxLookupFunc
	recipients []*mailbox.Mailbox
}

func Process(conn io.ReadWriter, auth AuthFunc, lookup MailboxLookupFunc) {
	s := &session{
		ReadWriter: NewWriter(conn),
		authorize:  auth,
		lookup:     lookup,
		recipients: make([]*mailbox.Mailbox, 0),
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("couldn't get hostname: %s", err.Error())
		hostname = "localhost"
	}
	s.Send(220, "%s ready", hostname)

	/*
	 * Go allows to organize the processing in a linear manner, but the
	 * SMTP standard was written around implementations of that time
	 * which maintained explicit state and thus allowed different
	 * commands like "HELP" to be issued out of context.
	 *
	 * Therefore we read commands here and dispatch them to separate
	 * command functions, passing them a pointer to the current state.
	 */
	for {
		cmd, err := s.ReadCommand()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Send(500, err.Error())
			continue
		}

		if cmd.Name == "QUIT" {
			s.Send(221, "So long, Bob")
			break
		}

		f, ok := commands[cmd.Name]
		if !ok {
			f, ok = smtpExts[cmd.Name]
		}
		if !ok {
			s.Send(500, "Unknown command")
			continue
		}
		f(s, cmd)
	}
}
