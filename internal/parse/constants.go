package parse

const (
	itemError itemType = iota // error occurred; value is text of error.
	itemEOF
	itemComment
	itemSpace
	itemString     // quoted string (including quotes)
	itemVariable   // variable starting with '$', such as '$hello'
	itemLeftBlock  // left block delimiter
	itemRightBlock // right block delimiter
)

const eof = -1
