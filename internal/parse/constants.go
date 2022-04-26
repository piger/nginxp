package parse

import "fmt"

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

// itemName maps item types to names that can be prettyprinted.
var itemName = map[itemType]string{
	itemError:      "error",
	itemEOF:        "EOF",
	itemComment:    "comment",
	itemString:     "quoted string",
	itemWord:       "word",
	itemNewline:    "newline",
	itemTerminator: "terminator",
	itemLeftBlock:  "open block",
	itemRightBlock: "close block",
	itemSpace:      "whitespace",
	itemVariable:   "variable",
	itemDirective:  "directive",
	itemArgument:   "argument",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

const eof = -1
