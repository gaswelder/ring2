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

// func (s *popState) Close() error {
// 	return s.inbox.Commit()
// }

// user := s.config.findUser(s.userName, password)
// 		if user == nil {
// 			s.Err("auth failed")
// 			return
// 		}
// 		box, err := s.config.mailbox(user)
// 		if err != nil {
// 			return err
// 		}
