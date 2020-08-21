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
	var items []item
	for t := range l.items {
		items = append(items, t)
	}

	for i, t := range items {
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
			n := i + 1
			var skipIndent bool

			if n < len(items) {
				if items[n].typ == itemCloseBlock {
					skipIndent = true
				}
			}
			if depth > 0 && startOfLine {
				var ilvl int
				if skipIndent {
					ilvl = (depth - 1) * indent
				} else {
					ilvl = depth * indent
				}
				fmt.Print(strings.Repeat(" ", ilvl))
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
