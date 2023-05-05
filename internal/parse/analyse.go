package parse

import "fmt"

type Section struct {
	Name       string
	Sections   []*Section
	Directives []*Directive // I'd use a map, but what about directive with a non-unique name like "if"?
	Args       []string     // used only by locations
}

type Directive struct {
	Name  string
	Args  []string
	Block *Block
}

type Block struct {
	Directives []*Directive
}

type Configuration struct {
	Sections []*Section
}

func (c *Configuration) Analyse(section string, node *ListNode) error {
	s := &Section{Name: section}
	c.Sections = append(c.Sections, s)

	// PROBLEM:
	// a "location" is being considered both a section and a directive; the effect of this
	// is that we end up with a location (directive) with no block, and the same location also
	// as a section, with no args.
	for _, nodeRaw := range node.Nodes {
		switch node := nodeRaw.(type) {
		case *DirectiveNode:
			d, err := c.iterateDirective(node.String(), node)
			if err != nil {
				return err
			}
			s.Directives = append(s.Directives, d)
		case *ListNode:
			panic("what now?")
		case *CommentNode:
		case *EmptyLineNode:
		default:
			fmt.Printf("%+v\n", node)
			panic("very unexpected")
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
		case *FreeformBlockNode:
			// panic("not implemented")
			fmt.Println("skipping freeform")
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
		case *CommentNode:
		case *EmptyLineNode:
		default:
			panic("unexpected!")
		}
	}
	return b, nil
}

func printOneDirective(directive *Directive) {
	fmt.Printf("directive: %+v\n", directive)
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
	if err := cfg.Analyse("main", t.Root); err != nil {
		return err
	}

	for _, section := range cfg.Sections {
		fmt.Printf("section: %q\n", section.Name)

		for _, directive := range section.Directives {
			printOneDirective(directive)
		}
	}

	return nil
}
