package parse

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemSpace
	itemComma
	itemEquals
	itemColon
	itemOpenCurly
	itemCloseCurly
	itemOpenSquare
	itemCloseSquare
	itemNewLine
	itemUnquotedText
	itemSubstitution
	itemComment
	itemPlusEquals
	itemString
	itemBool
	itemNumber
	itemComplex
	itemNull
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// reset steps back one token. Can only be called once per call of next.
func (l *lexer) reset() {
	l.pos = l.start
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexNextToken; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions

const (
	leftComment        = "/*"
	rightComment       = "*/"
	doubleSlashComment = "//"
)

// lexNextToken scans the elements.
func lexNextToken(l *lexer) stateFn {
	// Either number, quoted string, or identifier.
	// Spaces separate arguments; runs of spaces turn into itemSpace.
	// Pipe symbols separate and are emitted.
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
	case isEndOfLine(r):
		l.emit(itemNewLine)
	case isSpace(r):
		return lexSpace
	case r == '/':
		rr := l.next()
		if rr == '/' {
			return lexDoubleSlashComment
		} else if rr == '*' {
			return lexComment
		} else {
			return l.errorf("expected // or /*")
		}
	case r == '#':
		return lexDoubleSlashComment
	case r == '"':
		return lexQuote
	case r == '`':
		return lexRawQuote
	case r == ':':
		l.emit(itemColon)
	case r == ',':
		l.emit(itemComma)
	case r == '=':
		l.emit(itemEquals)
	case r == '{':
		l.emit(itemOpenCurly)
	case r == '}':
		l.emit(itemCloseCurly)
	case r == '[':
		l.emit(itemOpenSquare)
	case r == ']':
		l.emit(itemCloseSquare)
	case r == '$':
		return lexSubstitution
	case r == '+':
		return lexPlusEquals
	case r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexUnquotedText
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexNextToken
}

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(rightComment))
	l.ignore()
	return lexNextToken
}

func lexDoubleSlashComment(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof || isEndOfLine(r) {
			l.backup()
			l.ignore()
			break
		}
	}
	return lexNextToken
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(itemString)
	return lexNextToken
}

// lexRawQuote scans a raw quoted string.
func lexRawQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof, '\n':
			return l.errorf("unterminated raw quoted string")
		case '`':
			break Loop
		}
	}
	l.emit(itemString)
	return lexNextToken
}

func lexIgnoreIfEmptySubstitution(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r), r == '.', r == '_', r == '-':
		case r == '}':
			envName := l.input[l.start+3 : l.pos-1]
			setEnvValue(l, envName, false)
			break Loop
		// absorb.
		default:
			return l.errorf("variable substitution can only include letters, numbers, dot, dash or underscore.")
		}
	}
	return lexNextToken
}

func lexNormalSubstitution(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r), r == '.', r == '_', r == '-':
		case r == '}':
			envName := l.input[l.start+2 : l.pos-1]
			setEnvValue(l, envName, true)
			break Loop
		// absorb.
		default:
			return l.errorf("variable substitution can only include letters, numbers, dot, dash or underscore.")
		}
	}
	return lexNextToken
}

func setEnvValue(l *lexer, envName string, setNil bool) {

	if envVal, found := os.LookupEnv(envName); found {
		if strings.ContainsAny(envVal, ":") {
			envVal = `"` + envVal + `"`
		}
		// replace the ${...} with just the value from the env and reset so that it can be
		// parsed as whatever value it is
		l.input = l.input[:l.start] + envVal + l.input[l.pos:]
		l.reset()
	} else {
		// set it to nil value
		if setNil {
			l.input = l.input[:l.start] + "nil" + l.input[l.pos:]
			l.reset()
		} else {
			l.emit(itemSubstitution)
		}
	}

}

func lexSubstitution(l *lexer) stateFn {

	if l.next() == '{' {
		if l.peek() == '?' {
			l.next()
			return lexIgnoreIfEmptySubstitution(l)
		} else {
			return lexNormalSubstitution(l)

		}
	}
	return lexNextToken
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return lexNextToken
}

// lexUnquotedText scans an alphanumeric.
func lexUnquotedText(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r), r == '.':
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			switch {
			case word == "true", word == "false", word == "on", word == "off":
				l.emit(itemBool)
			case word == "nil":
				l.emit(itemNull)
			default:
				l.emit(itemUnquotedText)
			}
			break Loop
		}
	}
	return lexNextToken
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	if sign := l.peek(); sign == '+' || sign == '-' {
		// Complex: 1+2i. No spaces, must end in 'i'.
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}
		l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}
	return lexNextToken
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// Is it imaginary?
	l.accept("i")

	return true
}

// lexPlusEquals scans +=
func lexPlusEquals(l *lexer) stateFn {
	if r := l.next(); r == '=' {
		l.emit(itemPlusEquals)
	} else {
		return l.errorf("expected +=")
	}
	return lexNextToken
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
