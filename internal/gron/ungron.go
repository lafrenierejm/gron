// Ungronning is the reverse of gronning: turn statements
// back into JSON. The expected input grammar is:
//
//   Input ::= '--'* Statement (Statement | '--')*
//   Statement ::= Path Space* "=" Space* Value ";" "\n"
//   Path ::= (BareWord) ("." BareWord | ("[" Key "]"))*
//   Value ::= String | Number | "true" | "false" | "null" | "[]" | "{}"
//   BareWord ::= (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | '$' | '_') (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | UnicodeMn | UnicodeMc | UnicodeNd | UnicodePc | '$' | '_')*
//   Key ::= [0-9]+ | String
//   String ::= '"' (UnescapedRune | ("\" (["\/bfnrt] | ('u' Hex))))* '"'
//   UnescapedRune ::= [^#x0-#x1f"\]

package gron

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	json "github.com/virtuald/go-ordered-json"

	"github.com/pkg/errors"
)

// Ungron is the reverse of gron. Given assignment statements as input,
// it returns JSON. The only option is optMonochrome
func Ungron(r io.Reader, w io.Writer, outJson bool, colorize bool) (int, error) {
	scanner := bufio.NewScanner(r)
	var maker StatementMaker

	// Allow larger internal buffer of the scanner (min: 64KiB ~ max: 1MiB)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	if outJson {
		maker = StatementFromJSONSpec
	} else {
		maker = StatementFromStringMaker
	}

	// Make a list of statements from the input
	var ss Statements
	for scanner.Scan() {
		s, err := maker(scanner.Text())
		if err != nil {
			return exitParseStatements, err
		}
		ss.Add(s)
	}
	if err := scanner.Err(); err != nil {
		return exitReadInput, fmt.Errorf("failed to read input statements")
	}

	// turn the statements into a single merged interface{} type
	merged, err := ss.ToInterface()
	if err != nil {
		return exitParseStatements, err
	}

	// If there's only one top level key and it's "json", make that the top level thing
	switch merged.(type) {

	case json.OrderedObject:
		if mergedOrderObject, ok := merged.(json.OrderedObject); ok {
			if len(mergedOrderObject) == 1 && mergedOrderObject[0].Key == "json" {
				merged = mergedOrderObject[0].Value
			}
		}

	case map[string]interface{}:
		if mergedMap, ok := merged.(map[string]interface{}); ok {
			if len(mergedMap) == 1 {
				if _, exists := mergedMap["json"]; exists {
					merged = mergedMap["json"]
				}
			}
		}
	}

	// Marshal the output into JSON to display to the user
	out := &bytes.Buffer{}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err = enc.Encode(merged)
	if err != nil {
		return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
	}
	j := out.Bytes()

	// If the output isn't monochrome, add color to the JSON
	if colorize {
		c, err := colorizeJSON(j)

		// If we failed to colorize the JSON for whatever reason,
		// we'll just fall back to monochrome output, otherwise
		// replace the monochrome JSON with glorious technicolor
		if err == nil {
			j = c
		}
	}

	// For whatever reason, the monochrome version of the JSON
	// has a trailing newline character, but the colorized version
	// does not. Strip the whitespace so that neither has the newline
	// character on the end, and then we'll add a newline in the
	// Fprintf below
	j = bytes.TrimSpace(j)

	fmt.Fprintf(w, "%s\n", j)

	return exitOK, nil
}

// errRecoverable is an error type to represent errors that
// can be recovered from; e.g. an empty line in the input
type errRecoverable struct {
	msg string
}

func (e errRecoverable) Error() string {
	return e.msg
}

// A lexer holds the state for lexing statements
type lexer struct {
	text       string  // The raw input text
	pos        int     // The current byte offset in the text
	width      int     // The width of the current rune in bytes
	cur        rune    // The rune at the current position
	prev       rune    // The rune at the previous position
	tokens     []Token // The tokens that have been emitted
	tokenStart int     // The starting position of the current token
}

// newLexer returns a new lexer for the provided input string
func newLexer(text string) *lexer {
	return &lexer{
		text:       text,
		pos:        0,
		tokenStart: 0,
		tokens:     make([]Token, 0),
	}
}

// lex runs the lexer and returns the lexed statement
func (l *lexer) lex() Statement {
	for lexfn := lexStatement; lexfn != nil; {
		lexfn = lexfn(l)
	}
	return l.tokens
}

// next gets the next rune in the input and updates the lexer state
func (l *lexer) next() rune {
	r, w := utf8.DecodeRuneInString(l.text[l.pos:])

	l.pos += w
	l.width = w

	l.prev = l.cur
	l.cur = r

	return r
}

// backup moves the lexer back one rune
// can only be used once per call of next()
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns the next rune in the input
// without moving the internal pointer
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore skips the current token
func (l *lexer) ignore() {
	l.tokenStart = l.pos
}

// emit adds the current token to the token slice and
// moves the tokenStart pointer to the current position
func (l *lexer) emit(typ TokenTyp) {
	t := Token{
		Text: l.text[l.tokenStart:l.pos],
		Typ:  typ,
	}
	l.tokenStart = l.pos

	l.tokens = append(l.tokens, t)
}

// accept moves the pointer if the next rune is in
// the set of valid runes
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun continually accepts runes from the
// set of valid runes
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// a runeCheck is a function that determines if a rune is valid
// or not so that we can do complex checks against runes
type runeCheck func(rune) bool

// acceptFunc accepts a rune if the provided runeCheck
// function returns true
func (l *lexer) acceptFunc(fn runeCheck) bool {
	if fn(l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRunFunc continually accepts runes for as long
// as the runeCheck function returns true
func (l *lexer) acceptRunFunc(fn runeCheck) {
	for fn(l.next()) {
	}
	l.backup()
}

// acceptUntil accepts runes until it hits a delimiter
// rune contained in the provided string
func (l *lexer) acceptUntil(delims string) {
	for !strings.ContainsRune(delims, l.next()) {
		if l.cur == utf8.RuneError {
			return
		}
	}
	l.backup()
}

// acceptUntilUnescaped accepts runes until it hits a delimiter
// rune contained in the provided string, unless that rune was
// escaped with a backslash
func (l *lexer) acceptUntilUnescaped(delims string) {
	// Read until we hit an unescaped rune or the end of the input
	inEscape := false
	for {
		r := l.next()
		if r == '\\' && !inEscape {
			inEscape = true
			continue
		}
		if strings.ContainsRune(delims, r) && !inEscape {
			l.backup()
			return
		}
		if l.cur == utf8.RuneError {
			return
		}
		inEscape = false
	}
}

// a lexFn accepts a lexer, performs some action on it and
// then returns an appropriate lexFn for the next stage
type lexFn func(*lexer) lexFn

// lexStatement is the highest level lexFn. Its only job
// is to determine which more specific lexFn to use
func lexStatement(l *lexer) lexFn {
	r := l.peek()

	switch {
	case r == '.' || validFirstRune(r):
		return lexBareWord
	case r == '[':
		return lexBraces
	case r == ' ', r == '=':
		return lexValue
	case r == '-':
		// grep -A etc can add '--' lines to output
		// we'll save the text but not actually do
		// anything with them
		return lexIgnore
	case r == utf8.RuneError:
		return nil
	default:
		l.emit(TypError)
		return nil
	}
}

// lexBareWord lexes for bare identifiers.
// E.g: the 'foo' in 'foo.bar' or 'foo[0]' is a bare identifier
func lexBareWord(l *lexer) lexFn {
	if l.accept(".") {
		l.emit(TypDot)
	}

	if !l.acceptFunc(validFirstRune) {
		l.emit(TypError)
		return nil
	}
	l.acceptRunFunc(validSecondaryRune)
	l.emit(TypBare)

	return lexStatement
}

// lexBraces lexes keys contained within square braces
func lexBraces(l *lexer) lexFn {
	l.accept("[")
	l.emit(TypLBrace)

	switch {
	case unicode.IsNumber(l.peek()):
		return lexNumericKey
	case l.peek() == '"':
		return lexQuotedKey
	default:
		l.emit(TypError)
		return nil
	}
}

// lexNumericKey lexes numeric keys between square braces
func lexNumericKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.acceptRunFunc(unicode.IsNumber)
	l.emit(TypNumericKey)

	if l.accept("]") {
		l.emit(TypRBrace)
	} else {
		l.emit(TypError)
		return nil
	}
	l.ignore()
	return lexStatement
}

// lexQuotedKey lexes quoted keys between square braces
func lexQuotedKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.accept(`"`)

	l.acceptUntilUnescaped(`"`)
	l.accept(`"`)
	l.emit(TypQuotedKey)

	if l.accept("]") {
		l.emit(TypRBrace)
	} else {
		l.emit(TypError)
		return nil
	}
	l.ignore()
	return lexStatement
}

// lexValue lexes a value at the end of a statement
func lexValue(l *lexer) lexFn {
	l.acceptRun(" ")
	l.ignore()

	if l.accept("=") {
		l.emit(TypEquals)
	} else {
		return nil
	}
	l.acceptRun(" ")
	l.ignore()

	switch {

	case l.accept(`"`):
		l.acceptUntilUnescaped(`"`)
		l.accept(`"`)
		l.emit(TypString)

	case l.accept("t"):
		l.acceptRun("rue")
		l.emit(TypTrue)

	case l.accept("f"):
		l.acceptRun("alse")
		l.emit(TypFalse)

	case l.accept("n"):
		l.acceptRun("ul")
		l.emit(TypNull)

	case l.accept("["):
		l.accept("]")
		l.emit(TypEmptyArray)

	case l.accept("{"):
		l.accept("}")
		l.emit(TypEmptyObject)

	default:
		// Assume number
		l.acceptUntil(";")
		l.emit(TypNumber)
	}

	l.acceptRun(" ")
	l.ignore()

	if l.accept(";") {
		l.emit(TypSemi)
	}

	// The value should always be the last thing
	// in the statement
	return nil
}

// lexIgnore accepts runes until the end of the input
// and emits them as a typIgnored token
func lexIgnore(l *lexer) lexFn {
	l.acceptRunFunc(func(r rune) bool {
		return r != utf8.RuneError
	})
	l.emit(TypIgnored)
	return nil
}

// ungronTokens turns a slice of tokens into an actual datastructure
func ungronTokens(ts []Token) (interface{}, error) {
	if len(ts) == 0 {
		return nil, errRecoverable{"empty input"}
	}

	if ts[0].Typ == TypIgnored {
		return nil, errRecoverable{"ignored token"}
	}

	if ts[len(ts)-1].Typ == TypError {
		return nil, errors.New("invalid statement")
	}

	// The last token should be typSemi so we need to check
	// the second to last token is a value rather than the
	// last one
	if len(ts) > 1 && !ts[len(ts)-2].isValue() {
		return nil, errors.New("statement has no value")
	}

	t := ts[0]
	switch {
	case t.isPunct():
		// Skip the token
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		return val, nil

	case t.isValue():
		var val interface{}
		d := json.NewDecoder(strings.NewReader(t.Text))
		d.UseOrderedObject()
		d.UseNumber()
		err := d.Decode(&val)
		if err != nil {
			return nil, fmt.Errorf("invalid value `%s`", t.Text)
		}
		return val, nil

	case t.Typ == TypBare:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		out := json.OrderedObject{{Key: t.Text, Value: val}}
		return out, nil

	case t.Typ == TypQuotedKey:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		out := json.OrderedObject{{Key: strings.Trim(t.Text, `"`), Value: val}}
		return out, nil

	case t.Typ == TypNumericKey:
		key, err := strconv.Atoi(t.Text)
		if err != nil {
			return nil, fmt.Errorf("invalid integer key `%s`", t.Text)
		}

		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}

		// There needs to be at least key + 1 space in the array
		out := make([]interface{}, key+1)
		out[key] = val
		return out, nil

	default:
		return nil, fmt.Errorf("unexpected token `%s`", t.Text)
	}
}

// recursiveMerge merges maps and slices, or returns b for scalars
func recursiveMerge(a, b interface{}) (interface{}, error) {
	switch a.(type) {

	case json.OrderedObject:
		bCast, ok := b.(json.OrderedObject)
		if !ok {
			return nil, fmt.Errorf("cannot merge OrderedObject with non-OrderedObject %s", reflect.TypeOf(b))
		}
		return recursiveOrderedMerge(a.(json.OrderedObject), bCast)

	case map[string]interface{}:
		bCast, ok := b.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge map with non-map %s", reflect.TypeOf(b))
		}
		return recursiveMapMerge(a.(map[string]interface{}), bCast)

	case []interface{}:
		bSlice, ok := b.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge array with non-array")
		}
		return recursiveSliceMerge(a.([]interface{}), bSlice)

	case string, int, float64, bool, nil, json.Number:
		// Can't merge them, second one wins
		return b, nil

	default:
		return nil, fmt.Errorf("unexpected data type for merge: %s is %s", a, reflect.TypeOf(a))
	}
}

func recursiveOrderedMerge(a, b json.OrderedObject) (json.OrderedObject, error) {
	for _, bMember := range b {
		var found bool
		for i, aMember := range a {
			if bMember.Key == aMember.Key {
				found = true
				merged, err := recursiveMerge(aMember.Value, bMember.Value)
				if err != nil {
					return nil, err
				}
				a[i].Value = merged
			}
		}
		if !found {
			a = append(a, bMember)
		}
	}

	return a, nil
}

func recursiveMapMerge(a, b map[string]interface{}) (map[string]interface{}, error) {
	// Merge keys from b into a
	for k, v := range b {
		_, exists := a[k]
		if !exists {
			// Doesn't exist in a, just add it in
			a[k] = v
		} else {
			// Does exist, merge the values
			merged, err := recursiveMerge(a[k], b[k])
			if err != nil {
				return nil, err
			}

			a[k] = merged
		}
	}
	return a, nil
}

// recursiveSliceMerge recursively merged []interface{} values
func recursiveSliceMerge(a, b []interface{}) ([]interface{}, error) {
	// We need a new slice with the capacity of whichever
	// slive is biggest
	outLen := len(a)
	if len(b) > outLen {
		outLen = len(b)
	}
	out := make([]interface{}, outLen)

	// Copy the values from 'a' into the output slice
	copy(out, a)

	// Add the values from 'b'; merging existing keys
	for k, v := range b {
		if out[k] == nil {
			out[k] = v
		} else if v != nil {
			merged, err := recursiveMerge(out[k], b[k])
			if err != nil {
				return nil, err
			}
			out[k] = merged
		}
	}
	return out, nil
}
