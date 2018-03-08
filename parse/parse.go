package parse

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Tree is the representation of a single parsed template.
type Tree struct {
	Name      string // name of the template represented by the tree.
	ParseName string // name of the top-level template during parsing, for error messages.
	Root      Node   // top-level root of the tree.
	text      string // text parsed to create the template (or its parent)
	// Parsing only; cleared after parse.
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
	// immediate data structure
}

// Copy returns a copy of the Tree. Any parsing state is discarded.
func (t *Tree) Copy() *Tree {
	if t == nil {
		return nil
	}
	return &Tree{
		Name:      t.Name,
		ParseName: t.ParseName,
		Root:      t.Root.Copy(),
		text:      t.text,
	}
}

// Parse returns a map from template name to parse.Tree, created by parsing the
// templates described in the argument string. The top-level template will be
// given the specified name. If an error is encountered, parsing stops and an
// empty map is returned with the error.
func Parse(name, text string) (tree *Tree, err error) {
	t := New(name)
	t.text = text
	tree, err = t.Parse(text)
	return
}

// Parse from a file path
func ParseFile(path string) (*Tree, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("Failed to read config file")
	}
	tree, err := ParseBytes(bytes)
	return tree, err
}

// Parse from a byte slice
func ParseBytes(configFileData []byte) (tree *Tree, err error) {
	tree, err = Parse("config", string(configFileData))
	return
}

func (t *Tree) GetConfig() *Config {
	return &Config{root: t.Root}
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 item) { // Reverse order: we're pushing back. 0 is newest, 1 is old, 2 is older
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// nextNonSpaceIgnoreNewline returns the next non-space and non-newline token.
func (t *Tree) nextNonSpaceIgnoreNewline() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace && token.typ != itemNewLine {
			break
		}
	}
	return
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace {
			break
		}
	}
	return
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace || token.typ != itemNewLine {
			break
		}
	}
	t.backup()
	return token
}

// Parsing.

// New allocates a new parse tree with the given name.
func New(name string) *Tree {
	return &Tree{
		Name: name,
	}
}

// ErrorContext returns a textual representation of the location of the node in the input text.
// The receiver is only used when the node does not have a pointer to the tree inside,
// which can occur in old code.
func (t *Tree) ErrorContext(n Node) (location, context string) {
	pos := int(n.Position())
	tree := n.tree()
	if tree == nil {
		tree = t
	}
	text := tree.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = n.String()
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	return fmt.Sprintf("%s:%d:%d", tree.ParseName, lineNum, byteNum), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("template: %s:%d: %s", t.ParseName, t.lex.lineNumber(), format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (t *Tree) error(err error) {
	t.errorf("%s", err)
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected itemType, context string) item {
	token := t.nextNonSpaceIgnoreNewline()
	if token.typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (t *Tree) expectOneOf(expected1, expected2 itemType, context string) item {
	token := t.nextNonSpaceIgnoreNewline()
	if token.typ != expected1 && token.typ != expected2 {
		t.unexpected(token, context)
	}
	return token
}

// expected complains about the token and terminates processing.
func (t *Tree) expected(token item, expectToken string) {
	t.errorf("expected %s but token %s shows up", expectToken, token)
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token item, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
}

// startParse initializes the parser, using the lexer.
func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
}

// Parse parses the template definition string to construct a representation of
// the template for execution. If either action delimiter string is empty, the
// default ("{{" or "}}") is used. Embedded template definitions are added to
// the treeSet map.
func (t *Tree) Parse(text string) (tree *Tree, err error) {
	defer t.recover(&err)
	t.ParseName = t.Name
	t.startParse(lex(t.Name, text))
	t.text = text
	t.Root = t.parse()
	t.stopParse()
	return t, nil
}

// parse is the top-level parser for a template, essentially the same
// as itemList except it also parses {{define}} actions.
// It runs to EOF.
func (t *Tree) parse() (result Node) {
	switch token := t.nextNonSpaceIgnoreNewline(); token.typ {
	case itemOpenCurly, itemOpenSquare:
		result = t.parseValue(token)
	default:
		t.backup()
		result = t.parseObject(false)
	}
	t.expect(itemEOF, "EOF")
	return
}

func (t *Tree) parseValue(token item) Node {
	var v Node

	switch token.typ {
	case itemHardSubstitution:
		key := token.val[2 : len(token.val)-1]
		v = t.newField(token.pos, key, true)
	case itemSoftSubstitution:
		key := token.val[3 : len(token.val)-1]
		v = t.newField(token.pos, key, false)
	case itemBool:
		if boolValue, e := strconv.ParseBool(token.val); e != nil {
			if token.val == "on" {
				v = t.newBool(token.pos, true)
			} else if token.val == "off" {
				v = t.newBool(token.pos, false)
			} else {
				panic(e)
			}
		} else {
			v = t.newBool(token.pos, boolValue)
		}
	case itemNull:
		v = t.newNil(token.pos)
	case itemNumber:
		var e error
		v, e = t.newNumber(token.pos, token.val, itemNumber)
		if e != nil {
			panic(e)
		}
	case itemString:
		v = t.newString(token.pos, token.val, unquoteString(token.val))
	case itemUnquotedText:
		v = t.newString(token.pos, token.val, token.val)
	case itemOpenCurly:
		v = t.parseObject(true)
	case itemOpenSquare:
		v = t.parseArray()
	default:
		t.unexpected(token, "parse value")
	}

	return v
}

func (t *Tree) parseObject(hadOpenCurly bool) *MapNode {
	// invoked just after the OPEN_CURLY (or START, if !hadOpenCurly)
	result := t.newMap(t.peekNonSpace().pos)
Loop:
	for {
		switch token := t.nextNonSpaceIgnoreNewline(); {
		case token.typ == itemCloseCurly:
			if !hadOpenCurly {
				t.unexpected(token, "}")
			}
			break Loop
		case token.typ == itemEOF && !hadOpenCurly:
			t.backup()
			break Loop
		default:
			// parse key
			p := t.parseKey(token)
			// parse '=' or '{'
			afterKey := t.nextNonSpaceIgnoreNewline()
			var valueToken item
			if afterKey.typ == itemOpenCurly {
				valueToken = afterKey
			} else {
				if !isKeyValueSeparatorToken(afterKey) {
					t.unexpected(afterKey, "= object")
				}
				t.consolidateValueTokens()
				valueToken = t.nextNonSpaceIgnoreNewline()
			}

			sepIndex := strings.Index(p, ".")
			var key, remaining string
			if sepIndex == -1 {
				key, remaining = p, ""
			} else {
				key, remaining = string(p[:sepIndex]), string(p[sepIndex+1:])
			}

			newValue := t.parseValue(valueToken)

			if sepIndex == -1 {
				if existing, ok := result.Nodes[key]; ok {
					if newValue != nil {
						newValue = newValue.withFallback(existing)
						result.Nodes[key] = newValue
					}
					// TODO - do right (array merging etc), absorb for now
				} else {
					result.Nodes[key] = newValue
				}
			} else {
				obj := t.createValueUnderPath(remaining, newValue)
				if existing, ok := result.Nodes[key]; ok {
					obj = obj.withFallback(existing)
				}
				result.Nodes[key] = obj
			}

			if !t.checkElementSeparator() {
				nextToken := t.nextNonSpaceIgnoreNewline()
				if nextToken.typ == itemCloseCurly {
					if !hadOpenCurly {
						t.unexpected(nextToken, "unbalanced close brace")
					}
					break Loop
				} else if hadOpenCurly {
					t.expected(nextToken, "}")
				} else {
					if nextToken.typ == itemEOF {
						t.backup()
						break
					} else {
						t.expected(nextToken, "EOF")
					}
				}
			}
		}
	}

	return result
}

func (t *Tree) parseArray() *ListNode {
	// invoked just after the OPEN_SQUARE
	result := t.newList(t.peekNonSpace().pos)
	switch token := t.nextNonSpaceIgnoreNewline(); {
	//TODO - do right, absorb for now
	case token.typ == itemCloseSquare:
		return result
	case isValue(token) || token.typ == itemOpenCurly || token.typ == itemOpenSquare || token.typ == itemSoftSubstitution || token.typ == itemHardSubstitution:
		v := t.parseValue(token)
		result.append(v)
	default:
		t.unexpected(token, "ListNode")

	}

	for {
		var token item
		if !t.checkElementSeparator() {
			token = t.nextNonSpaceIgnoreNewline()
			if token.typ == itemCloseSquare {
				break
			}
		}
		nextToken := t.peek()
		if nextToken.typ != itemOpenCurly && nextToken.typ != itemOpenSquare {
			t.consolidateValueTokens()
		}

		token = t.nextNonSpaceIgnoreNewline()
		if isValue(token) || token.typ == itemOpenCurly || token.typ == itemOpenSquare {
			v := t.parseValue(token)
			result.append(v)
		} else if token.typ == itemCloseSquare {
			// we allow one trailing comma
			t.backup()
		} else {
			t.unexpected(token, "ListNode")
		}
	}
	return result
}

func (t *Tree) parseKey(token item) string {
	return token.val
}

func (t *Tree) consolidateValueTokens() {
	var tokens []item
	token := t.nextNonSpaceIgnoreNewline()
	for {
		if isValue(token) || token.typ == itemUnquotedText {
			tokens = append(tokens, token)
		} else {
			break
		}
		token = t.nextNonSpace()
	}

	if tokens == nil {
		t.backup()
		return
	} else {
		t.backup2(consolidate(tokens))
	}
}

func consolidate(tokens []item) item {
	if len(tokens) == 1 {
		return tokens[0]
	} else {
		consolidatedToken := tokens[0].val
		for i := 1; i < len(tokens); i++ {
			consolidatedToken += " " + tokens[i].val
		}
		return item{itemString, tokens[0].pos, consolidatedToken}
	}
}

func isKeyValueSeparatorToken(token item) bool {
	return token.typ == itemColon || token.typ == itemEquals || token.typ == itemPlusEquals
}

func isValue(token item) bool {
	switch token.typ {
	case itemBool, itemNull, itemNumber, itemString:
		return true
	}
	return false
}

func unquoteString(value string) string {
	re := regexp.MustCompile("^\"(.*)\"$")
	if strippedVal := re.FindStringSubmatch(value); strippedVal != nil {
		return strippedVal[1]
	} else {
		return value
	}
}

func (t *Tree) checkElementSeparator() bool {
	token := t.next()
	sawSeparatorOrNewline := false
	for {
		if token.typ == itemNewLine {
			sawSeparatorOrNewline = true
		} else if token.typ == itemComma {
			return true
		} else {
			t.backup()
			return sawSeparatorOrNewline
		}
		token = t.next()
	}
}

func (t *Tree) createValueUnderPath(remaining string, newValue Node) Node {
	ps := strings.Split(remaining, ".")
	prevObj := newValue
	for i := len(ps) - 1; i >= 0; i-- {
		obj := t.newMap(newValue.Position())
		key := ps[i]
		obj.Nodes[key] = prevObj
		prevObj = obj
	}
	return prevObj
}
