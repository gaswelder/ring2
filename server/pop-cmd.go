package server

import (
	"fmt"
	"strings"

	"github.com/gaswelder/ring2/server/pop"
)

type popfunc func(s *popState, c *pop.Command)

var popFuncs = make(map[string]popfunc)

func popCmd(name string, f popfunc) {
	popFuncs[name] = f
}

func execPopCmd(s *popState, c *pop.Command) bool {
	f, ok := popFuncs[c.Name]
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
	popCmd("USER", func(s *popState, cmd *pop.Command) {
		if s.inbox != nil {
			s.Err("Wrong commands order")
			return
		}
		if cmd.Arg == "" {
			s.Err("username expected")
			return
		}
		s.userName = cmd.Arg
		s.OK("")
	})

	/*
	 * PASS <key>
	 */
	popCmd("PASS", func(s *popState, c *pop.Command) {
		if s.inbox != nil || s.userName == "" {
			s.Err("Wrong commands order")
			return
		}
		user := s.server.findUser(s.userName, c.Arg)
		if user == nil {
			s.Err("Auth failed")
			s.userName = ""
			return
		}

		/*
		 * Lock and parse the user's box. If failed, reset back
		 * to authentication phase.
		 */
		err := s.begin(user)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("")
	})

	/*
	 * STAT
	 */
	popCmd("STAT", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		count, size, err := s.inbox.stat()
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("%d %d", count, size)
	})

	/*
	 * LIST [<id>]
	 */
	popCmd("LIST", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}

		/*
		 * If no argument, show all undeleted messages
		 */
		if c.Arg == "" {
			s.OK("List follows")
			for _, entry := range s.inbox.entries() {
				s.Send("%d %d", entry.id, entry.msg.Size())
			}
			s.Send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		msg := s.inbox.findEntryByID(c.Arg)
		if msg == nil {
			s.Err("no such message")
			return
		}

		s.OK("%d %d", msg.id, msg.msg.Size())
	})

	/*
	 * RETR <id>
	 */
	popCmd("RETR", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}

		entry := s.inbox.findEntryByID(c.Arg)
		if entry == nil {
			s.Err("no such message")
			return
		}

		data, err := entry.msg.Content()
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("%d octets", entry.msg.Size())
		s.SendData(data)
		s.inbox.markRetrieved(entry)
	})

	/*
	 * DELE <id>
	 */
	popCmd("DELE", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		err := s.inbox.markAsDeleted(c.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("message %d deleted", c.Arg)
	})

	/*
	 * NOOP
	 */
	popCmd("NOOP", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("")
	})

	/*
	 * LAST
	 */
	popCmd("LAST", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("%d", s.inbox.lastID)
	})

	/*
	 * RSET
	 */
	popCmd("RSET", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		s.inbox.reset()
		s.OK("")
	})

	/*
	 * Optional commands
	 */

	/*
	 * UIDL[ <msg>]
	 */
	popCmd("UIDL", func(s *popState, c *pop.Command) {
		if !checkAuth(s) {
			return
		}
		if c.Arg == "" {
			s.OK("")
			for _, entry := range s.inbox.entries() {
				s.Send("%d %s", entry.id, entry.msg.Filename())
			}
			s.Send(".")
			return
		}

		msg := s.inbox.findEntryByID(c.Arg)
		if msg == nil {
			s.Err("no such message")
			return
		}
		s.OK("%d %s", msg.id, msg.msg.Filename())
	})

	/*
	 * TOP <msg> <n>
	 */
	popCmd("TOP", func(s *popState, c *pop.Command) {

		var n int
		var id string
		_, err := fmt.Sscanf(c.Arg, "%s %d", &id, &n)
		if err != nil {
			s.Err(err.Error())
			return
		}

		entry := s.inbox.findEntryByID(id)
		if entry == nil {
			s.Err("No such message")
			return
		}

		text, err := entry.msg.Content()
		if err != nil {
			s.Err(err.Error())
			return
		}

		lines := strings.Split(text, "\r\n")
		size := len(lines)
		i := 0

		/*
		 * Send all headers
		 */
		s.OK("")
		for i < size {
			s.Send("%s", lines[i])
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
			s.Send("---%d", n)
			s.Send("%s", lines[i])
			i++
			n--
		}
		s.Send(".")
	})

	popCmd("RPOP", func(s *popState, c *pop.Command) {
		s.Err("How such a command got into the RFC at all?")
	})
}

func checkAuth(s *popState) bool {
	if s.inbox == nil {
		s.Err("Unauthorized")
		return false
	}
	return true
}
