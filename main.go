package main

import "log"

func main() {

	err := readConfig("conf")
	if err != nil {
		log.Fatal(err)
	}

	runSMTP()
}

