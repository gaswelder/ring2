package main

import (
	"errors"
	"fmt"

	"github.com/gaswelder/cfg"
)

type serverConfig struct {
	debug    bool
	hostname string
	maildir  string
	smtp     string
	pop      string
	lists    map[string][]*userRec
	users    map[string]*userRec
}

func readConfig(path string) (*serverConfig, error) {
	cnf := serverConfig{
		hostname: "localhost",
		maildir:  "./mail",
		lists:    make(map[string][]*userRec),
		users:    make(map[string]*userRec),
	}

	conf, err := cfg.ParseFile(path)
	if err != nil {
		return nil, err
	}

	sec, ok := conf["server"]
	if ok {
		for key, val := range sec {
			switch key {
			case "smtp":
				cnf.smtp = val
			case "pop":
				cnf.pop = val
			case "maildir":
				cnf.maildir = val
			case "hostname":
				cnf.hostname = val
			case "debug":
				cnf.debug = true
			default:
				return nil, fmt.Errorf("Unknown param %s", key)
			}
		}
	}

	sec, ok = conf["lists"]
	if ok {
		for key, val := range sec {
			if val != "" {
				return nil, fmt.Errorf("Unexpected argument: %s %s", key, val)
			}
			cnf.lists[key] = make([]*userRec, 0)
		}
	}

	sec, ok = conf["users"]
	if ok {
		for key, val := range sec {
			user, err := parseUserSpec(val)
			if err != nil {
				return nil, err
			}
			user.name = key
			cnf.users[key] = user
			for _, listname := range user.lists {
				_, ok := cnf.lists[listname]
				if !ok {
					return nil, fmt.Errorf("Unknown list: %s", listname)
				}
				cnf.lists[listname] = append(cnf.lists[listname], user)
			}
		}
	}
	return &cnf, nil
}

func parseUserSpec(spec string) (*userRec, error) {

	user := new(userRec)
	b := newScanner(spec)

	if b.next() == '$' {
		for b.more() && !isSpace(b.next()) {
			user.pwhash += string(b.get())
		}
	} else if b.next() == '"' {
		b.get()
		for b.more() && b.next() != '"' {
			user.password += string(b.get())
		}
		if b.get() != '"' {
			return nil, errors.New("Unmatched password quote")
		}
	}

	user.lists = make([]string, 0)
	lists, err := parseLists(b)
	if err != nil {
		return nil, err
	}
	for _, name := range lists {
		user.lists = append(user.lists, name)
	}
	return user, nil
}

func parseLists(b *scanner) ([]string, error) {
	lists := make([]string, 0)
	skipSpace(b)
	if b.next() != '[' {
		return lists, nil
	}
	b.get()

	skipSpace(b)
	for b.more() && b.next() != ']' {
		name := readName(b)
		if name == "" {
			return lists, errors.New("Empty list name before " + b.rest())
		}
		lists = append(lists, name)
		skipSpace(b)
		if b.next() == ',' {
			b.get()
			skipSpace(b)
		}
	}

	if b.next() != ']' {
		return lists, errors.New("']' expected")
	}
	return lists, nil
}

func skipSpace(b *scanner) {
	for b.more() && isSpace(b.next()) {
		b.get()
	}
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}
