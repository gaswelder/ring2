package server

import "github.com/gaswelder/ring2/server/mailbox"

func (u *UserRec) mailbox(config *Config) (*mailbox.Mailbox, error) {
	path := config.Maildir + "/" + u.Name
	return mailbox.New(path)
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
