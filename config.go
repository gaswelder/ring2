package main

import (
	"fmt"
	"github.com/gaswelder/cfg"
)

var config struct {
	hostname string
	relay    bool
	spooldir string
	smtp     string
	pop      string
	lists    []string
	users    []*userRec
}

func readConfig(path string) error {
	/*
	 * Init to default values
	 */
	config.hostname = "localhost"
	config.relay = false
	config.spooldir = "./spool"
	config.lists = make([]string, 0)
	config.users = make([]*userRec, 0)

	conf, err := cfg.ParseFile(path)
	if err != nil {
		return err
	}

	sec, ok := conf["server"]
	if ok {
		for key, val := range sec {
			switch key {
			case "smtp":
				config.smtp = val
			case "pop":
				config.pop = val
			case "spooldir":
				config.spooldir = val
			case "hostname":
				config.hostname = val
			default:
				return fmt.Errorf("Unknown param %s", key)
			}
		}
	}

	sec, ok = conf["lists"]
	if ok {
		for key, val := range sec {
			if val != "" {
				return fmt.Errorf("Unexpected argument: %s %s", key, val)
			}
			config.lists = append(config.lists, key)
		}
	}

	sec, ok = conf["users"]
	if ok {
		for key, val := range sec {
			name := key
			_, remote, err := parseUserSpec(val)
			if err != nil {
				return err
			}
			user := &userRec{name, remote}
			config.users = append(config.users, user)
		}
	}
	return nil
}

func parseUserSpec(spec string) ([]string, string, error) {
	lists := make([]string, 0)
	remote := ""
	return lists, remote, nil
}
