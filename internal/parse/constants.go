package parse

import "fmt"

const (
	itemError itemType = iota // error occurred; value is text of error.
	itemEOF
	itemComment    // a comment, which can occupy the whole line or be inline.
	itemString     // quoted string (including quotes)
	itemWord       // a generic word, which can be a directive or an argument for a directive
	itemNewline    // a newline token
	itemTerminator // the character ';' which terminates a directive
	itemLeftBlock  // left block delimiter
	itemRightBlock // right block delimiter
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
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

const eof = -1
