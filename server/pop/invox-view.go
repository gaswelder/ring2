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
	id      int
	msg     *mailbox.Message
	deleted bool
}

type inboxView struct {
	box         *mailbox.Mailbox
	messageList []*popMessageEntry
	lastID      int
}

func makeInboxView(box *mailbox.Mailbox) (*inboxView, error) {
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
			id:      id,
			msg:     msg,
			deleted: false,
		})
	}

	v := &inboxView{
		box:         box,
		messageList: messageList,
	}

	err = v.resetLastID()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *inboxView) resetLastID() error {
	last, err := v.box.LastRetrievedMessage()
	if err != nil {
		return err
	}
	if last == nil {
		v.lastID = 0
		return nil
	}
	for _, entry := range v.messageList {
		if entry.msg.Filename() == last.Filename() {
			v.lastID = entry.id
			return nil
		}
	}
	return errors.New("failed to find the lastID message")
}

func (v *inboxView) entries() []*popMessageEntry {
	list := make([]*popMessageEntry, 0)
	for _, msg := range v.messageList {
		if msg.deleted {
			continue
		}
		list = append(list, msg)
	}
	return list
}

func (v *inboxView) reset() error {
	// Reset all deleted flags
	for _, entry := range v.messageList {
		entry.deleted = false
	}
	err := v.resetLastID()
	return err
}

func (v *inboxView) markRetrieved(entry *popMessageEntry) {
	if entry.id > v.lastID {
		v.lastID = entry.id
	}
}

func (v *inboxView) findEntry(msgid string) *popMessageEntry {
	if msgid == "" {
		return nil
	}
	id, err := strconv.Atoi(msgid)
	if err != nil {
		return nil
	}
	for _, entry := range v.messageList {
		if entry.id == id {
			return entry
		}
	}
	return nil
}

func (v *inboxView) markDeleted(msgid string) error {
	e := v.findEntry(msgid)
	if e == nil {
		return errors.New("no such message")
	}
	e.deleted = true
	return nil
}

func (v *inboxView) commit() error {
	if v.box == nil {
		return nil
	}
	last := v.lastRetrievedEntry()
	if last != nil {
		v.box.SetLast(last.msg)
	}
	err := v.purge()
	return err
}

func (v *inboxView) lastRetrievedEntry() *popMessageEntry {
	for _, entry := range v.entries() {
		if entry.id == v.lastID {
			return entry
		}
	}
	return nil
}

func (v *inboxView) stat() (count int, size int64, err error) {
	for _, entry := range v.entries() {
		count++
		size += entry.msg.Size()
	}
	return count, size, err
}

// Remove messages marked to be deleted
func (v *inboxView) purge() error {
	l := make([]*popMessageEntry, 0)

	for _, entry := range v.messageList {
		if !entry.deleted {
			l = append(l, entry)
			continue
		}
		err := v.box.Remove(entry.msg)
		if err != nil {
			return err
		}
	}
	v.messageList = l
	return nil
}
