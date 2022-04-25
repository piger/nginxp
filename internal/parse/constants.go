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
	itemDirective  // a configuration directive, such as "server", "location" or "proxy_set_header"
	itemArgument   // an argument of a directive, which can be a quoted string or a raw string
	itemTerminator // the character ';' which terminates a directive
	itemWord
	itemNewline
)

const eof = -1

var itemName = map[itemType]string{
	itemError:   "error",
	itemEOF:     "EOF",
	itemComment: "comment",
	itemString:  "quoted string",
	itemWord:    "word",
	itemNewline: "newline",
}
