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
		if s.box != nil {
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
		if s.box != nil || s.userName == "" {
			s.err("Wrong commands order")
			return
		}
		user := findUser(s.userName, c.arg, s.config)
		if user == nil {
			s.err("Auth failed")
			s.userName = ""
			return
		}

		/*
		 * Lock and parse the user's box. If failed, reset back
		 * to authentication phase.
		 */
		err := s.begin(user)
		if err != nil {
			s.err(err.Error())
			return
		}
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
			for _, entry := range s.entries() {
				s.send("%d %d", entry.id, entry.msg.size)
			}
			s.send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		msg := s.findEntryByID(c.arg)
		if msg == nil {
			s.err("no such message")
			return
		}

		s.ok("%d %d", msg.id, msg.msg.size)
	})

	/*
	 * RETR <id>
	 */
	popCmd("RETR", func(s *popState, c *command) {
		if !checkAuth(s) {
			return
		}

		entry := s.findEntryByID(c.arg)
		if entry == nil {
			s.err("no such message")
			return
		}

		data, err := entry.msg.Content()
		if err != nil {
			s.err(err.Error())
			return
		}
		s.ok("%d octets", entry.msg.size)
		s.sendData(data)
		s.markRetrieved(entry)
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
		s.reset()
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
			for _, entry := range s.entries() {
				s.send("%d %s", entry.id, entry.msg.filename)
			}
			s.send(".")
			return
		}

		msg := s.findEntryByID(c.arg)
		if msg == nil {
			s.err("no such message")
			return
		}
		s.ok("%d %s", msg.id, msg.msg.filename)
	})

	/*
	 * TOP <msg> <n>
	 */
	popCmd("TOP", func(s *popState, c *command) {

		var n int
		var id string
		_, err := fmt.Sscanf(c.arg, "%s %d", &id, &n)
		if err != nil {
			s.err(err.Error())
			return
		}

		entry := s.findEntryByID(id)
		if entry == nil {
			s.err("No such message")
			return
		}

		text, err := entry.msg.Content()
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

func checkAuth(s *popState) bool {
	if s.box == nil {
		s.err("Unauthorized")
		return false
	}
	return true
}
