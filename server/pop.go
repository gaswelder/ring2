package server

import (
	"log"
	"net"
)

func processPOP(conn net.Conn, server *Server) {
	s := newPopSession(conn, server)
	s.OK("Hello")

	for {
		cmd, err := s.ReadCommand()
		if err != nil {
			s.Err(err.Error())
			continue
		}

		if cmd.Name == "QUIT" {
			err = s.commit()
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
	conn.Close()
}
