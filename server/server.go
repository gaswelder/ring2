package server

import (
	"log"
	"net"
)

func Run(config *Config) {
	debugLog = config.Debug

	err := createDir(config.Maildir)
	if err != nil {
		log.Fatal(err)
	}

	ok := false
	if config.Smtp != "" {
		go run(config.Smtp, processSMTP, config)
		ok = true
	}
	if config.Pop != "" {
		go run(config.Pop, processPOP, config)
		ok = true
	}
	if !ok {
		log.Fatal("Both SMTP and POP disabled")
	}
}

func run(addr string, f func(net.Conn, *Config), config *Config) error {
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
