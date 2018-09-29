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

	err = createDir(config.maildir)
	if err != nil {
		log.Fatal(err)
	}

	ok := false
	if config.smtp != "" {
		go server(config.smtp, processSMTP)
		ok = true
	}
	if config.pop != "" {
		go server(config.pop, processPOP)
		ok = true
	}
	if !ok {
		log.Fatal("Both SMTP and POP disabled")
	}
	select {}
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
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		go f(conn)
	}
}
