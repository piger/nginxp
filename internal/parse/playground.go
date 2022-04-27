package parse

import (
	"fmt"
	"strings"
)

func printDirective(node *DirectiveNode, indent int) {
	if indent > 0 {
		fmt.Print(strings.Repeat(" ", indent))
	}

	fmt.Printf("%s", node)
	mustTerminate := true

	for _, x := range node.Args {
		switch arg := x.(type) {
		case *ArgumentNode:
			fmt.Printf(" %s", arg)
		case *BlockNode:
			fmt.Printf(" {\n")
			printList(arg.List, indent+4)
			if indent > 0 {
				fmt.Print(strings.Repeat(" ", indent))
			}
			fmt.Printf("}\n")
			mustTerminate = false
		case *FreeformBlockNode:
			fmt.Printf(" {\n")
			printList(arg.List, indent+4)
			if indent > 0 {
				fmt.Print(strings.Repeat(" ", indent))
			}
			fmt.Printf("}\n")
			mustTerminate = false
		}
	}
	if mustTerminate {
		fmt.Printf(";\n")
	}
}

func printList(node *ListNode, indent int) {
	for _, x := range node.Nodes {
		switch sub := x.(type) {
		case *DirectiveNode:
			printDirective(sub, indent)
		case *CommentNode:
			if indent > 0 {
				fmt.Print(strings.Repeat(" ", indent))
			}
			fmt.Println(sub)
		case *EmptyLineNode:
			fmt.Println()
		case *ArgumentNode:
			if indent > 0 {
				fmt.Print(strings.Repeat(" ", indent))
			}
			fmt.Print(sub)
		case *ListNode:
			printList(sub, indent)
			fmt.Println(";")
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

	printList(t.Root, 0)
}
