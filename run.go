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

	// the list of longest starting word per block
	var longest []int
	// the list of longest word
	var lt int
	startOfLine = true
	for _, t := range items {
		switch {
		case startOfLine && t.typ == itemStatement:
			ll := len(t.val)
			if ll > lt {
				lt = ll
			}
			startOfLine = false
		case t.typ == itemNewLine:
			c := strings.Count(t.val, "\n")
			if c >= 2 {
				longest = append(longest, lt)
				lt = 0
			}

			startOfLine = true
		}
	}
	longest = append(longest, lt)

	fmt.Printf("longest = %v", longest)

	startOfLine = false
	var firstOfLine bool
	var longestIdx int

	for i, t := range items {
		switch {
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
				switch {
				case items[n].typ == itemCloseBlock:
					skipIndent = true
				case items[n].typ == itemOpenBlock:
					fmt.Print(" ")
					continue
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
			startOfLine = true
			c := strings.Count(t.val, "\n")
			if c > 2 {
				c = 2
			}
			// fmt.Printf("c = %d\n", c)
			fmt.Print(strings.Repeat("\n", c))
			firstOfLine = true
			// if next is another "\n", increment longestIdx
			if c >= 2 {
				longestIdx++
			}
		default:
			switch {
			case t.val == "location" || t.val == "upstream" || t.val == "server":
				fmt.Print(t.val)
			case firstOfLine:
				firstOfLine = false
				sfmt := fmt.Sprintf("%%-%ds", longest[longestIdx])
				fmt.Printf(sfmt, t.val)
			default:
				fmt.Print(t.val)
			}
		}
	}
}
