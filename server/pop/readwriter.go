package pop

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type readWriter struct {
	writer io.Writer
	reader *bufio.Reader
}

func makeReadWriter(c io.ReadWriter) *readWriter {
	return &readWriter{
		writer: c,
		reader: bufio.NewReader(c),
	}
}

func (rw *readWriter) readCommand() (*command, error) {
	line, err := rw.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return parseCommand(line)
}

// Send a success response with optional comment
func (rw *readWriter) OK(comment string, args ...interface{}) {
	if comment != "" {
		rw.Send("+OK " + fmt.Sprintf(comment, args...))
	} else {
		rw.Send("+OK")
	}
}

// Send an error response with optinal comment
func (rw *readWriter) Err(comment string) {
	if comment != "" {
		rw.Send("-ERR " + comment)
	} else {
		rw.Send("-ERR")
	}
}

// Send a line
func (rw *readWriter) Send(format string, args ...interface{}) error {
	line := fmt.Sprintf(format+"\r\n", args...)
	_, err := rw.writer.Write([]byte(line))
	return err
}

// Send a multiline data
func (rw *readWriter) SendData(data string) error {
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
func (rw *readWriter) SendDataLine(line string) error {
	if len(line) > 0 && line[0] == '.' {
		line = "." + line
	}
	line += "\r\n"
	_, err := rw.writer.Write([]byte(line))
	return err
}
