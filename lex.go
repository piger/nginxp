package nginxp

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemStatement // a single configuration statement, like "location"
	itemSpace
	itemTerminator // the expression terminator (;)
	itemOpenBlock  // opening of a block ({)
	itemCloseBlock // closing of a block (})
	itemExpression // a series of statements
	itemComment    // a comment
	itemLeftDelim  // template opening delimiter
	itemRightDelim // template closing delimiter
	itemTemplateVar
	itemString
	itemNewLine
)

// Pos represents a byte position in the original input text from which
// this thing was parsed
type Pos int

// Position satisifies some kind of interface...
func (p Pos) Position() Pos {
	return p
}

// item represents a token or text string returned from the scanner
type item struct {
	typ  itemType
	pos  Pos
	val  string
	line int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	}
	return fmt.Sprintf("<%q>", i.val)
}

const eof = -1

const (
	spaceChars = " \t\r\n"
	leftDelim  = "{{"
	rightDelim = "}}"
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of () exps
	line       int       // 1+number of lines seen
	startLine  int       // start line of this item
}

// next returns the next rune in the input
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
	l.start = l.pos
	l.startLine = l.line
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

func (l *lexer) nextItem() item {
	return <-l.items
}

func (l *lexer) drain() {
	for range l.items {

	}
}

func lex(name, input string) *lexer {
	l := &lexer{
		name:      name,
		input:     input,
		items:     make(chan item),
		line:      1,
		startLine: 1,
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexText(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == ';':
		l.emit(itemTerminator)
	case r == '{':
		nextc := l.peek()
		if nextc == '{' {
			return lexTemplateVariable
		}

		l.emit(itemOpenBlock)
	case r == '}':
		l.emit(itemCloseBlock)
	case r == '#':
		return lexComment
	case r == '\'' || r == '"':
		l.backup()
		return lexQuote
	case isSpace(r):
		return lexSpace
	case isEndOfLine(r):
		return lexNewlines
	case r == eof:
		return nil
	default:
		return lexStatement
	}

	// fmt.Printf("l.pos = %d, len(input)=%d\n", l.pos, len(l.input))
	if l.pos == Pos(len(l.input)) {
		fmt.Printf("EOF reached")
		return nil
	}

	return lexText
}

func lexComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], "\n")
	if i < 0 {
		// EOF
		l.pos = Pos(len(l.input))
		l.emit(itemEOF)
		return nil
	}
	l.pos += Pos(i)
	l.emit(itemComment)
	return lexNewlines
}

func lexNewlines(l *lexer) stateFn {
	var r rune
	for {
		r = l.peek()
		if !isEndOfLine(r) {
			break
		}
		l.next()
		// next increments the newline counter!
	}
	l.emit(itemNewLine)
	return lexText
}

func lexSpace(l *lexer) stateFn {
	var r rune
	for {
		r = l.peek()
		if !isSpace(r) {
			break
		}
		l.next()
	}
	l.emit(itemSpace)
	return lexText
}

// must read until: end of line (\n), end of statement (;), space ( ), start of block ({) (not end of block!)
// XXX yeah but what about quoted string and other things found in statements??
func lexStatement(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case r == '\n' || r == ';' || r == '{' || r == eof || r == ' ' || r == '\'' || r == '"':
			break Loop
		default:
			// consume
		}
	}
	l.backup()
	l.emit(itemStatement)
	return lexText
}

func lexQuote(l *lexer) stateFn {
	// this can either be a single quote or a double quote
	quoteChr := l.next()

Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case quoteChr:
			break Loop
		}
	}
	l.emit(itemString)
	return lexText
}

func lexTemplateVariable(l *lexer) stateFn {
	if x := strings.Index(l.input[l.pos:], rightDelim); x >= 0 {
		l.pos += Pos(x) + Pos(len(rightDelim))
		l.emit(itemTemplateVar)

		return lexText
	}
	return l.errorf("unterminated template variable")

}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
