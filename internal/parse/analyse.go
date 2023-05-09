package parse

import (
	"fmt"
)

type Directive struct {
	Name  string   `json:"name"`
	Args  []string `json:"args"`
	Block *Block   `json:"block"`
}

type Block struct {
	Directives []*Directive `json:"directives"`
}

type Configuration struct {
	Directives []*Directive `json:"directives"`
}

func (c *Configuration) Analyse(node *ListNode) error {
	for _, nodeRaw := range node.Nodes {
		switch node := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := c.iterateDirective(node.String(), node)
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

func (c *Configuration) iterateDirective(name string, node *DirectiveNode) (*Directive, error) {
	d := &Directive{Name: name}
	for _, argRaw := range node.Args {
		switch arg := argRaw.(type) {
		case *ArgumentNode:
			d.Args = append(d.Args, arg.String())
		case *BlockNode:
			b, err := c.iterateBlock(arg)
			if err != nil {
				return nil, err
			}
			d.Block = b
		}
	}
	return d, nil
}

func (c *Configuration) iterateBlock(node *BlockNode) (*Block, error) {
	b := &Block{}
	for _, nodeRaw := range node.List.Nodes {
		switch subNode := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := c.iterateDirective(subNode.String(), subNode)
			if err != nil {
				return nil, err
			}
			b.Directives = append(b.Directives, d)
		case *CommentNode, *EmptyLineNode:
			continue
		default:
			panic(fmt.Sprintf("unexpected node type in block: %s\n", subNode.Type()))
		}
	}
	return b, nil
}

func printOneDirective(directive *Directive) {
	fmt.Printf("directive: Name=%s, Args=%q, Block? (%v)\n", directive.Name, directive.Args, directive.Block != nil)
	if directive.Block != nil {
		for _, d := range directive.Block.Directives {
			printOneDirective(d)
		}
	}
}

func Analyse(filename, contents string) error {
	t, err := Parse(filename, contents)
	if err != nil {
		return err
	}

	cfg := Configuration{}
	if err := cfg.Analyse(t.Root); err != nil {
		return err
	}

	for _, directive := range cfg.Directives {
		printOneDirective(directive)
	}

	return nil
}
