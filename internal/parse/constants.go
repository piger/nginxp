package parse

import (
	"fmt"
)

//go:generate go run gen_analyser.go
//go:generate go fmt bitmasks.go

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

type confContext int

var contextNames = map[int]string{
	NGX_MAIN_CONF:        "NGX_MAIN_CONF",
	NGX_EVENT_CONF:       "NGX_EVENT_CONF",
	NGX_MAIL_MAIN_CONF:   "NGX_MAIL_MAIN_CONF",
	NGX_MAIL_SRV_CONF:    "NGX_MAIL_SRV_CONF",
	NGX_STREAM_MAIN_CONF: "NGX_STREAM_MAIN_CONF",
	NGX_STREAM_SRV_CONF:  "NGX_STREAM_SRV_CONF",
	NGX_STREAM_UPS_CONF:  "NGX_STREAM_UPS_CONF",
	NGX_HTTP_MAIN_CONF:   "NGX_HTTP_MAIN_CONF",
	NGX_HTTP_SRV_CONF:    "NGX_HTTP_SRV_CONF",
	NGX_HTTP_LOC_CONF:    "NGX_HTTP_LOC_CONF",
	NGX_HTTP_UPS_CONF:    "NGX_HTTP_UPS_CONF",
	NGX_HTTP_SIF_CONF:    "NGX_HTTP_SIF_CONF",
	NGX_HTTP_LIF_CONF:    "NGX_HTTP_LIF_CONF",
	NGX_HTTP_LMT_CONF:    "NGX_HTTP_LMT_CONF",
}

func ConfContextName(c int) string {
	if name, ok := contextNames[c]; ok {
		return name
	}
	return "UNKNOWN"
}

// context is a stack that keeps track of the current context; it is used by the parser
// while navigating the tree (the configuration file). Each time the parser steps into a
// directive whose name appears in `ctxLevels`, it must call Push().
type context map[string]bool

var ctxLevels = []string{"root", "events", "mail", "server", "stream", "upstream", "http", "location", "if", "limit_except"}

func NewCtx() *context {
	c := make(context)
	for _, lvl := range ctxLevels {
		c[lvl] = false
	}
	return &c
}

func (c context) IsContext(dirname string) bool {
	_, ok := c[dirname]
	return ok
}

// Push should be called before parsing a directive's block.
func (c context) Push(level string) {
	if _, ok := c[level]; !ok {
		panic(fmt.Sprintf("unknown context level %q", level))
	}
	c[level] = true
}

// Pop should be called after parsing a directive's block.
func (c context) Pop(level string) {
	if _, ok := c[level]; !ok {
		panic(fmt.Sprintf("unknown context level %q", level))
	}
	c[level] = false
}

// curContext return the current context; to determine the current context we check which contexts
// have been "activated" in the stack.
func (c context) curContext() int {
	switch {
	case c["events"]:
		return NGX_EVENT_CONF
	case c["mail"] && c["server"]:
		return NGX_MAIL_SRV_CONF
	case c["mail"]:
		return NGX_MAIL_MAIN_CONF
	case c["stream"] && c["upstream"]:
		return NGX_STREAM_UPS_CONF
	case c["stream"] && c["server"]:
		return NGX_STREAM_SRV_CONF
	case c["stream"]:
		return NGX_STREAM_MAIN_CONF
	case c["http"] && c["location"] && c["limit_except"]:
		return NGX_HTTP_LMT_CONF
	case c["http"] && c["location"] && c["if"]:
		return NGX_HTTP_LIF_CONF
	case c["http"] && c["server"] && c["if"]:
		return NGX_HTTP_SIF_CONF
	case c["http"] && c["upstream"]:
		return NGX_HTTP_UPS_CONF
	case c["http"] && c["location"]:
		return NGX_HTTP_LOC_CONF
	case c["http"] && c["server"]:
		return NGX_HTTP_SRV_CONF
	case c["http"]:
		return NGX_HTTP_MAIN_CONF
	case c["root"]:
		return NGX_MAIN_CONF
	}
	panic("no context")
}
