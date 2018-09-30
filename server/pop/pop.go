package pop

import (
	"io"
	"log"
)

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

		if !execPopCmd(s, cmd) {
			s.Err("Unknown command")
		}
	}
}
