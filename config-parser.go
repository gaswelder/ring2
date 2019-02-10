package main

import (
	"errors"
	"fmt"

	"github.com/gaswelder/ring2/cfg"
	"github.com/gaswelder/ring2/scanner"
	"github.com/gaswelder/ring2/server"
)

func readConfig(path string) (*server.Config, error) {
	cnf := server.Config{
		Hostname: "localhost",
		Maildir:  "./mail",
		Lists:    make(map[string][]*server.UserRec),
		Users:    make(map[string]*server.UserRec),
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
				cnf.Smtp = val
			case "pop":
				cnf.Pop = val
			case "maildir":
				cnf.Maildir = val
			case "hostname":
				cnf.Hostname = val
			case "debug":
				cnf.Debug = true
			default:
				return nil, fmt.Errorf("Unknown param %s", key)
			}
		}
	}

	sec, ok = conf["lists"]
	if ok {
		for key, val := range sec {
			if val != "true" {
				return nil, fmt.Errorf("Unexpected argument to the maillist '%s': %s", key, val)
			}
			cnf.Lists[key] = make([]*server.UserRec, 0)
		}
	}

	sec, ok = conf["users"]
	if ok {
		for key, val := range sec {
			user, err := parseUserSpec(val)
			if err != nil {
				return nil, err
			}
			user.Name = key
			cnf.Users[key] = user
			for _, listname := range user.Lists {
				_, ok := cnf.Lists[listname]
				if !ok {
					return nil, fmt.Errorf("Unknown list: %s", listname)
				}
				cnf.Lists[listname] = append(cnf.Lists[listname], user)
			}
		}
	}
	return &cnf, nil
}

func parseUserSpec(spec string) (*server.UserRec, error) {

	user := new(server.UserRec)
	b := scanner.New(spec)

	if b.Next() == '$' {
		for b.More() && !isSpace(b.Next()) {
			user.Pwhash += string(b.Get())
		}
	} else if b.Next() == '"' {
		b.Get()
		for b.More() && b.Next() != '"' {
			user.Password += string(b.Get())
		}
		if b.Get() != '"' {
			return nil, errors.New("Unmatched password quote")
		}
	}

	user.Lists = make([]string, 0)
	lists, err := parseLists(b)
	if err != nil {
		return nil, err
	}
	for _, name := range lists {
		user.Lists = append(user.Lists, name)
	}
	return user, nil
}

func parseLists(b *scanner.Scanner) ([]string, error) {
	lists := make([]string, 0)
	skipSpace(b)
	if b.Next() != '[' {
		return lists, nil
	}
	b.Get()

	skipSpace(b)
	for b.More() && b.Next() != ']' {
		name := b.ReadName()
		if name == "" {
			return lists, errors.New("Empty list name before " + b.Rest())
		}
		lists = append(lists, name)
		skipSpace(b)
		if b.Next() == ',' {
			b.Get()
			skipSpace(b)
		}
	}

	if b.Next() != ']' {
		return lists, errors.New("']' expected")
	}
	return lists, nil
}

func skipSpace(b *scanner.Scanner) {
	for b.More() && isSpace(b.Next()) {
		b.Get()
	}
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}
