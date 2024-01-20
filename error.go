package notify


type Error interface {
	Temporary() bool
}

func IsTemporary(err error) bool {
	if e, ok := err.(Error); ok {
		return e.Temporary()
	}

	return false
}
