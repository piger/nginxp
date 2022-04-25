package parse

const (
	itemError itemType = iota // error occurred; value is text of error.
	itemEOF

	eof = -1
)
