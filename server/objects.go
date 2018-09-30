package server

import "github.com/gaswelder/ring2/server/smtp"

/*
 * A mail draft
 */
type mail struct {
	sender     *smtp.Path
	recipients []*smtp.Path
}

func newDraft(from *smtp.Path) *mail {
	return &mail{
		from,
		make([]*smtp.Path, 0),
	}
}
