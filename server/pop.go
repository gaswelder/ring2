package server

import (
	"log"
	"net"
)

func processPOP(conn net.Conn, server *Server) {
	s := newPopSession(conn, server)
	s.ok("Hello")

	for {
		line, err := s.r.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		debMsg(line)

		cmd, err := parseCommand(line)
		if err != nil {
			s.err(err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			err = s.commit()
			if err != nil {
				log.Println(err)
				s.err(err.Error())
			} else {
				s.ok("")
			}
			break
		}

		if !execPopCmd(s, cmd) {
			s.err("Unknown command")
		}
	}
	conn.Close()
}
