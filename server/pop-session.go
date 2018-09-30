package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

/*
 * User POP session
 */
type popState struct {
	userName string
	user     *UserRec
	box      *mailbox
	lastID   int
	conn     net.Conn
	r        *bufio.Reader
	config   *Config
}

func newPopSession(c net.Conn, config *Config) *popState {
	s := new(popState)
	s.conn = c
	s.r = bufio.NewReader(c)
	s.config = config
	return s
}

// Send a success response with optional comment
func (s *popState) ok(comment string, args ...interface{}) {
	if comment != "" {
		s.send("+OK " + fmt.Sprintf(comment, args...))
	} else {
		s.send("+OK")
	}
}

// Send an error response with optinal comment
func (s *popState) err(comment string) {
	if comment != "" {
		s.send("-ERR " + comment)
	} else {
		s.send("-ERR")
	}
}

// Send a line
func (s *popState) send(format string, args ...interface{}) error {
	line := fmt.Sprintf(format+"\r\n", args...)
	_, err := s.conn.Write([]byte(line))
	return err
}

// Send a multiline data
func (s *popState) sendData(data string) error {
	var err error
	lines := strings.Split(data, "\r\n")
	for _, line := range lines {
		err = s.sendDataLine(line)
		if err != nil {
			return err
		}
	}
	err = s.send(".")
	return err
}

// Sends a line of data, taking care of the "dot-stuffing"
func (s *popState) sendDataLine(line string) error {
	if len(line) > 0 && line[0] == '.' {
		line = "." + line
	}
	line += "\r\n"
	_, err := s.conn.Write([]byte(line))
	return err
}

func (s *popState) messages() []*message {
	list := make([]*message, 0)
	for _, msg := range s.box.messages {
		if msg.deleted {
			continue
		}
		list = append(list, msg)
	}
	return list
}

func (s *popState) undelete() {
	for _, msg := range s.box.messages {
		msg.deleted = false
	}
}

func (s *popState) findMessage(id int) *message {
	for _, msg := range s.messages() {
		if msg.id == id {
			return msg
		}
	}
	return nil
}

func (s *popState) getMessage(arg string) (*message, error) {
	if arg == "" {
		return nil, errors.New("Missing argument")
	}
	id, err := strconv.Atoi(arg)
	if err != nil {
		return nil, err
	}

	msg := s.findMessage(id)
	if msg != nil {
		return msg, nil
	}
	return nil, errors.New("No such message")
}

func (s *popState) markAsDeleted(msgid string) error {
	msg, err := s.getMessage(msgid)
	if err != nil {
		return err
	}
	msg.deleted = true
	return nil
}
