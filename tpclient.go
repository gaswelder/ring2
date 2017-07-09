package main

import (
	"fmt"
	"io"
)

// "Transmission protocol" reader/writer from the client side.
// Provides handling for response format used in FTP and SMTP.
type tpclient struct {
	f   io.ReadWriter
	err error
}

// Returns a new tpclient
func newTpClient(f io.ReadWriter) *tpclient {
	return &tpclient{f, nil}
}

// Reads a status response and returns true if the
// response code is the same as the argument. If not,
// sets an error and returns false.
func (w *tpclient) Expect(code int) bool {
	if w.err != nil {
		return false
	}

	var rcode int
	var rmsg string
	var n int

	n, w.err = fmt.Fscanf(w.f, "%d %s\r\n", &rcode, &rmsg)
	if w.err != nil {
		return false
	}
	if n < 2 {
		w.err = fmt.Errorf("Couldn't scan the line")
		return false
	}
	if rcode != code {
		w.err = fmt.Errorf("%d response expected, got %d %s", code, rcode, rmsg)
		return false
	}

	return true
}

// Sends a free-form line terminated with "\r\n".
// The arguments themselves must not contain "\r\n".
func (w *tpclient) WriteLine(format string, args ...interface{}) {
	if w.err != nil {
		return
	}
	format += "\r\n"
	fmt.Fprintf(w.f, format, args...)
}

// Sends an arbitrary string exactly as is.
func (w *tpclient) Write(s string) {
	if w.err != nil {
		return
	}
	io.WriteString(w.f, s)
}

// Returns internal error state.
func (w *tpclient) Err() error {
	return w.err
}
