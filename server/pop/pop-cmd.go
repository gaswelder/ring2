package pop

import (
	"fmt"
	"strings"
)

type popfunc func(s *session, c *Command)

var popFuncs = make(map[string]popfunc)

func popCmd(name string, f popfunc) {
	popFuncs[name] = f
}

func init() {
	/*
	 * USER <name>
	 */
	popCmd("USER", func(s *session, cmd *Command) {
		if s.inbox != nil {
			s.Err("already authorized")
			return
		}
		name := cmd.Arg
		if name == "" {
			s.Err("empty username")
			return
		}
		s.userName = name
		s.OK("")
	})

	/*
	 * PASS <key>
	 */
	popCmd("PASS", func(s *session, c *Command) {
		if s.inbox != nil {
			s.Err("Session already started")
			return
		}
		if s.userName == "" {
			s.Err("Wrong commands order")
			return
		}

		box, err := s.auth(s.userName, c.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}

		m, err := NewInboxView(box)
		if err != nil {
			s.Err(err.Error())
			return
		}

		s.inbox = m
		s.OK("")
	})

	/*
	 * STAT
	 */
	popCmd("STAT", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		count, size, err := s.inbox.Stat()
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("%d %d", count, size)
	})

	/*
	 * LIST [<id>]
	 */
	popCmd("LIST", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}

		/*
		 * If no argument, show all undeleted messages
		 */
		if c.Arg == "" {
			s.OK("List follows")
			for _, entry := range s.inbox.Entries() {
				s.Send("%d %d", entry.Id, entry.Msg.Size())
			}
			s.Send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		entry := s.inbox.FindEntryByID(c.Arg)
		if entry == nil {
			s.Err("no such message")
			return
		}

		s.OK("%d %d", entry.Id, entry.Msg.Size())
	})

	/*
	 * RETR <id>
	 */
	popCmd("RETR", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}

		entry := s.inbox.FindEntryByID(c.Arg)
		if entry == nil {
			s.Err("no such message")
			return
		}

		data, err := entry.Msg.Content()
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("%d octets", entry.Msg.Size())
		s.SendData(data)
		s.inbox.MarkRetrieved(entry)
	})

	/*
	 * DELE <id>
	 */
	popCmd("DELE", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		err := s.inbox.MarkAsDeleted(c.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("message %d deleted", c.Arg)
	})

	/*
	 * NOOP
	 */
	popCmd("NOOP", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("")
	})

	/*
	 * LAST
	 */
	popCmd("LAST", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("%d", s.inbox.LastID())
	})

	/*
	 * RSET
	 */
	popCmd("RSET", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.inbox.Reset()
		s.OK("")
	})

	/*
	 * Optional commands
	 */

	/*
	 * UIDL[ <msg>]
	 */
	popCmd("UIDL", func(s *session, c *Command) {
		if !checkAuth(s) {
			return
		}
		if c.Arg == "" {
			s.OK("")
			for _, entry := range s.inbox.Entries() {
				s.Send("%d %s", entry.Id, entry.Msg.Filename())
			}
			s.Send(".")
			return
		}

		msg := s.inbox.FindEntryByID(c.Arg)
		if msg == nil {
			s.Err("no such message")
			return
		}
		s.OK("%d %s", msg.Id, msg.Msg.Filename())
	})

	/*
	 * TOP <msg> <n>
	 */
	popCmd("TOP", func(s *session, c *Command) {

		var n int
		var id string
		_, err := fmt.Sscanf(c.Arg, "%s %d", &id, &n)
		if err != nil {
			s.Err(err.Error())
			return
		}

		entry := s.inbox.FindEntryByID(id)
		if entry == nil {
			s.Err("No such message")
			return
		}

		text, err := entry.Msg.Content()
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

	popCmd("RPOP", func(s *session, c *Command) {
		s.Err("How such a command got into the RFC at all?")
	})
}

func checkAuth(s *session) bool {
	if s.inbox == nil {
		s.Err("Unauthorized")
		return false
	}
	return true
}
