package cfg

import (
	"testing"
)

func TestParsing(t *testing.T) {
	example := `	# comment
	server {
		listen :2525
		relay
		hostname home.loc # comment
	}

	users {
		bob
		alice
		jack
	}
`
	cfg, err := parseString(example)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["server"]["listen"] != ":2525" {
		t.Fatal("expected server.listen to equal :2525")
	}
	if cfg["users"]["bob"] == "" {
		t.Fatal("expected users.bob to be present")
	}
	if cfg["server"]["relay"] == cfg["server"]["nonexistent-key"] {
		t.Fatal("the 'relay' key shouldn't be comparable to a nonexistent key")
	}
}

func TestParsing2(t *testing.T) {
	example := `roofgraf {
		dep rg-api, rg-web
	}
	
	rg-api {
		dep redis
		run node index.js
	}
	
	rg-web {
		dep rg-dll
		run yarn dev
	}
	
	rg-dll {
		cmd yarn build:dll
	}
	
	# Any stateful process should be listed as a separate block
	redis {
		dir roofgraf
		run redis-server redis.conf
	}
	
	solargraf {
		dep redis, sg-gateway, sg-web
	}
	
	sg-gateway {
		run yarn dev:gateway
	}
	
	sg-web {
		run yarn dev
	}
	`
	cfg, err := parseString(example)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["sg-web"]["run"] != "yarn dev" {
		t.Fatal("expected sg-web.run to be 'yarn dev'")
	}
}
