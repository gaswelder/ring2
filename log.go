package main

import "fmt"
import "os"

var debugLog = false

func debMsg(tpl string, args ...interface{}) {
	if !debugLog {
		return
	}
	fmt.Fprintf(os.Stderr, tpl+"\n", args...)
}
