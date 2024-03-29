package parse

import (
	"testing"
)

type lexTest struct {
	name  string
	input string
	items []item
}

func mkItem(typ itemType, text string) item {
	return item{
		typ: typ,
		val: text,
	}
}

var (
	tEOF         = mkItem(itemEOF, "")
	tQuote       = mkItem(itemString, `"tis a string"`)
	tQuoteMixed  = mkItem(itemString, `"tis 'a string"`)
	tQuoteSingle = mkItem(itemString, `'it\'s a me'`)
	tNewLine     = mkItem(itemNewline, "\n")
	tTerm        = mkItem(itemTerminator, ";")
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"quote", `"tis a string"`, []item{tQuote, tEOF}},
	{"quote mixed", `"tis 'a string"`, []item{tQuoteMixed, tEOF}},
	{"quote single", `'it\'s a me'`, []item{tQuoteSingle, tEOF}},
	{"comment", "# I'm a comment line", []item{
		mkItem(itemComment, " I'm a comment line"),
		tEOF,
	}},
	{"two comments", "# I'm a comment line\n#and more", []item{
		mkItem(itemComment, " I'm a comment line"),
		mkItem(itemNewline, "\n"),
		mkItem(itemComment, "and more"),
		tEOF,
	}},
	{"directive with arguments", `proxy_set_header Host "Foo-Bar";`, []item{
		mkItem(itemWord, "proxy_set_header"),
		mkItem(itemWord, "Host"),
		mkItem(itemString, `"Foo-Bar"`),
		mkItem(itemTerminator, ";"),
		tEOF,
	}},
	{"words ending with comment", `proxy_set_header Host "Foo"; # a comment`, []item{
		mkItem(itemWord, "proxy_set_header"),
		mkItem(itemWord, "Host"),
		mkItem(itemString, `"Foo"`),
		mkItem(itemTerminator, ";"),
		mkItem(itemComment, " a comment"),
		tEOF,
	}},
	{"newline", "\n", []item{tNewLine, tEOF}},
	{"newlines", "\n\n", []item{tNewLine, tNewLine, tEOF}},
	{"word with newlines", "foo;\n\n", []item{
		mkItem(itemWord, "foo"),
		tTerm,
		tNewLine,
		tNewLine,
		tEOF,
	}},
	{"word with variable", "access_by_lua_file code/${something};", []item{
		mkItem(itemWord, "access_by_lua_file"),
		mkItem(itemWord, "code/${something}"),
		tTerm,
		tEOF,
	}},
	// errors
	{"unclosed quoted string", `"I'm unclosed`, []item{
		mkItem(itemError, "unterminated quoted string"),
	}},
}

func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
		if checkPos && i1[k].line != i2[k].line {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, tt := range lexTests {
		t.Run(tt.name, func(t *testing.T) {
			items := collect(&tt)
			if !equal(items, tt.items, false) {
				t.Fatalf("expected:\n\t%+v\ngot:\n\t%+v", tt.items, items)
			}
		})
	}
}
