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

// XXX not sure if this should return a Tree, we don't expect to have multiple trees.
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
		default:
			t.next()
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
		case itemWord, itemString:
			// XXX should itemString be un-quoted before being added as a Node?
			item := t.next()
			arg := t.newArgument(item.pos, item.val)
			n.append(arg)
		case itemLeftBlock:
			// should parse the entire block until itemRightBlock
			// should a block node just be a ListNode? Not really because a Block node
			// I think should know its own context...
			t.next() // drain
		case itemTerminator:
			break Loop
		case itemNewline:
			node := t.parseEmptyLines()
			if node != nil {
				t.Root.append(node)
			}
		default:
			t.errorf("unterminated directive: found %s", p.typ)
		}
	}

	return n
}

// parseEmptyLines parse one or more newlines; it only emits a EmptyLineNode
// when one or more _empty lines_ are found.
// The general idea is that we don't care about newlines, but we care to keep
// empty lines (multiple empty lines compressed into one) so that we can format the
// tree later.
func (t *Tree) parseEmptyLines() Node {
	var res Node
	this := t.next()

	for t.peek().typ == itemNewline {
		// discard all following newlines, but ensure we return at least one.
		t.next()
		if res == nil {
			res = t.newEmptyLine(this.pos)
		}
	}

	return res
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
