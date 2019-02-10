package mailbox

import "io/ioutil"

// Message represents a message saved in the mailbox.
type Message struct {
	size     int64
	path     string
	filename string
}

// Content returns contents of the message.
func (m *Message) Content() (string, error) {
	v, err := ioutil.ReadFile(m.path)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// Size returns size of the message in bytes.
func (m *Message) Size() int64 {
	return m.size
}

// Filename returns local filename of the message.
func (m *Message) Filename() string {
	return m.filename
}
