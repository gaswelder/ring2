package main

import (
	"log"

	"github.com/gaswelder/ring2/server"
)

func main() {

	config, err := readConfig("conf")
	if err != nil {
		log.Fatal(err)
	}
	server.Run(config)
	select {}
}
