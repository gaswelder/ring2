package main

import (
	"bytes"
	"fmt"
)

type bufFmt struct {
	b bytes.Buffer
}

func (b *bufFmt) put(format string, args ...interface{}) {
	format += "\r\n"
	fmt.Fprintf(&b.b, format, args...)
}

func (b *bufFmt) String() string {
	return b.b.String()
}
