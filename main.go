package main

import (
	"log"
	"net"
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

	go server(config.listen, processSMTP)
	go server(":11000", processPOP)
	select{}
}

func server(addr string, f func(net.Conn)) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Listening on %s\n", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go f(conn)
	}
}
