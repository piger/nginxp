package parse

import (
	"fmt"
)

type Directive struct {
	Name  string       `json:"name"`
	Args  []string     `json:"args"`
	Block []*Directive `json:"block,omitempty"`
}

type Configuration struct {
	Filename   string       `json:"filename"`
	Directives []*Directive `json:"directives"`
}

func (c *Configuration) ParseTree(tree *Tree) error {
	for _, nodeRaw := range tree.Root.Nodes {
		switch node := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := c.iterateDirective(node)
			if err != nil {
				return err
			}
			c.Directives = append(c.Directives, d)
		case *CommentNode, *EmptyLineNode:
			continue
		default:
			panic(fmt.Sprintf("Unhandled node type: %s\n", node.Type()))
		}
	}
	return nil
}

func (c *Configuration) iterateDirective(node *DirectiveNode) (*Directive, error) {
	d := &Directive{Name: node.String(), Args: []string{}}

	for _, argRaw := range node.Args {
		switch arg := argRaw.(type) {
		case *ArgumentNode:
			d.Args = append(d.Args, arg.String())
		case *BlockNode:
			dirs, err := c.iterateBlock(arg)
			if err != nil {
				return nil, err
			}
			d.Block = dirs
		}
	}
	return d, nil
}

func (c *Configuration) iterateBlock(node *BlockNode) ([]*Directive, error) {
	var dirs []*Directive

	for _, nodeRaw := range node.List.Nodes {
		switch subNode := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := c.iterateDirective(subNode)
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

func printOneDirective(directive *Directive) {
	fmt.Printf("directive: Name=%s, Args=%q, Block? (%v)\n", directive.Name, directive.Args, len(directive.Block) > 0)
	if directive.Block != nil {
		for _, d := range directive.Block {
			printOneDirective(d)
		}
	}
}

func Analyse(filename, contents string) error {
	t, err := Parse(filename, contents)
	if err != nil {
		return err
	}

	cfg := Configuration{Filename: filename}
	if err := cfg.ParseTree(t); err != nil {
		return err
	}

	for _, directive := range cfg.Directives {
		printOneDirective(directive)
	}

	return nil
}
