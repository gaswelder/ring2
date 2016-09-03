package main

import (
	"net"
	"bufio"
)

/*
 * User record
 */
type userRec struct {
	name string
	remote string
}

/*
 * Client's command
 */
type command struct {
	name string
	arg string
}

/*
 * Forward or reverse path
 */
type path struct {
	// zero or more lists of hostnames like foo.com
	hosts []string
	// address endpoint, like bob@example.net
	address string
}

/*
 * A mail draft
 */
type mail struct {
	sender *path
	recipients []*path
}

func newDraft(from *path) *mail {
	return &mail{
		path,
		make([]*path, 0),
	}
}

/*
 * A user session, or context.
 */
type session struct {
	senderHost string
	conn net.Conn
	r bufio.Reader
	draft *mail
}
