package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// itemType identifies the type of lex items.
type itemType int

// item represent a token or text strin returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 20 {
		return fmt.Sprintf("%.20q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// stateFn represents the state of the scanner as a function that returns
// the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name      string    // used only for error reports.
	input     string    // the string being scanned
	start     Pos       // start position of this item
	pos       Pos       // current position in the input
	width     Pos       // width of the last rune read
	items     chan item // channel of scanned items
	line      int       // 1+number of newlines seen
	startLine int       // start line of this item
	depth     int       // depth level of nested blocks
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
	l.start = l.pos
	l.startLine = l.line
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
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

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
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

// run lexes the input by executing state functions until the state is nil.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items) // no more tokens will be delivered.
}

// state functions

// lexText scans the input until EOF.
func lexText(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		if l.depth > 0 {
			return l.errorf("unclosed block")
		}
		l.emit(itemEOF)
		return nil
	case r == '"' || r == '\'':
		return lexQuote
	case r == '#':
		l.ignore()
		return lexComment
	case r == ';':
		l.emit(itemTerminator)
		return lexText
	case r == '{':
		l.depth++
		l.emit(itemLeftBlock)
		return lexText
	case r == '}':
		l.depth--
		if l.depth < 0 {
			return l.errorf("unmatched closing block")
		}
		l.emit(itemRightBlock)
		return lexText
	case r == '\n':
		l.emit(itemNewline)
		return lexText
	case isSpace(r):
		l.ignore()
		return lexText
	default:
		l.backup()
		return lexWord
	}
}

// lexQuote scans a quoted string (with either single or double quotes).
func lexQuote(l *lexer) stateFn {
	marker := rune(l.input[l.start])
Loop:
	for {
		switch r := l.next(); {
		case r == '\\':
			if n := l.next(); n != eof && n != '\n' {
				break
			}
			fallthrough
		case r == eof || r == '\n':
			return l.errorf("unterminated quoted string")
		case r == marker:
			break Loop
		}
	}
	l.emit(itemString)
	return lexText
}

// lexComment scans a comment. The left marker is known to be present.
func lexComment(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof, '\n':
			l.backup()
			break Loop
		}
	}
	l.emit(itemComment)
	return lexText
}

// lexWord scans a word, which can be a directive or an argument for a directive.
// A "word" can be terminated by space, ';' or the start of a new block '{'.
func lexWord(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isSpace(r) || r == ';' || r == '{':
			break Loop
		case r == eof:
			return l.errorf("unterminated line") // XXX
		}
	}
	l.backup()
	l.emit(itemWord)
	return lexText
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
