package pop

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ReadWriter struct {
	writer io.Writer
	reader *bufio.Reader
}

func NewReadWriter(c io.ReadWriter) *ReadWriter {
	return &ReadWriter{
		writer: c,
		reader: bufio.NewReader(c),
	}
}

func (rw *ReadWriter) ReadCommand() (*Command, error) {
	line, err := rw.reader.ReadString('\n')
	// debMsg(line)
	if err != nil {
		return nil, err
	}
	return parseCommand(line)
}

// Send a success response with optional comment
func (rw *ReadWriter) OK(comment string, args ...interface{}) {
	if comment != "" {
		rw.Send("+OK " + fmt.Sprintf(comment, args...))
	} else {
		rw.Send("+OK")
	}
}

// Send an error response with optinal comment
func (rw *ReadWriter) Err(comment string) {
	if comment != "" {
		rw.Send("-ERR " + comment)
	} else {
		rw.Send("-ERR")
	}
}

// Send a line
func (rw *ReadWriter) Send(format string, args ...interface{}) error {
	line := fmt.Sprintf(format+"\r\n", args...)
	// debMsg("pop send: %s", line)
	_, err := rw.writer.Write([]byte(line))
	return err
}

// Send a multiline data
func (rw *ReadWriter) SendData(data string) error {
	var err error
	lines := strings.Split(data, "\r\n")
	for _, line := range lines {
		err = rw.SendDataLine(line)
		if err != nil {
			return err
		}
	}
	err = rw.Send(".")
	return err
}

// Sends a line of data, taking care of the "dot-stuffing"
func (rw *ReadWriter) SendDataLine(line string) error {
	if len(line) > 0 && line[0] == '.' {
		line = "." + line
	}
	line += "\r\n"
	_, err := rw.writer.Write([]byte(line))
	return err
}
