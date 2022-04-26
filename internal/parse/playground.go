package parse

import (
	"fmt"
)

func printDirective(node *DirectiveNode) {
	fmt.Printf("%s ", node)
	mustTerminate := false
	for _, x := range node.Args {
		switch arg := x.(type) {
		case *ArgumentNode:
			fmt.Printf("%s ", arg)
			mustTerminate = true
		case *BlockNode:
			fmt.Printf("{\n")
			printList(arg.List)
			fmt.Printf("}\n")
			mustTerminate = false
		}
	}
	if mustTerminate {
		fmt.Printf(";\n")
	}
}

func printList(node *ListNode) {
	for _, x := range node.Nodes {
		switch sub := x.(type) {
		case *DirectiveNode:
			printDirective(sub)
		case *CommentNode:
			fmt.Println(sub)
		case *EmptyLineNode:
			fmt.Println()
		default:
			panic("dunno")
		}
	}
}

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

	printList(t.Root)
}
