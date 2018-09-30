package server

import (
	"fmt"
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
		user := findUser(s.userName, c.arg, s.config)
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
		box, err := openBox(user, s.config)
		if err != nil {
			s.err(err.Error())
			s.userName = ""
			s.user = nil
			return
		}
		s.box = box
		s.lastID = box.lastID
		s.ok("")
	})

	/*
	 * STAT
	 */
	popCmd("STAT", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		count, size, err := s.stat()
		if err != nil {
			s.err(err.Error())
			return
		}
		s.ok("%d %d", count, size)
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
			for _, msg := range s.messages() {
				s.send("%d %d", msg.id, msg.size)
			}
			s.send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		msg, err := s.getMessage(c.arg)
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

		msg, err := s.getMessage(c.arg)
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
		if msg.id > s.lastID {
			s.lastID = msg.id
		}
	})

	/*
	 * DELE <id>
	 */
	popCmd("DELE", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		err := s.markAsDeleted(c.arg)
		if err != nil {
			s.err(err.Error())
			return
		}
		s.ok("message %d deleted", c.arg)
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
		s.ok("%d", s.lastID)
	})

	/*
	 * RSET
	 */
	popCmd("RSET", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		// Unmark deleted messages and reset last ID
		s.undelete()
		s.lastID = s.box.lastID
		s.ok("")
	})

	/*
	 * Optional commands
	 */

	/*
	 * UIDL[ <msg>]
	 */
	popCmd("UIDL", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}
		if c.arg == "" {
			s.ok("")
			for _, msg := range s.messages() {
				s.send("%d %s", msg.id, msg.filename)
			}
			s.send(".")
			return
		}

		msg, err := s.getMessage(c.arg)
		if err != nil {
			s.err(err.Error())
			return
		}
		s.ok("%d %s", msg.id, msg.filename)
	})

	/*
	 * TOP <msg> <n>
	 */
	popCmd("TOP", func(s *popState, c *command) {

		var id, n int
		_, err := fmt.Sscanf(c.arg, "%d %d", &id, &n)
		if err != nil {
			s.err(err.Error())
			return
		}

		msg := s.findMessage(id)
		if msg == nil {
			s.err("No such message")
			return
		}

		text, err := msg.Content()
		if err != nil {
			s.err(err.Error())
			return
		}

		lines := strings.Split(text, "\r\n")
		size := len(lines)
		i := 0

		/*
		 * Send all headers
		 */
		s.ok("")
		for i < size {
			s.send("%s", lines[i])
			if lines[i] == "" {
				break
			}
			i++
		}

		/*
		 * Send no more than 'n' lines of the body
		 */
		i++
		for i < size && n > 0 {
			s.send("---%d", n)
			s.send("%s", lines[i])
			i++
			n--
		}
		s.send(".")
	})

	popCmd("RPOP", func(s *popState, c *command) {
		s.err("How such a command got into the RFC at all?")
	})
}

// Load a user's box, lock it and parse the messages list
func openBox(user *UserRec, config *Config) (box *mailbox, err error) {
	box, err = newBox(user, config)
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
