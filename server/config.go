package server

import (
	"github.com/gaswelder/ring2/server/mailbox"
	"golang.org/x/crypto/bcrypt"
)

/*
 * User record
 */
type UserRec struct {
	Name     string
	Pwhash   string
	Password string
	Lists    []string
}

// Config is a structure to keep user-provided
// server parameters.
type Config struct {
	Debug    bool
	Hostname string
	Maildir  string
	Smtp     string
	Pop      string
	Lists    map[string][]*UserRec
	Users    map[string]*UserRec
}

// Returns user record with given name and password.
// Returns nil if there is no such user.
func (c *Config) findUser(name, pass string) *UserRec {
	for _, user := range c.Users {
		if user.Name != name {
			continue
		}

		if user.Password != "" {
			if user.Password == pass {
				return user
			}
			return nil
		}

		if user.Pwhash != "" {
			err := bcrypt.CompareHashAndPassword([]byte(user.Pwhash), []byte(pass))
			if err == nil {
				return user
			}
			return nil
		}
	}
	return nil
}

func (c *Config) mailbox(u *UserRec) (*mailbox.Mailbox, error) {
	path := c.Maildir + "/" + u.Name
	return mailbox.New(path)
}
