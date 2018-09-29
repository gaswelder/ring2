package main

import (
	"golang.org/x/crypto/bcrypt"
)

// Returns user record with given name and password.
// Returns nil if there is no such user.
func findUser(name, pass string) *userRec {
	for _, user := range config.users {
		if user.name != name {
			continue
		}

		if user.password != "" {
			if user.password == pass {
				return user
			}
			return nil
		}

		if user.pwhash != "" {
			err := bcrypt.CompareHashAndPassword([]byte(user.pwhash), []byte(pass))
			if err == nil {
				return user
			}
			return nil
		}
	}
	return nil
}
