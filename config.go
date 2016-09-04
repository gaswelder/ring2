package main

import (
	"fmt"
	"genera/cfg"
)

var config struct {
	hostname string
	relay bool
	spooldir string
	listen string

	lists []string
	users []*userRec
}

func readConfig(path string) error {
	/*
	 * Init to default values
	 */
	config.listen = "localhost:2525"
	config.hostname = "localhost"
	config.relay = false
	config.spooldir = "./spool"
	config.lists = make([]string, 0)
	config.users = make([]*userRec, 0)

	r := cfg.NewReader()

	r.DefineSection("server", func(vals [][2]string) error {
		for _, val := range(vals) {
			switch val[0] {
				case "listen":
					config.listen = val[1]
				case "spooldir":
					config.spooldir = val[1]
				case "hostname":
					config.hostname = val[1]
				default:
					return fmt.Errorf("Unknown param %s", val[0])
			}
		}
		return nil
	})

	r.DefineSection("lists", func(vals [][2]string) error {
		for _, val := range(vals) {
			if val[1] != "" {
				return fmt.Errorf("Unexpected argument: %s %s", val[0], val[1])
			}
			config.lists = append(config.lists, val[0])
		}
		return nil
	})

	r.DefineSection("users", func(vals [][2]string) error {
		for _, val := range(vals) {
			name := val[0]
			_, remote, err := parseUserSpec(val[1])
			if err != nil {
				return err
			}
			user := &userRec{name, remote}
			config.users = append(config.users, user)
		}
		return nil
	})
	return r.ParseFile(path)
}

func parseUserSpec(spec string) ([]string, string, error) {
	lists := make([]string, 0)
	remote := ""
	return lists, remote, nil
}
