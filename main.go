package main

import (
	"log"
	"net"
)

func main() {

	config, err := readConfig("conf")
	if err != nil {
		log.Fatal(err)
	}
	debugLog = config.debug

	err = createDir(config.maildir)
	if err != nil {
		log.Fatal(err)
	}

	ok := false
	if config.smtp != "" {
		go server(config.smtp, processSMTP, config)
		ok = true
	}
	if config.pop != "" {
		go server(config.pop, processPOP, config)
		ok = true
	}
	if !ok {
		log.Fatal("Both SMTP and POP disabled")
	}
	select {}
}

func server(addr string, f func(net.Conn, *serverConfig), config *serverConfig) error {
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
		go f(conn, config)
	}
}
