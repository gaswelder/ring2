package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gaswelder/ring2/server/mailbox"
)

// popMessageEntry represents a message entry used
// in a POP session. It contains a cached message entry
// from the mailbox plus session-related data like ID
// and the `deleted` flag.
type popMessageEntry struct {
	id      int
	msg     *mailbox.Message
	deleted bool
}

/*
 * User POP session
 */
type popState struct {
	userName    string
	box         *mailbox.Mailbox
	lastID      int
	conn        net.Conn
	r           *bufio.Reader
	config      *Config
	messageList []*popMessageEntry
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
	debMsg("pop send: %s", line)
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

func (s *popState) entries() []*popMessageEntry {
	list := make([]*popMessageEntry, 0)
	for _, msg := range s.messageList {
		if msg.deleted {
			continue
		}
		list = append(list, msg)
	}
	return list
}

func (s *popState) actualLastRetrievedEntry() (*popMessageEntry, error) {
	last, err := s.box.LastRetrievedMessage()
	if err != nil {
		return nil, err
	}
	if last == nil {
		return nil, nil
	}
	for _, entry := range s.messageList {
		if entry.msg.Filename() == last.Filename() {
			return entry, nil
		}
	}
	return nil, nil
}

func (s *popState) reset() error {
	// Reset all deleted flags
	for _, entry := range s.messageList {
		entry.deleted = false
	}
	last, err := s.actualLastRetrievedEntry()
	if err != nil {
		return err
	}
	s.lastID = last.id
	return nil
}

func (s *popState) markRetrieved(entry *popMessageEntry) {
	if entry.id > s.lastID {
		s.lastID = entry.id
	}
}

func (s *popState) getMessage(arg string) (*mailbox.Message, error) {
	if arg == "" {
		return nil, errors.New("Missing argument")
	}
	id, err := strconv.Atoi(arg)
	if err != nil {
		return nil, err
	}

	for _, entry := range s.messageList {
		if entry.id == id && !entry.deleted {
			return entry.msg, nil
		}
	}
	return nil, errors.New("No such message")
}

func (s *popState) findEntryByID(msgid string) *popMessageEntry {
	if msgid == "" {
		return nil
	}
	id, err := strconv.Atoi(msgid)
	if err != nil {
		return nil
	}
	for _, entry := range s.messageList {
		if entry.id == id {
			return entry
		}
	}
	return nil
}

func (s *popState) markAsDeleted(msgid string) error {
	e := s.findEntryByID(msgid)
	if e == nil {
		return errors.New("no such message")
	}
	e.deleted = true
	return nil
}

func (s *popState) begin(user *UserRec) error {
	box, err := user.mailbox(s.config)
	if err != nil {
		return err
	}

	id := 0
	ls, err := box.List()
	if err != nil {
		return err
	}
	for _, msg := range ls {
		id++
		log.Println(id, msg)
		s.messageList = append(s.messageList, &popMessageEntry{
			id:      id,
			msg:     msg,
			deleted: false,
		})
	}
	s.box = box
	last, err := s.actualLastRetrievedEntry()
	if err != nil {
		return err
	}
	if last != nil {
		s.lastID = last.id
	}
	return nil
}

func (s *popState) commit() error {
	if s.box == nil {
		return nil
	}
	last := s.lastRetrievedEntry()
	if last != nil {
		s.box.SetLast(last.msg)
	}
	err := s.purge()
	// s.box.unlock()
	return err
}

func (s *popState) lastRetrievedEntry() *popMessageEntry {
	for _, entry := range s.entries() {
		if entry.id == s.lastID {
			return entry
		}
	}
	return nil
}

func (s *popState) stat() (count int, size int64, err error) {
	for _, msg := range s.entries() {
		count++
		size += msg.msg.Size()
	}
	return count, size, err
}

// Remove messages marked to be deleted
func (s *popState) purge() error {
	l := make([]*popMessageEntry, 0)

	for _, entry := range s.messageList {
		if !entry.deleted {
			l = append(l, entry)
			continue
		}
		err := s.box.Remove(entry.msg)
		if err != nil {
			return err
		}
	}
	s.messageList = l
	return nil
}
