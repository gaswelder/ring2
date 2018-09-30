package pop

import (
	"fmt"
	"strings"
)

type popfunc func(s Session, c *Command)

var popFuncs = make(map[string]popfunc)

func popCmd(name string, f popfunc) {
	popFuncs[name] = f
}

func init() {
	/*
	 * USER <name>
	 */
	popCmd("USER", func(s Session, cmd *Command) {
		err := s.SetUserName(cmd.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("")
	})

	/*
	 * PASS <key>
	 */
	popCmd("PASS", func(s Session, c *Command) {
		err := s.Open(c.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("")
	})

	/*
	 * STAT
	 */
	popCmd("STAT", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		count, size, err := s.Inbox().Stat()
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("%d %d", count, size)
	})

	/*
	 * LIST [<id>]
	 */
	popCmd("LIST", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}

		/*
		 * If no argument, show all undeleted messages
		 */
		if c.Arg == "" {
			s.OK("List follows")
			for _, entry := range s.Inbox().Entries() {
				s.Send("%d %d", entry.Id, entry.Msg.Size())
			}
			s.Send(".")
			return
		}

		/*
		 * Otherwise treat as LIST <id>
		 */
		entry := s.Inbox().FindEntryByID(c.Arg)
		if entry == nil {
			s.Err("no such message")
			return
		}

		s.OK("%d %d", entry.Id, entry.Msg.Size())
	})

	/*
	 * RETR <id>
	 */
	popCmd("RETR", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}

		entry := s.Inbox().FindEntryByID(c.Arg)
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
		s.Inbox().MarkRetrieved(entry)
	})

	/*
	 * DELE <id>
	 */
	popCmd("DELE", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		err := s.Inbox().MarkAsDeleted(c.Arg)
		if err != nil {
			s.Err(err.Error())
			return
		}
		s.OK("message %d deleted", c.Arg)
	})

	/*
	 * NOOP
	 */
	popCmd("NOOP", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("")
	})

	/*
	 * LAST
	 */
	popCmd("LAST", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.OK("%d", s.Inbox().LastID())
	})

	/*
	 * RSET
	 */
	popCmd("RSET", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		s.Inbox().Reset()
		s.OK("")
	})

	/*
	 * Optional commands
	 */

	/*
	 * UIDL[ <msg>]
	 */
	popCmd("UIDL", func(s Session, c *Command) {
		if !checkAuth(s) {
			return
		}
		if c.Arg == "" {
			s.OK("")
			for _, entry := range s.Inbox().Entries() {
				s.Send("%d %s", entry.Id, entry.Msg.Filename())
			}
			s.Send(".")
			return
		}

		msg := s.Inbox().FindEntryByID(c.Arg)
		if msg == nil {
			s.Err("no such message")
			return
		}
		s.OK("%d %s", msg.Id, msg.Msg.Filename())
	})

	/*
	 * TOP <msg> <n>
	 */
	popCmd("TOP", func(s Session, c *Command) {

		var n int
		var id string
		_, err := fmt.Sscanf(c.Arg, "%s %d", &id, &n)
		if err != nil {
			s.Err(err.Error())
			return
		}

		entry := s.Inbox().FindEntryByID(id)
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

	popCmd("RPOP", func(s Session, c *Command) {
		s.Err("How such a command got into the RFC at all?")
	})
}

func checkAuth(s Session) bool {
	if s.Inbox() == nil {
		s.Err("Unauthorized")
		return false
	}
	return true
}
