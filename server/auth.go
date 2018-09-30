package server

import (
	"golang.org/x/crypto/bcrypt"
)

// Returns user record with given name and password.
// Returns nil if there is no such user.
func findUser(name, pass string, config *Config) *UserRec {
	for _, user := range config.Users {
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
