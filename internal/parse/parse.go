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
	ctx := NewCtx()
	ctx.Push("root")

	for t.peek().typ != itemEOF {
		switch t.peek().typ {
		case itemWord:
			t.Root.append(t.parseDirective(ctx, true))
		case itemNewline:
			if node := t.parseEmptyLines(); node != nil {
				t.Root.append(node)
			}
		case itemComment:
			item := t.next()
			t.Root.append(t.newComment(item.pos, item.val))
		default:
			t.errorf("unexpected item in parse: %s", t.peek().typ)
		}
	}
}

// isFreeFormDirective is used to determine certain kind of directives for which validation
// should be skipped, because they are not actual directives but just "arguments" for maps,
// split_clients, etc.
func isFreeFormDirective(name string) bool {
	for _, d := range []string{"map", "split_clients", "geo", "types", "match"} {
		if name == d {
			return true
		}
	}
	return false
}

func isLuaCodeBlockDirective(name string) bool {
	for _, d := range []string{"access_by_lua_block"} {
		if name == d {
			return true
		}
	}
	return false
}

// XXX this function currently does not preserve inline comments.
func (t *Tree) parseDirective(ctx *context, validate bool) Node {
	item := t.next()
	dirName := item.val

	masks, ok := dirMask[dirName]
	if validate && !ok {
		t.errorf("invalid directive: %q (%d/%d)", item.val, item.pos, item.line)
	}

	n := t.newDirective(item.pos, item.val)

Loop:
	for {
		p := t.peek()
		switch p.typ {
		case itemWord, itemString:
			item := t.next()
			arg := t.newArgument(item.pos, item.val)
			n.append(arg)
		case itemLeftBlock:
			var block Node
			if isFreeFormDirective(dirName) {
				// parse "map" and "split_clients" as normal blocks, but don't validate directives;
				// this allows for parsing of map entries as normal directives.
				block = t.parseBlock(ctx, false)
			} else if isLuaCodeBlockDirective(dirName) {
				// TODO: should call a function that just bypass a Lua code block altogether,
				// or at least it just copy it verbatim
			} else {
				block = t.parseBlock(ctx, validate)
			}
			n.append(block)
			// A block always terminate a directive!
			break Loop
		case itemTerminator:
			t.next()
			break Loop
		case itemNewline:
			// XXX newlines found while scanning a directive should be safe to ignore.
			t.next()
		default:
			t.errorf("unterminated directive: found %s", p.typ)
		}
	}

	// now that we have built the full node we can perform some validation,
	// like checking the number of expected arguments and the presence of blocks.
	if !validate {
		return n
	}

	var args int
	var hasBlock bool
	for _, a := range n.Args {
		switch a.(type) {
		case *ArgumentNode:
			args++
		case *BlockNode:
			hasBlock = true
		}
	}

	for _, mask := range masks {
		if mask&NGX_CONF_TAKE1 == 1 && args > 1 {
			t.errorf("invalid number of arguments for %q", n)
		}

		if mask&NGX_CONF_BLOCK == 1 && !hasBlock {
			t.errorf("directive %q expects a block", n)
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

// this is awfully similar to the global parse() method...
func (t *Tree) parseBlock(ctx *context, validate bool) Node {
	blockStart := t.next()
	block := t.newBlock(blockStart.pos)

Loop:
	for {
		n := t.peek()
		switch n.typ {
		case itemWord:
			block.append(t.parseDirective(ctx, validate))
		case itemNewline:
			if node := t.parseEmptyLines(); node != nil {
				block.append(node)
			}
		case itemComment:
			item := t.next()
			block.append(t.newComment(item.pos, item.val))
		case itemRightBlock:
			t.next()
			break Loop
		default:
			t.errorf("unterminated block: %s", t.peek().typ)
		}
	}

	return block
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
