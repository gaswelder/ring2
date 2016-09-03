package main

import (
	"net"
	"log"
)

func main() {

	err := readConfig("conf")
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", config.listen);
	if err != nil {
		log.Fatal(err);
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go processClient(conn)
	}
}

func processClient(conn net.Conn) {

	state = newState(conn)
	s.send(220, "%s ready", __thisHost)

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
		cmd, err := s.readCommand()

		if err != nil {
			log.Println(err)
			break
		}

		if cmd.name == "QUIT" {
			s.send(250, "So long, Bob")
			break
		}

		if !processCommand(state, cmd) {
			s.send(500, "Unknown command")
		}
	}

	conn.Close()
}
