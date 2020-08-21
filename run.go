package nginxp

import "fmt"

// Stuff is a testing function
func Stuff(input string) {
	l := lex("test", input)
	for t := range l.items {
		fmt.Println(t)
	}
}
