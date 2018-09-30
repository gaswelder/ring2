package pop

import (
	"errors"
	"log"
	"strconv"

	"github.com/gaswelder/ring2/server/mailbox"
)

// popMessageEntry represents a message entry used
// in a POP session. It contains a cached message entry
// from the mailbox plus session-related data like ID
// and the `deleted` flag.
type popMessageEntry struct {
	Id      int
	Msg     *mailbox.Message
	deleted bool
}

type InboxView struct {
	box         *mailbox.Mailbox
	messageList []*popMessageEntry
	lastID      int
}

func (i *InboxView) LastID() int {
	return i.lastID
}

func NewInboxView(box *mailbox.Mailbox) (*InboxView, error) {
	// List the letters in the given box and assign them
	// session identifiers.
	id := 0
	ls, err := box.List()
	if err != nil {
		return nil, err
	}
	messageList := make([]*popMessageEntry, 0)
	for _, msg := range ls {
		id++
		log.Println(id, msg)
		messageList = append(messageList, &popMessageEntry{
			Id:      id,
			Msg:     msg,
			deleted: false,
		})
	}

	v := &InboxView{
		box:         box,
		messageList: messageList,
	}

	err = v.resetLastID()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *InboxView) resetLastID() error {
	last, err := v.box.LastRetrievedMessage()
	if err != nil {
		return err
	}
	if last == nil {
		v.lastID = 0
		return nil
	}
	for _, entry := range v.messageList {
		if entry.Msg.Filename() == last.Filename() {
			v.lastID = entry.Id
			return nil
		}
	}
	return errors.New("failed to find the lastID message")
}

func (v *InboxView) Entries() []*popMessageEntry {
	list := make([]*popMessageEntry, 0)
	for _, msg := range v.messageList {
		if msg.deleted {
			continue
		}
		list = append(list, msg)
	}
	return list
}

func (v *InboxView) Reset() error {
	// Reset all deleted flags
	for _, entry := range v.messageList {
		entry.deleted = false
	}
	err := v.resetLastID()
	return err
}

func (v *InboxView) MarkRetrieved(entry *popMessageEntry) {
	if entry.Id > v.lastID {
		v.lastID = entry.Id
	}
}

func (v *InboxView) FindEntryByID(msgid string) *popMessageEntry {
	if msgid == "" {
		return nil
	}
	id, err := strconv.Atoi(msgid)
	if err != nil {
		return nil
	}
	for _, entry := range v.messageList {
		if entry.Id == id {
			return entry
		}
	}
	return nil
}

func (v *InboxView) MarkAsDeleted(msgid string) error {
	e := v.FindEntryByID(msgid)
	if e == nil {
		return errors.New("no such message")
	}
	e.deleted = true
	return nil
}

func (v *InboxView) Commit() error {
	if v.box == nil {
		return nil
	}
	last := v.lastRetrievedEntry()
	if last != nil {
		v.box.SetLast(last.Msg)
	}
	err := v.purge()
	return err
}

func (v *InboxView) lastRetrievedEntry() *popMessageEntry {
	for _, entry := range v.Entries() {
		if entry.Id == v.lastID {
			return entry
		}
	}
	return nil
}

func (v *InboxView) Stat() (count int, size int64, err error) {
	for _, entry := range v.Entries() {
		count++
		size += entry.Msg.Size()
	}
	return count, size, err
}

// Remove messages marked to be deleted
func (v *InboxView) purge() error {
	l := make([]*popMessageEntry, 0)

	for _, entry := range v.messageList {
		if !entry.deleted {
			l = append(l, entry)
			continue
		}
		err := v.box.Remove(entry.Msg)
		if err != nil {
			return err
		}
	}
	v.messageList = l
	return nil
}
