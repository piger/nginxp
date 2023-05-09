package parse

import "fmt"

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation for
// embedding in a Node.
func (t NodeType) Type() NodeType {
	return t
}

func (t NodeType) String() string {
	return nodeTypes[t]
}

const (
	NodeText      NodeType = iota
	NodeComment            // A comment.
	NodeList               // A list of Nodes
	NodeDirective          // An nginx configuration directive
	NodeBlock              // A configuration block
	NodeArgument           // An argument for a directive
	NodeEmptyLine          // An empty line; will be used for formatting
)

var nodeTypes = map[NodeType]string{
	NodeText:      "NodeText",
	NodeComment:   "NodeComment",
	NodeList:      "NodeList",
	NodeDirective: "NodeDirective",
	NodeBlock:     "NodeBlock",
	NodeArgument:  "NodeArgument",
	NodeEmptyLine: "NodeEmptyLine",
}

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

func (l *ListNode) String() string {
	return ""
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

func (d *DirectiveNode) String() string {
	return d.Text
}

func (d *DirectiveNode) tree() *Tree {
	return d.tr
}

func (d *DirectiveNode) Copy() Node {
	n := &DirectiveNode{tr: d.tr, NodeType: NodeDirective, Pos: d.Pos, Text: d.Text}
	for _, arg := range d.Args {
		n.Args = append(n.Args, arg.Copy())
	}
	return n
}

func (d *DirectiveNode) append(arg Node) {
	d.Args = append(d.Args, arg)
}

// ArgumentNode contains one argument (string) for a directive.
type ArgumentNode struct {
	NodeType
	Pos
	tr   *Tree
	Text string
}

func (t *Tree) newArgument(pos Pos, text string) *ArgumentNode {
	return &ArgumentNode{tr: t, NodeType: NodeArgument, Pos: pos, Text: text}
}

func (a *ArgumentNode) String() string {
	return a.Text
}

func (a *ArgumentNode) tree() *Tree {
	return a.tr
}

func (a *ArgumentNode) Copy() Node {
	return &ArgumentNode{tr: a.tr, NodeType: NodeArgument, Pos: a.Pos, Text: a.Text}
}

type EmptyLineNode struct {
	NodeType
	Pos
	tr *Tree
}

func (t *Tree) newEmptyLine(pos Pos) *EmptyLineNode {
	return &EmptyLineNode{tr: t, NodeType: NodeEmptyLine, Pos: pos}
}

func (e *EmptyLineNode) String() string {
	return "\n"
}

func (e *EmptyLineNode) tree() *Tree {
	return e.tr
}

func (e *EmptyLineNode) Copy() Node {
	return &EmptyLineNode{tr: e.tr, NodeType: NodeEmptyLine, Pos: e.Pos}
}

type BlockNode struct {
	NodeType
	Pos
	tr   *Tree
	List *ListNode // The list of nodes in this block
}

func (t *Tree) newBlock(pos Pos) *BlockNode {
	l := t.newList(pos)
	return &BlockNode{tr: t, NodeType: NodeBlock, Pos: pos, List: l}
}

func (b *BlockNode) String() string {
	return ""
}

func (b *BlockNode) tree() *Tree {
	return b.tr
}

func (b *BlockNode) Copy() Node {
	n := &BlockNode{tr: b.tr, NodeType: NodeBlock, Pos: b.Pos, List: b.List.CopyList()}
	return n
}

func (b *BlockNode) append(node Node) {
	b.List.append(node)
}
