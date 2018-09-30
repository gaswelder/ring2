package server

import (
	"log"
	"net"
	"os"

	"github.com/gaswelder/ring2/server/pop"
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

	go runSMTP(s.config)
	go runPOP(s.config)
}

func createDir(path string) error {
	stat, err := os.Stat(path)
	if stat != nil && err == nil {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func runPOP(config *Config) error {
	ln, err := net.Listen("tcp", config.Pop)
	if err != nil {
		return err
	}
	log.Printf("POP: listening on %s\n", config.Pop)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		go func() {
			s := newPopSession(conn, config)
			pop.Process(s)
			conn.Close()
			log.Printf("%s disconnected\n", conn.RemoteAddr().String())
		}()
	}
}

func runSMTP(config *Config) error {
	ln, err := net.Listen("tcp", config.Smtp)
	if err != nil {
		return err
	}
	log.Printf("SMTP: listening on %s\n", config.Smtp)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		go func() {
			processSMTP(conn, config)
			conn.Close()
			log.Printf("%s disconnected\n", conn.RemoteAddr().String())
		}()
	}
}
