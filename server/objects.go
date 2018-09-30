package server

import "github.com/gaswelder/ring2/server/mailbox"

/*
 * User record
 */
type UserRec struct {
	Name     string
	Pwhash   string
	Password string
	Lists    []string
}

func (u *UserRec) mailbox(config *Config) (*mailbox.Mailbox, error) {
	path := config.Maildir + "/" + u.Name
	return mailbox.New(path)
}

/*
 * Client's command
 */
type command struct {
	name string
	arg  string
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
	sender     *path
	recipients []*path
}

func newDraft(from *path) *mail {
	return &mail{
		from,
		make([]*path, 0),
	}
}
