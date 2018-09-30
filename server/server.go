package server

import (
	"log"
	"net"
	"os"
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

	go s.smtp()
	go s.pop()
}

func createDir(path string) error {
	stat, err := os.Stat(path)
	if stat != nil && err == nil {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func (s *Server) pop() error {
	ln, err := net.Listen("tcp", s.config.Pop)
	if err != nil {
		return err
	}
	log.Printf("POP: listening on %s\n", s.config.Pop)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		go func() {
			processPOP(conn, s.config)
			conn.Close()
		}()
	}
}

func (s *Server) smtp() error {
	ln, err := net.Listen("tcp", s.config.Smtp)
	if err != nil {
		return err
	}
	log.Printf("SMTP: listening on %s\n", s.config.Smtp)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		go func() {
			processSMTP(conn, s.config)
			conn.Close()
		}()
	}
}
