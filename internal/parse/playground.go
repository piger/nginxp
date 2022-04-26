package parse

import "fmt"

// LexerPlayground is a "playground" function that showcase the lexer.
func LexerPlayground(filename, contents string) {
	lex := lex(filename, contents)
	for token := range lex.items {
		fmt.Printf("%+v (%s)\n", token, token.typ)
	}
}
