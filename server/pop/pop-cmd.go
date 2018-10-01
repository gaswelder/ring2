package pop

import (
	"fmt"
	"strings"
)

/*
 * USER <name>
 */
func cmdUser(s *session, cmd *command) {
	if s.inbox != nil {
		s.Err("already authorized")
		return
	}
	name := cmd.arg
	if name == "" {
		s.Err("empty username")
		return
	}
	s.userName = name
	s.OK("")
}

/*
 * PASS <key>
 */
func cmdPass(s *session, c *command) {
	if s.inbox != nil {
		s.Err("Session already started")
		return
	}
	if s.userName == "" {
		s.Err("Wrong commands order")
		return
	}

	box, err := s.auth(s.userName, c.arg)
	if err != nil {
		s.Err(err.Error())
		return
	}

	m, err := makeInboxView(box)
	if err != nil {
		s.Err(err.Error())
		return
	}

	s.inbox = m
	s.OK("")
}

/*
 * STAT
 */
func cmdStat(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	count, size, err := s.inbox.stat()
	if err != nil {
		s.Err(err.Error())
		return
	}
	s.OK("%d %d", count, size)
}

/*
 * LIST [<id>]
 */
func cmdList(s *session, c *command) {
	if !checkAuth(s) {
		return
	}

	/*
	 * If no argument, show all undeleted messages
	 */
	if c.arg == "" {
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
	entry := s.inbox.findEntry(c.arg)
	if entry == nil {
		s.Err("no such message")
		return
	}

	s.OK("%d %d", entry.id, entry.msg.Size())
}

/*
 * RETR <id>
 */
func cmdRetr(s *session, c *command) {
	if !checkAuth(s) {
		return
	}

	entry := s.inbox.findEntry(c.arg)
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
}

/*
 * DELE <id>
 */
func cmdDele(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	err := s.inbox.markDeleted(c.arg)
	if err != nil {
		s.Err(err.Error())
		return
	}
	s.OK("message %d deleted", c.arg)
}

/*
 * NOOP
 */
func cmdNoop(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	s.OK("")
}

/*
 * LAST
 */
func cmdLast(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	s.OK("%d", s.inbox.lastID)
}

/*
 * RSET
 */
func cmdRset(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	s.inbox.reset()
	s.OK("")
}

/*
 * UIDL[ <msg>]
 */
func cmdUidl(s *session, c *command) {
	if !checkAuth(s) {
		return
	}
	if c.arg == "" {
		s.OK("")
		for _, entry := range s.inbox.entries() {
			s.Send("%d %s", entry.id, entry.msg.Filename())
		}
		s.Send(".")
		return
	}

	msg := s.inbox.findEntry(c.arg)
	if msg == nil {
		s.Err("no such message")
		return
	}
	s.OK("%d %s", msg.id, msg.msg.Filename())
}

/*
 * TOP <msg> <n>
 */
func cmdTop(s *session, c *command) {

	var n int
	var id string
	_, err := fmt.Sscanf(c.arg, "%s %d", &id, &n)
	if err != nil {
		s.Err(err.Error())
		return
	}

	entry := s.inbox.findEntry(id)
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
}

func checkAuth(s *session) bool {
	if s.inbox == nil {
		s.Err("Unauthorized")
		return false
	}
	return true
}
