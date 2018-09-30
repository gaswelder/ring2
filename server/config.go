package server

// Config is a structure to keep user-provided
// server parameters.
type Config struct {
	Debug    bool
	Hostname string
	Maildir  string
	Smtp     string
	Pop      string
	Lists    map[string][]*UserRec
	Users    map[string]*UserRec
}
