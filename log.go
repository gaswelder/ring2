package main

import "fmt"
import "os"

func debMsg(tpl string, args ...interface{}) {
	if !config.debug {
		return
	}
	fmt.Fprintf(os.Stderr, tpl+"\n", args...)
}
