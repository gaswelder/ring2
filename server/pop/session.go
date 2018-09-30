package pop

import (
	"io"
)

type session struct {
	userName string
	inbox    *inboxView
	*readWriter
	auth AuthFunc
}

func makeSession(c io.ReadWriter, auth AuthFunc) *session {
	return &session{
		readWriter: makeReadWriter(c),
		auth:       auth,
	}
}
