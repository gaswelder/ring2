package main

import (
	"net/smtp"
	"testing"
)

func TestMain(t *testing.T) {
	var err error

	addr := "localhost:2525"
	msg := "From: nobody\r\nSubject: whatever\r\n\r\nHey you!"

	t.Run("no auth", func(t *testing.T) {
		err = smtp.SendMail(addr, nil, "nobody@localhost", []string{"joe@localhost"}, []byte(msg))
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("plain auth", func(t *testing.T) {
		plain := smtp.PlainAuth("", "joe", "123", "localhost")
		err = smtp.SendMail(addr, plain, "nobody@localhost", []string{"joe@localhost"}, []byte(msg))
		if err != nil {
			t.Error(err)
		}
	})
}
