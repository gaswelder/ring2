package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func processSMTP(conn net.Conn) {

	log.Printf("%s connected\n", conn.RemoteAddr().String())
	s := newSession(conn)
	s.send(220, "%s ready", config.hostname)

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
		line, err := s.r.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}

		fmt.Print(line)

		cmd, err := parseCommand(line)

		if err != nil {
			s.send(500, err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			s.send(221, "So long, Bob")
			break
		}

		if !processCmd(s, cmd) {
			s.send(500, "Unknown command")
		}
	}

	conn.Close()
	log.Printf("%s disconnected\n", conn.RemoteAddr().String())
}

func createDir(path string) error {

	stat, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		stat, err = os.Stat(config.maildir)
	}

	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}
