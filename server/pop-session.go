package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

/*
 * User POP session
 */
type popState struct {
	userName            string
	user                *UserRec
	box                 *mailbox
	lastID              int
	conn                net.Conn
	r                   *bufio.Reader
	config              *Config
	deletedMessageNames []string
}

func newPopSession(c net.Conn, config *Config) *popState {
	s := new(popState)
	s.conn = c
	s.r = bufio.NewReader(c)
	s.config = config
	s.deletedMessageNames = make([]string, 0)
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

func (s *popState) deleted(msg *message) bool {
	for _, fn := range s.deletedMessageNames {
		if fn == msg.filename {
			return true
		}
	}
	return false
}

func (s *popState) messages() []*message {
	list := make([]*message, 0)
	for _, msg := range s.box.messages {
		if s.deleted(msg) {
			continue
		}
		list = append(list, msg)
	}
	return list
}

func (s *popState) undelete() {
	s.deletedMessageNames = make([]string, 0)
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
	s.deletedMessageNames = append(s.deletedMessageNames, msg.filename)
	return nil
}

func (s *popState) begin() error {
	box, err := newBox(s.user, s.config)
	if err != nil {
		return err
	}
	err = box.lock()
	if err != nil {
		return err
	}
	err = box.parseMessages()
	if err != nil {
		return err
	}
	s.box = box
	s.lastID = box.lastID
	return nil
}

func (s *popState) commit() error {
	if s.box == nil {
		return nil
	}
	s.box.setLast(s.lastID)
	err := s.purge()
	s.box.unlock()
	return err
}

func (s *popState) stat() (count int, size int64, err error) {
	for _, msg := range s.messages() {
		count++
		size += msg.size
	}
	return count, size, err
}

// Remove messages marked to be deleted
func (s *popState) purge() error {
	l := make([]*message, 0)

	for _, msg := range s.box.messages {
		if !s.deleted(msg) {
			l = append(l, msg)
			continue
		}
		err := os.Remove(s.box.path + "/" + msg.filename)
		if err != nil {
			return err
		}
	}

	s.box.messages = l
	return nil
}
