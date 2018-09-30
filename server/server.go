package server

import (
	"log"
	"net"
)

type Server struct {
	config *Config
}

func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() {
	debugLog = s.config.Debug

	err := createDir(s.config.Maildir)
	if err != nil {
		log.Fatal(err)
	}

	ok := false
	if s.config.Smtp != "" {
		go run(s.config.Smtp, processSMTP, s.config)
		ok = true
	}
	if s.config.Pop != "" {
		go run(s.config.Pop, processPOP, s.config)
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
