package gron

import (
	"encoding/json"
	"testing"
)

var cases = []struct {
	in   interface{}
	want Token
}{
	{make(map[string]interface{}), Token{"{}", TypEmptyObject}},
	{make([]interface{}, 0), Token{"[]", TypEmptyArray}},
	{json.Number("1.2"), Token{"1.2", TypNumber}},
	{"foo", Token{`"foo"`, TypString}},
	{"<3", Token{`"<3"`, TypString}},
	{"&", Token{`"&"`, TypString}},
	{"\b", Token{`"\b"`, TypString}},
	{"\f", Token{`"\f"`, TypString}},
	{"\n", Token{`"\n"`, TypString}},
	{"\r", Token{`"\r"`, TypString}},
	{"\t", Token{`"\t"`, TypString}},
	{"wat \u001e", Token{`"wat \u001E"`, TypString}},
	{"Hello, 世界", Token{`"Hello, 世界"`, TypString}},
	{true, Token{"true", TypTrue}},
	{false, Token{"false", TypFalse}},
	{nil, Token{"null", TypNull}},
	{struct{}{}, Token{"", TypError}},
}

func TestValueTokenFromInterface(t *testing.T) {
	for _, c := range cases {
		have := valueTokenFromInterface(c.in)

		if have != c.want {
			t.Logf("input: %#v", have)
			t.Logf("have: %#v", have)
			t.Logf("want: %#v", c.want)
			t.Errorf("have != want")
		}
	}
}

func BenchmarkValueTokenFromInterface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			_ = valueTokenFromInterface(c.in)
		}
	}
}
