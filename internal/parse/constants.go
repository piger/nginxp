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

// These contexts maps to: https://github.com/nginxinc/crossplane/blob/ad3d23094bbd8b1f5586b48b883b2e48d6587e49/crossplane/analyzer.py#L2095
// Their purpose is to be used with the map that follows to associate a context with a bitmask.
const (
	contextRoot                    confContext = iota // main configuration section
	contextEvents                                     // events
	contextMail                                       // mail
	contextMailServer                                 // mail -> server
	contextStream                                     // stream
	contextStreamServer                               // stream -> server
	contextStreamUpstream                             // stream -> upstream
	contextHttp                                       // http
	contextHttpServer                                 // http -> server
	contextHttpLocation                               // http -> location
	contextHttpUpstream                               // http -> upstream
	contextHttpServerIf                               // http -> server -> if
	contextHttpLocationIf                             // http -> location -> if
	contextHttpLocationLimitExcept                    // http -> location -> limit_except
)

var contextBitmask = map[confContext]int{
	contextRoot:                    NGX_MAIN_CONF,
	contextEvents:                  NGX_EVENT_CONF,
	contextMail:                    NGX_MAIL_MAIN_CONF,
	contextMailServer:              NGX_MAIL_SRV_CONF,
	contextStream:                  NGX_STREAM_MAIN_CONF,
	contextStreamServer:            NGX_STREAM_SRV_CONF,
	contextStreamUpstream:          NGX_STREAM_UPS_CONF,
	contextHttp:                    NGX_HTTP_MAIN_CONF,
	contextHttpServer:              NGX_HTTP_SRV_CONF,
	contextHttpLocation:            NGX_HTTP_LOC_CONF,
	contextHttpUpstream:            NGX_HTTP_UPS_CONF,
	contextHttpServerIf:            NGX_HTTP_SIF_CONF,
	contextHttpLocationIf:          NGX_HTTP_LIF_CONF,
	contextHttpLocationLimitExcept: NGX_HTTP_LMT_CONF,
}

type context map[string]bool

var ctxLevels = []string{"root", "events", "mail", "server", "stream", "upstream", "http", "location", "if", "limit_except"}

func NewCtx() *context {
	c := make(context)
	for _, lvl := range ctxLevels {
		c[lvl] = false
	}
	return &c
}

func (c context) Push(level string) {
	if _, ok := c[level]; !ok {
		panic(fmt.Sprintf("unknown context level %q", level))
	}
	c[level] = true
}

func (c context) Pop(level string) {
	if _, ok := c[level]; !ok {
		panic(fmt.Sprintf("unknown context level %q", level))
	}
	c[level] = false
}

func (c context) ID() confContext {
	switch {
	case c["events"]:
		return contextEvents
	case c["mail"] && c["server"]:
		return contextMailServer
	case c["mail"]:
		return contextMail
	case c["stream"] && c["upstream"]:
		return contextStreamUpstream
	case c["stream"] && c["server"]:
		return contextStreamServer
	case c["stream"]:
		return contextStream
	case c["http"] && c["location"] && c["limit_except"]:
		return contextHttpLocationLimitExcept
	case c["http"] && c["location"] && c["if"]:
		return contextHttpLocationIf
	case c["http"] && c["server"] && c["if"]:
		return contextHttpServerIf
	case c["http"] && c["upstream"]:
		return contextHttpUpstream
	case c["http"] && c["location"]:
		return contextHttpLocation
	case c["http"] && c["server"]:
		return contextHttpServer
	case c["http"]:
		return contextHttp
	case c["root"]:
		return contextRoot
	}
	panic("no context")
}
