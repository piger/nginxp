package parse

import "fmt"

// NodeType identifies the type of a parse tree node.
type NodeType int

func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText      NodeType = iota
	NodeComment            // A comment.
	NodeList               // A list of Nodes
	NodeDirective          // An nginx configuration directive
	NodeBlock              // A configuration block
)

// Pos represents a byte position in the original input text from which this
// file was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// A Node is an element in the parse tree.
type Node interface {
	Type() NodeType
	String() string
	Copy() Node
	Position() Pos
	tree() *Tree // unexported so that only local types can satisfy it.
}

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	tr    *Tree
	Nodes []Node
}

func (t *Tree) newList(pos Pos) *ListNode {
	return &ListNode{tr: t, NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) tree() *Tree {
	return l.tr
}

// XXX
func (l *ListNode) String() string {
	return "XXX"
}

func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := l.tr.newList(l.Pos)
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *ListNode) Copy() Node {
	return l.CopyList()
}

// CommentNode holds a comment.
type CommentNode struct {
	NodeType
	Pos
	tr   *Tree
	Text string
}

func (t *Tree) newComment(pos Pos, text string) *CommentNode {
	return &CommentNode{tr: t, NodeType: NodeComment, Pos: pos, Text: text}
}

func (c *CommentNode) String() string {
	return fmt.Sprintf("#%s", c.Text)
}

func (c *CommentNode) tree() *Tree {
	return c.tr
}

func (c *CommentNode) Copy() Node {
	return &CommentNode{tr: c.tr, NodeType: NodeComment, Pos: c.Pos, Text: c.Text}
}

// DirectiveNode contains a directive and is linked to its arguments, including an optional block.
type DirectiveNode struct {
	NodeType
	Pos
	tr   *Tree
	Text string
	Args []Node // Arguments, which can include a "Block"
}

func (t *Tree) newDirective(pos Pos, text string) *DirectiveNode {
	return &DirectiveNode{tr: t, NodeType: NodeDirective, Pos: pos, Text: text}
}
