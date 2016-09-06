package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type popfunc func(s *popState, c *command)

var popFuncs = make(map[string]popfunc)

func popCmd(name string, f popfunc) {
	popFuncs[name] = f
}

func execPopCmd(s *popState, c *command) bool {
	f, ok := popFuncs[c.name]
	if !ok {
		return false
	}
	f(s, c)
	return true
}

func init() {
	/*
	 * USER <name>
	 */
	popCmd("USER", func(s *popState, cmd *command) {
		if s.user != nil {
			s.err("Wrong commands order")
			return
		}
		if cmd.arg == "" {
			s.err("username expected")
			return
		}
		s.userName = cmd.arg
		s.ok("")
	})

	/*
	 * PASS <key>
	 */
	popCmd("PASS", func(s *popState, c *command) {
		if s.user != nil || s.userName == "" {
			s.err("Wrong commands order")
			return
		}
		user := checkUser(s.userName, c.arg)
		if user == nil {
			s.err("Auth failed")
			s.userName = ""
			return
		}
		s.user = user

		/*
		 * Lock and parse the user's box. If failed, reset back
		 * to authentication phase.
		 */
		box, err := openBox(user)
		if err != nil {
			s.err(err.Error())
			s.userName = ""
			s.user = nil
			return
		}
		s.box = box
		s.lastId = box.lastId
		s.ok("")
	})

	/*
	 * STAT
	 */
	popCmd("STAT", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		s.ok("%d %d", s.box.count(), s.box.size())
	})

	/*
	 * LIST [<id>]
	 */
	popCmd("LIST", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}

		/*
		 * If no argument, show all undeleted messages
		 */
		if c.arg == "" {
			s.ok("List follows")
			for _, msg := range s.box.messages {
				if msg.deleted {
					continue
				}
				s.send("%d %d", msg.id, msg.size)
			}
			s.send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		msg, err := getMessage(s, c)
		if err != nil {
			s.err(err.Error())
			return
		}

		s.ok("%d %d", msg.id, msg.size)
	})

	/*
	 * RETR <id>
	 */
	popCmd("RETR", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}

		msg, err := getMessage(s, c)
		if err != nil {
			s.err(err.Error())
			return
		}

		data, err := msg.Content()
		if err != nil {
			s.err(err.Error())
			return
		}
		s.ok("%d octets", msg.size)
		s.sendData(data)

		/*
		 * Update the highest id
		 */
		if msg.id > s.lastId {
			s.lastId = msg.id
		}
	})

	/*
	 * DELE <id>
	 */
	popCmd("DELE", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}

		msg, err := getMessage(s, c)
		if err != nil {
			s.err(err.Error())
			return
		}

		// Mark the message to delete
		msg.deleted = true
		s.ok("message %d deleted", msg.id)
	})

	/*
	 * NOOP
	 */
	popCmd("NOOP", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		s.ok("")
	})

	/*
	 * LAST
	 */
	popCmd("LAST", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		s.ok("%d", s.lastId)
	})

	/*
	 * RSET
	 */
	popCmd("RSET", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		/*
		 * Unmark deleted messages
		 */
		for _, msg := range s.box.messages {
			msg.deleted = false
		}
		/*
		 * Reset last id
		 */
		s.lastId = s.box.lastId
		s.ok("")
	})

	/*
	 * Optional commands
	 */
	popCmd("TOP", func(s *popState, c *command) {
		s.err("Not implemented")
		/*
			TOP msg n
			+OK
			send all headers
			send n lines of the text
		*/
	})
	popCmd("RPOP", func(s *popState, c *command) {
		s.err("How such a command got into the RFC at all?")
	})
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
		if len(line) > 0 && line[0] == '.' {
			line = "." + line
		}
		err = s.send("%s", line)
		if err != nil {
			return err
		}
	}
	err = s.send(".")
	return err
}

// Returns user record with given name and password.
// Returns nil if there is no such user.
func checkUser(name, pass string) *userRec {
	fmt.Printf("Implement checkUser!\n")
	for _, user := range config.users {
		if user.name == name {
			return user
		}
	}
	return nil
}

// Load a user's box, lock it and parse the messages list
func openBox(user *userRec) (box *mailbox, err error) {
	box, err = newBox(user)
	if err != nil {
		return
	}
	err = box.lock()
	if err != nil {
		return
	}
	err = box.parseMessages()
	return
}

func checkAuth(s *popState) bool {
	if s.user == nil {
		s.err("Unauthorized")
		return false
	}
	return true
}

func getMessage(s *popState, c *command) (*message, error) {
	if c.arg == "" {
		return nil, errors.New("Missing argument")
	}

	id, err := strconv.Atoi(c.arg)
	if err != nil {
		return nil, err
	}
	for _, msg := range s.box.messages {
		if msg.id == id {
			return msg, nil
		}
	}
	return nil, errors.New("No such message")
}
