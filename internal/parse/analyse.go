package parse

import (
	"fmt"
)

// Directive contains a single nginx configuration directive; it has a number of optional
// Args, according to the bitmask in bitmask.go, and an optional Block.
type Directive struct {
	Name  string       `json:"name"`
	Args  []string     `json:"args"`
	Block []*Directive `json:"block,omitempty"`
}

// Configuration contains the nginx configuration from a single configuration file.
type Configuration struct {
	Filename   string       `json:"filename"`
	Directives []*Directive `json:"directives"`
}

// NewConfiguration creates a Configuration from a parsed Tree.
func NewConfiguration(tree *Tree) (*Configuration, error) {
	cfg := &Configuration{
		Filename: tree.Filename,
	}

	for _, nodeRaw := range tree.Root.Nodes {
		switch node := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := iterateDirective(node)
			if err != nil {
				return nil, err
			}
			cfg.Directives = append(cfg.Directives, d)
		case *CommentNode, *EmptyLineNode:
			continue
		default:
			panic(fmt.Sprintf("Unhandled node type: %s\n", node.Type()))
		}
	}
	return cfg, nil
}

func iterateDirective(node *DirectiveNode) (*Directive, error) {
	d := &Directive{Name: node.String(), Args: []string{}}

	for _, argRaw := range node.Args {
		switch arg := argRaw.(type) {
		case *ArgumentNode:
			d.Args = append(d.Args, arg.String())
		case *BlockNode:
			dirs, err := iterateBlock(arg)
			if err != nil {
				return nil, err
			}
			d.Block = dirs
		}
	}
	return d, nil
}

func iterateBlock(node *BlockNode) ([]*Directive, error) {
	var dirs []*Directive

	for _, nodeRaw := range node.List.Nodes {
		switch subNode := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := iterateDirective(subNode)
			if err != nil {
				return nil, err
			}
			dirs = append(dirs, d)
		case *CommentNode, *EmptyLineNode:
			continue
		default:
			panic(fmt.Sprintf("unexpected node type in block: %s\n", subNode.Type()))
		}
	}
	return dirs, nil
}
