package parse

import (
	"fmt"
)

// LexerPlayground is a "playground" function that showcase the lexer.
func LexerPlayground(filename, contents string) {
	lex := lex(filename, contents)
	for token := range lex.items {
		if token.val == "\n" {
			fmt.Println()
		} else {
			fmt.Printf("%s (%s)", token.val, token.typ)
		}
	}

	fmt.Printf("\n\nPARSER TREE TEST ===>\n")

	t, err := Parse(filename, contents)
	if err != nil {
		panic(err)
	}

	for _, node := range t.Root.Nodes {
		switch n := node.(type) {
		case *DirectiveNode:
			fmt.Printf("%s ", n)
			for _, arg := range n.Args {
				fmt.Printf("%s ", arg)
			}
			fmt.Println()
		case *EmptyLineNode:
			fmt.Println()
		case *CommentNode:
			fmt.Println(n)
		}

	}
}
