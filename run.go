package nginxp

import (
	"fmt"
	"strings"
)

const indent = 4

// Stuff is a testing function
func Stuff(input string) {
	var depth int
	var startOfLine bool

	l := lex("test", input)
	for t := range l.items {
		switch {
		case t.typ == itemNewLine:
			startOfLine = true
			fmt.Println()
		case t.typ == itemOpenBlock:
			depth++
			fmt.Print("{")
		case t.typ == itemCloseBlock:
			depth--
			fmt.Print("}")
		case t.typ == itemSpace:
			if depth > 0 && startOfLine {
				fmt.Print(strings.Repeat(" ", depth*indent))
				startOfLine = false
			} else {
				fmt.Print(" ")
			}
		case t.typ == itemNewLine:
			fmt.Println()
		default:
			fmt.Print(t.val)
		}
	}
}
