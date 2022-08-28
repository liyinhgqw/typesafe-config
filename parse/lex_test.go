package parse

import (
	"fmt"
	"testing"
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError: "error",
	itemBool:  "bool",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
}

var (
	tEOF     = item{itemEOF, 0, ""}
	tNewLine = item{itemNewLine, 0, "\n"}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t", []item{{itemSpace, 0, " \t"}, tEOF}},
	{"newline", " \n", []item{{itemSpace, 0, " "}, tNewLine, tEOF}},
	{"comment", "/* abc */", []item{tEOF}},
	{"double slash comment", "// abc", []item{tEOF}},
	{"double slash comment", "# abc", []item{tEOF}},
	{"quote", `/* abc */"def"/* gh */`, []item{{itemString, 0, `"def"`}, tEOF}},
	{"raw quote", "/* abc */`def`/* gh */", []item{{itemString, 0, "`def`"}, tEOF}},
	{"comma", "a,b", []item{{itemUnquotedText, 0, "a"}, {itemComma, 0, ","}, {itemUnquotedText, 0, "b"}, tEOF}},
	{"colon", "a:b", []item{{itemUnquotedText, 0, "a"}, {itemColon, 0, ":"}, {itemUnquotedText, 0, "b"}, tEOF}},
	{"equal", "a=b", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemUnquotedText, 0, "b"}, tEOF}},
	{"curly", "{a=b}", []item{{itemOpenCurly, 0, "{"}, {itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemUnquotedText, 0, "b"}, {itemCloseCurly, 0, "}"}, tEOF}},
	{"square", "[a,b]", []item{{itemOpenSquare, 0, "["}, {itemUnquotedText, 0, "a"}, {itemComma, 0, ","}, {itemUnquotedText, 0, "b"}, {itemCloseSquare, 0, "]"}, tEOF}},
	{"plus equal", "a+=b", []item{{itemUnquotedText, 0, "a"}, {itemPlusEquals, 0, "+="}, {itemUnquotedText, 0, "b"}, tEOF}},
	{"number", "a=-1.2", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemNumber, 0, "-1.2"}, tEOF}},
	{"hard substitution", "a=${b}", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemHardSubstitution, 0, "${b}"}, tEOF}},
	{"soft substitution", "a=${?b}", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemSoftSubstitution, 0, "${?b}"}, tEOF}},
	{"unquote", "a=-1.2 min", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemNumber, 0, "-1.2"}, {itemSpace, 0, " "}, {itemUnquotedText, 0, "min"}, tEOF}},
	{"true", "a=true", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemBool, 0, "true"}, tEOF}},
	{"nil", "a=nil", []item{{itemUnquotedText, 0, "a"}, {itemEquals, 0, "="}, {itemNull, 0, "nil"}, tEOF}},
}

// collect gathers the emitted items into a slice.
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

func equal(i1, i2 []item, checkPos bool, t *testing.T) bool {
	if len(i1) != len(i2) {
		t.Errorf("len: %v != %v", i1, i2)
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			t.Errorf("type: %v != %v", i1[k].typ, i2[k].typ)
			return false
		}
		if i1[k].val != i2[k].val {
			t.Errorf("val: %v != %v", i1[k].val, i2[k].val)
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !equal(items, test.items, false, t) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}
