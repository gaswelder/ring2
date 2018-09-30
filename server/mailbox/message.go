package mailbox

import "io/ioutil"

type Message struct {
	size     int64
	path     string
	filename string
}

// Returns contents of the message
func (m *Message) Content() (string, error) {
	v, err := ioutil.ReadFile(m.path)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (m *Message) Size() int64 {
	return m.size
}

func (m *Message) Filename() string {
	return m.filename
}
