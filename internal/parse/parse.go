package parse

import (
	"fmt"
	"runtime"
)

// Tree is the representation of a single parsed file.
type Tree struct {
	Filename string    // name of the file represented by this tree.
	Root     *ListNode // top-level root of the tree.
	text     string    // text parsed to create this Tree.
	// Parsing only; cleared after parse.
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

func (t *Tree) stopParse() {
	t.lex = nil
}

func (t *Tree) Parse(text string) (tree *Tree, err error) {
	defer t.recover(&err)

	t.startParse(lex(t.Filename, text))
	t.text = text
	t.parse()
	t.stopParse()
	return t, nil
}

func (t *Tree) parse() {
	t.Root = t.newList(t.peek().pos)
	for t.peek().typ != itemEOF {
		switch t.peek().typ {
		case itemWord:
			t.Root.append(t.parseDirective())
		case itemComment:
			// should create a CommentNode and append.
			// t.Root.append(...)
		}
	}
}

func (t *Tree) parseDirective() Node {
	item := t.next()
	n := t.newDirective(item.pos, item.val)

Loop:
	for {
		p := t.peek()
		switch p.typ {
		case itemWord:
			item := t.next()
			arg := t.newArgument(item.pos, item.val)
			n.append(arg)
		case itemLeftBlock:
			// should parse the entire block until itemRightBlock
			// should a block node just be a ListNode? Not really because a Block node
			// I think should know its own context...
		case itemTerminator:
			break Loop
		default:
			t.errorf("unterminated directive")
		}
	}

	return n
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.lex.drain()
			t.stopParse()
		}
		*errp = e.(error)
	}
}

func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil // XXX why?
	panic(fmt.Errorf(format, args...))
}

func (t *Tree) error(err error) {
	t.errorf("%s", err)
}

func Parse(name, text string) (*Tree, error) {
	t := &Tree{
		Filename: name,
		Root:     nil,
	}

	_, err := t.Parse(text)
	if err != nil {
		panic(err)
	}

	return t, nil
}
