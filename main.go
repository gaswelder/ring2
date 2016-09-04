package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {

	err := readConfig("conf")
	if err != nil {
		log.Fatal(err)
	}

	err = createDir(config.spooldir)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", config.listen)
	if err != nil {
		log.Fatal(err)
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

		cmd, err := parseCommand(line)

		if err != nil {
			s.send(500, err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			s.send(250, "So long, Bob")
			break
		}

		if !processCmd(s, cmd) {
			s.send(500, "Unknown command")
		}
	}

	conn.Close()
}

func createDir(path string) error {

	stat, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		stat, err = os.Stat(config.spooldir)
	}

	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}
