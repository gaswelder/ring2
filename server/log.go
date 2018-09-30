package server

import "fmt"
import "os"

var debugLog = true

func debMsg(tpl string, args ...interface{}) {
	if !debugLog {
		return
	}
	fmt.Fprintf(os.Stderr, tpl+"\n", args...)
}
