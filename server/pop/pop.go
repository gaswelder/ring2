package pop

import (
	"io"
	"log"
)

type Session = interface {
	ReadCommand() (*Command, error)
	Err(string)
	OK(fmt string, args ...interface{})
	Send(fmt string, args ...interface{}) error
	SendData(string) error
	SetUserName(string) error
	Open(password string) error
	Inbox() *InboxView
	Close() error
}

func Process(s Session) {
	s.OK("Hello")
	for {
		cmd, err := s.ReadCommand()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Err(err.Error())
			continue
		}

		if cmd.Name == "QUIT" {
			err = s.Close()
			if err != nil {
				log.Println(err)
				s.Err(err.Error())
			} else {
				s.OK("")
			}
			break
		}

		cmdfunc, ok := popFuncs[cmd.Name]
		if !ok {
			s.Err("Unknown command")
			continue
		}
		cmdfunc(s, cmd)
	}
}
