package gron

import (
	"bytes"

	"github.com/fatih/color"
	"github.com/nwidger/jsoncolor"
)

var (
	strColor   = color.New(color.FgYellow)
	braceColor = color.New(color.FgMagenta)
	bareColor  = color.New(color.FgBlue, color.Bold)
	numColor   = color.New(color.FgRed)
	boolColor  = color.New(color.FgCyan)
)

// a sprintFn adds color to its input
type sprintFn func(...interface{}) string

// mapping of token types to the appropriate color sprintFn
var sprintFns = map[TokenTyp]sprintFn{
	TypBare:        bareColor.SprintFunc(),
	TypNumericKey:  numColor.SprintFunc(),
	TypQuotedKey:   strColor.SprintFunc(),
	TypLBrace:      braceColor.SprintFunc(),
	TypRBrace:      braceColor.SprintFunc(),
	TypString:      strColor.SprintFunc(),
	TypNumber:      numColor.SprintFunc(),
	TypTrue:        boolColor.SprintFunc(),
	TypFalse:       boolColor.SprintFunc(),
	TypNull:        boolColor.SprintFunc(),
	TypEmptyArray:  braceColor.SprintFunc(),
	TypEmptyObject: braceColor.SprintFunc(),
}

func colorizeJSON(src []byte) ([]byte, error) {
	out := &bytes.Buffer{}
	f := jsoncolor.NewFormatter()

	f.StringColor = strColor
	f.ObjectColor = braceColor
	f.ArrayColor = braceColor
	f.FieldColor = bareColor
	f.NumberColor = numColor
	f.TrueColor = boolColor
	f.FalseColor = boolColor
	f.NullColor = boolColor

	err := f.Format(out, src)
	if err != nil {
		return out.Bytes(), err
	}
	return out.Bytes(), nil
}
