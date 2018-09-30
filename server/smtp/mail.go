package smtp

/*
 * A mail draft
 */
type Mail struct {
	Sender     *Path
	Recipients []*Path
}

func NewDraft(from *Path) *Mail {
	return &Mail{
		from,
		make([]*Path, 0),
	}
}
