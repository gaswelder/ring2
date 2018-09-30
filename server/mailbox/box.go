package mailbox

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
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

type Mailbox struct {
	path string
}

func New(path string) (*Mailbox, error) {
	box := &Mailbox{path}
	err := createDir(path)
	if err != nil {
		return nil, err
	}
	return box, nil
}

// Machinery to sort FileInfo arrays
type kludge []os.FileInfo

func (a kludge) Len() int      { return len(a) }
func (a kludge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a kludge) Less(i, j int) bool {
	return strings.Compare(a[i].Name(), a[j].Name()) == -1
}

func (b *Mailbox) List() ([]*Message, error) {
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
	messages := make([]*Message, 0)
	for _, info := range files {
		if info.Name()[0] == '.' {
			continue
		}
		if info.Name() == "last" {
			continue
		}

		m := new(Message)
		m.size = info.Size()
		m.filename = info.Name()
		m.path = b.path + "/" + m.filename
		messages = append(messages, m)
	}
	return messages, nil
}

func (b *Mailbox) LastRetrievedMessage() (*Message, error) {
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
	m := new(Message)
	m.size = stat.Size()
	m.filename = lastName
	m.path = b.path + "/" + lastName
	return m, nil
}

func (b *Mailbox) Remove(msg *Message) error {
	return os.Remove(b.path + "/" + msg.filename)
}

func (b *Mailbox) Add(text string) error {
	err := b.lock()
	if err != nil {
		return err
	}
	defer b.unlock()

	name := time.Now().Format("20060102-150405-") + fmt.Sprintf("%x", md5.Sum([]byte(text)))
	return b.writeFile(name, text)
}

// Scan the directory and define b.messages and b.lastID fields.
// func (b *Mailbox) parseMessages() error {
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
// 	b.messages = make([]*Message, 0)
// 	id := 0
// 	for _, info := range files {
// 		if info.Name()[0] == '.' {
// 			continue
// 		}

// 		if info.Name() == "last" {
// 			continue
// 		}

// 		m := new(Message)
// 		m.size = info.Size()
// 		m.filename = info.Name()
// 		m.path = b.path + "/" + m.filename
// 		b.messages = append(b.messages, m)
// 	}

// 	return nil
// }

// Returns contents of a file in the directory
func (b *Mailbox) readFile(name string) (string, error) {
	path := b.path + "/" + name
	val, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// Writes data to a file in the directory
func (b *Mailbox) writeFile(name string, data string) error {
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

func (b *Mailbox) lock() error {
	register.Lock()
	defer register.Unlock()

	_, ok := register.boxes[b.path]
	if ok {
		return errors.New("Busy")
	}
	return nil
}

func (b *Mailbox) unlock() error {
	register.Lock()
	defer register.Unlock()
	delete(register.boxes, b.path)
	return nil
}

// Update the 'last' to point to the message with the given id
func (b *Mailbox) SetLast(msg *Message) {
	b.writeFile("last", msg.filename)
}

func createDir(path string) error {
	stat, err := os.Stat(path)
	if stat != nil && err == nil {
		return nil
	}
	return os.MkdirAll(path, 0755)
}
