package server

import (
	"errors"
	"io"
	"log"
	"net"
	"os"

	"github.com/gaswelder/ring2/server/mailbox"
	"github.com/gaswelder/ring2/server/pop"
	"github.com/gaswelder/ring2/server/smtp"
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

func auth(config *Config) pop.AuthFunc {
	return func(name, password string) (*mailbox.Mailbox, error) {
		user := config.findUser(name, password)
		if user == nil {
			return nil, errors.New("invalid credentials")
		}
		return config.mailbox(user)
	}
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
		var rw io.ReadWriter = conn
		if config.Debug {
			rw = &tap{rw}
		}
		go func() {
			pop.Process(rw, auth(config))
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

	auth := func(name, password string) error {
		u := config.findUser(name, password)
		if u != nil {
			return nil
		}
		return errors.New("Invalid authorization data")
	}

	getbox := func(name string) ([]*mailbox.Mailbox, error) {
		boxes := make([]*mailbox.Mailbox, 0)

		list, _ := config.Lists[name]
		if list != nil {
			for _, user := range list {
				box, err := config.mailbox(user)
				if err != nil {
					return nil, err
				}
				boxes = append(boxes, box)
			}
			return boxes, nil
		}

		user, ok := config.Users[name]
		if ok {
			box, err := config.mailbox(user)
			if err != nil {
				return nil, err
			}
			boxes = append(boxes, box)
			return boxes, nil

		}
		return nil, errors.New("unknown recipient")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s connected\n", conn.RemoteAddr().String())
		var rw io.ReadWriter = conn
		if config.Debug {
			rw = &tap{rw}
		}
		go func() {
			smtp.Process(rw, auth, getbox)
			conn.Close()
			log.Printf("%s disconnected\n", conn.RemoteAddr().String())
		}()
	}
}

type tap struct {
	rw io.ReadWriter
}

func (t *tap) Read(p []byte) (n int, err error) {
	n, err = t.rw.Read(p)
	if err == nil {
		os.Stderr.WriteString("> " + string(p[:n]))
	}
	return n, err
}

func (t *tap) Write(p []byte) (n int, err error) {
	n, err = t.rw.Write(p)
	os.Stderr.WriteString("< " + string(p[:n]))
	return n, err
}
