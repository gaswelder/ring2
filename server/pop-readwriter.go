package server

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type popReadWriter struct {
	writer io.Writer
	reader *bufio.Reader
}

func (rw *popReadWriter) readCommand() (*command, error) {
	line, err := rw.reader.ReadString('\n')
	debMsg(line)
	if err != nil {
		return nil, err
	}
	return parseCommand(line)
}

// Send a success response with optional comment
func (rw *popReadWriter) ok(comment string, args ...interface{}) {
	if comment != "" {
		rw.send("+OK " + fmt.Sprintf(comment, args...))
	} else {
		rw.send("+OK")
	}
}

// Send an error response with optinal comment
func (rw *popReadWriter) err(comment string) {
	if comment != "" {
		rw.send("-ERR " + comment)
	} else {
		rw.send("-ERR")
	}
}

// Send a line
func (rw *popReadWriter) send(format string, args ...interface{}) error {
	line := fmt.Sprintf(format+"\r\n", args...)
	debMsg("pop send: %s", line)
	_, err := rw.writer.Write([]byte(line))
	return err
}

// Send a multiline data
func (rw *popReadWriter) sendData(data string) error {
	var err error
	lines := strings.Split(data, "\r\n")
	for _, line := range lines {
		err = rw.sendDataLine(line)
		if err != nil {
			return err
		}
	}
	err = rw.send(".")
	return err
}

// Sends a line of data, taking care of the "dot-stuffing"
func (rw *popReadWriter) sendDataLine(line string) error {
	if len(line) > 0 && line[0] == '.' {
		line = "." + line
	}
	line += "\r\n"
	_, err := rw.writer.Write([]byte(line))
	return err
}
