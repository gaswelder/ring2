package pop

import (
	"io"
	"log"

	"github.com/gaswelder/ring2/server/mailbox"
)

type AuthFunc func(name, password string) (*mailbox.Mailbox, error)

type popfunc func(s *session, c *command)

var popFuncs = map[string]popfunc{
	"USER": cmdUser,
	"PASS": cmdPass,
	"STAT": cmdStat,
	"LIST": cmdList,
	"RETR": cmdRetr,
	"DELE": cmdDele,
	"NOOP": cmdNoop,
	"LAST": cmdLast,
	"RSET": cmdRset,
	// Optional
	"UIDL": cmdUidl,
	"TOP":  cmdTop,
}

func Process(conn io.ReadWriter, auth AuthFunc) {
	s := makeSession(conn, auth)
	s.OK("Hello")
	for {
		cmd, err := s.readCommand()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Err(err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			if s.inbox == nil {
				s.OK("")
				break
			}

			err = s.inbox.commit()
			if err != nil {
				log.Println(err)
				s.Err(err.Error())
			} else {
				s.OK("")
			}
			break
		}

		cmdfunc, ok := popFuncs[cmd.name]
		if !ok {
			s.Err("Unknown command")
			continue
		}
		cmdfunc(s, cmd)
	}
}
