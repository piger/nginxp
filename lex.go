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

const eof = -1

const (
	spaceChars = " \t\r\n"
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
	switch r := l.next(); {
	case r == ';':
		l.emit(itemTerminator)
	case r == '{':
		l.emit(itemOpenBlock)
	case r == '}':
		l.emit(itemCloseBlock)
	case r == '#':
		return lexComment
	}
	return nil
}

func lexComment(l *lexer) stateFn {
	return nil
}
