package mailbox

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// Mailbox is a directory with message files and a designated "last
// retrieved" message.
type Mailbox struct {
	path string
}

// New returns a mailbox that keeps its data in the
// given directory.
func New(path string) (*Mailbox, error) {
	box := &Mailbox{path}
	err := createDir(path)
	if err != nil {
		return nil, err
	}
	return box, nil
}

// Name returns the name of the mailbox for logging purposes.
func (b *Mailbox) Name() string {
	return b.path
}

// Machinery to sort FileInfo arrays
type kludge []os.FileInfo

func (a kludge) Len() int      { return len(a) }
func (a kludge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a kludge) Less(i, j int) bool {
	return strings.Compare(a[i].Name(), a[j].Name()) == -1
}

// List returns a list of messages stored in this mailbox.
func (b *Mailbox) List() ([]*Message, error) {
	// If the directory doesn't exist, we assume that this mailbox simply
	// haven't been written to, and return an empty list.
	_, err := os.Stat(b.path)
	if os.IsNotExist(err) {
		return make([]*Message, 0), nil
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

// LastRetrievedMessage returns the message marked as "last retrieved".
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

// SetLast sets the mailbox's "last retrieved message" pointer to the given message.
func (b *Mailbox) SetLast(msg *Message) {
	log.Printf("Setting last message to %s", msg.filename)
	b.writeFile("last", msg.filename)
}

// Remove removes the given message from the mailbox.
func (b *Mailbox) Remove(msg *Message) error {
	log.Printf("Deleting message %s", msg.filename)
	return os.Remove(b.path + "/" + msg.filename)
}

// Add creates a new message from the given text and saves it to the mailbox.
func (b *Mailbox) Add(text string) error {
	name := time.Now().Format("20060102-150405-") + fmt.Sprintf("%x", md5.Sum([]byte(text)))
	log.Printf("Saving message %s", name)
	return b.writeFile(name, text)
}

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
	err := createDir(b.path)
	if err != nil {
		return err
	}
	path := b.path + "/" + name
	return ioutil.WriteFile(path, []byte(data), 0600)
}

func createDir(path string) error {
	stat, err := os.Stat(path)
	if stat != nil && err == nil {
		return nil
	}
	return os.MkdirAll(path, 0755)
}
