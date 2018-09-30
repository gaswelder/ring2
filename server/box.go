package server

import (
	"errors"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
)

/*
 * A mailbox is a directory with message files named in some
 * increasing manner (using time, for example). There is also a file
 * "last" which contains the filename of the message that was accessed
 * last.
 *
 * When a POP session in started, all messages are assigned numbers
 * starting from 1. The "last id" in the session then has to be derived
 * from the name. It is also possible that the file named in the "last"
 * was deleted, so in that case the "last id" will be zero.
 */

type mailbox struct {
	path string
}

func newBox(u *UserRec, config *Config) (*mailbox, error) {
	b := new(mailbox)
	b.path = config.Maildir + "/" + u.Name
	err := createDir(b.path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Machinery to sort FileInfo arrays
type kludge []os.FileInfo

func (a kludge) Len() int      { return len(a) }
func (a kludge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a kludge) Less(i, j int) bool {
	return strings.Compare(a[i].Name(), a[j].Name()) == -1
}

func (b *mailbox) list() ([]*message, error) {
	/*
	 * Make sure the directory exists
	 */
	err := createDir(b.path)
	if err != nil {
		return nil, err
	}

	/*
	 * Read and sort all directory entries
	 */
	d, err := os.Open(b.path)
	if err != nil {
		return nil, err
	}
	files, err := d.Readdir(0)
	d.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(kludge(files))

	/*
	 * Scan the directory and fill the messages array
	 */
	messages := make([]*message, 0)
	for _, info := range files {
		if info.Name()[0] == '.' {
			continue
		}
		if info.Name() == "last" {
			continue
		}

		m := new(message)
		m.size = info.Size()
		m.filename = info.Name()
		m.path = b.path + "/" + m.filename
		messages = append(messages, m)
	}
	return messages, nil
}

func (b *mailbox) lastRetrievedMessage() (*message, error) {
	lastName, err := b.readFile("last")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(b.path + "/" + lastName)
	if err != nil {
		return nil, err
	}
	m := new(message)
	m.size = stat.Size()
	m.filename = lastName
	m.path = b.path + "/" + lastName
	return m, nil
}

// Scan the directory and define b.messages and b.lastID fields.
// func (b *mailbox) parseMessages() error {
// 	/*
// 	 * Make sure the directory exists
// 	 */
// 	err := createDir(b.path)
// 	if err != nil {
// 		return err
// 	}

// 	lastName, err := b.readFile("last")
// 	if os.IsNotExist(err) {
// 		lastName = ""
// 		err = nil
// 	}

// 	if err != nil {
// 		return err
// 	}

// 	/*
// 	 * Read and sort all directory entries
// 	 */
// 	d, err := os.Open(b.path)
// 	if err != nil {
// 		return err
// 	}
// 	files, err := d.Readdir(0)
// 	d.Close()
// 	if err != nil {
// 		return err
// 	}
// 	sort.Sort(kludge(files))

// 	/*
// 	 * Scan the directory and fill the messages array
// 	 */
// 	b.messages = make([]*message, 0)
// 	id := 0
// 	for _, info := range files {
// 		if info.Name()[0] == '.' {
// 			continue
// 		}

// 		if info.Name() == "last" {
// 			continue
// 		}

// 		m := new(message)
// 		m.size = info.Size()
// 		m.filename = info.Name()
// 		m.path = b.path + "/" + m.filename
// 		b.messages = append(b.messages, m)
// 	}

// 	return nil
// }

// Returns contents of a file in the directory
func (b *mailbox) readFile(name string) (string, error) {
	path := b.path + "/" + name
	val, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// Writes data to a file in the directory
func (b *mailbox) writeFile(name string, data string) error {
	path := b.path + "/" + name
	return ioutil.WriteFile(path, []byte(data), 0600)
}

var register struct {
	sync.Mutex
	boxes map[string]bool
}

func init() {
	register.boxes = make(map[string]bool)
}

func (b *mailbox) lock() error {
	register.Lock()
	defer register.Unlock()

	_, ok := register.boxes[b.path]
	if ok {
		return errors.New("Busy")
	}
	return nil
}

func (b *mailbox) unlock() error {
	register.Lock()
	defer register.Unlock()
	delete(register.boxes, b.path)
	return nil
}

// Update the 'last' to point to the message with the given id
func (b *mailbox) setLast(msg *message) {
	b.writeFile("last", msg.filename)
}

func createDir(path string) error {
	stat, err := os.Stat(path)
	if stat != nil && err == nil {
		return nil
	}
	return os.MkdirAll(path, 0755)
}
