package main

import (
	"net"
	"log"
)

func processPOP(conn net.Conn) {
	s := newPopSession(conn)
	s.ok("Hello")

	for {
		line, err := s.r.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}

		cmd, err := parseCommand(line)
		if err != nil {
			s.err(err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			s.ok("")
			break
		}

		if !execPopCmd(s, cmd) {
			s.err("Unknown command")
		}
	}

	if s.box != nil {
		s.box.setLast(s.lastId)
		err := s.box.purge()
		if err != nil {
			log.Println(err)
		}
		s.box.unlock()
	}
	conn.Close()
}
