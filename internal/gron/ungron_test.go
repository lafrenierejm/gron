package gron

import (
	"reflect"
	"testing"
)

func TestLex(t *testing.T) {
	cases := []struct {
		in   string
		want []Token
	}{
		{`json.foo = 1;`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`foo`, TypBare},
			{`=`, TypEquals},
			{`1`, TypNumber},
			{`;`, TypSemi},
		}},

		{`json.foo = "bar";`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`foo`, TypBare},
			{`=`, TypEquals},
			{`"bar"`, TypString},
			{`;`, TypSemi},
		}},

		{`json.foo = "ba;r";`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`foo`, TypBare},
			{`=`, TypEquals},
			{`"ba;r"`, TypString},
			{`;`, TypSemi},
		}},

		{`json.foo = "ba\"r ;";`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`foo`, TypBare},
			{`=`, TypEquals},
			{`"ba\"r ;"`, TypString},
			{`;`, TypSemi},
		}},

		{`json = "\\";`, []Token{
			{`json`, TypBare},
			{`=`, TypEquals},
			{`"\\"`, TypString},
			{`;`, TypSemi},
		}},

		{`json = "\\\\";`, []Token{
			{`json`, TypBare},
			{`=`, TypEquals},
			{`"\\\\"`, TypString},
			{`;`, TypSemi},
		}},

		{`json = "f\oo\\";`, []Token{
			{`json`, TypBare},
			{`=`, TypEquals},
			{`"f\oo\\"`, TypString},
			{`;`, TypSemi},
		}},

		{`json.value = "\u003c ;";`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`value`, TypBare},
			{`=`, TypEquals},
			{`"\u003c ;"`, TypString},
			{`;`, TypSemi},
		}},

		{`json[0] = "bar";`, []Token{
			{`json`, TypBare},
			{`[`, TypLBrace},
			{`0`, TypNumericKey},
			{`]`, TypRBrace},
			{`=`, TypEquals},
			{`"bar"`, TypString},
			{`;`, TypSemi},
		}},

		{`json["foo"] = "bar";`, []Token{
			{`json`, TypBare},
			{`[`, TypLBrace},
			{`"foo"`, TypQuotedKey},
			{`]`, TypRBrace},
			{`=`, TypEquals},
			{`"bar"`, TypString},
			{`;`, TypSemi},
		}},

		{`json.foo["bar"][0] = "bar";`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{`foo`, TypBare},
			{`[`, TypLBrace},
			{`"bar"`, TypQuotedKey},
			{`]`, TypRBrace},
			{`[`, TypLBrace},
			{`0`, TypNumericKey},
			{`]`, TypRBrace},
			{`=`, TypEquals},
			{`"bar"`, TypString},
			{`;`, TypSemi},
		}},

		{`not an identifier at all`, []Token{
			{`not`, TypBare},
		}},

		{`alsonotanidentifier`, []Token{
			{`alsonotanidentifier`, TypBare},
		}},

		{`wat!`, []Token{
			{`wat`, TypBare},
			{``, TypError},
		}},

		{`json[ = 1;`, []Token{
			{`json`, TypBare},
			{`[`, TypLBrace},
			{``, TypError},
		}},

		{`json.[2] = 1;`, []Token{
			{`json`, TypBare},
			{`.`, TypDot},
			{``, TypError},
		}},

		{`json[1 = 1;`, []Token{
			{`json`, TypBare},
			{`[`, TypLBrace},
			{`1`, TypNumericKey},
			{``, TypError},
		}},

		{`json["foo] = 1;`, []Token{
			{`json`, TypBare},
			{`[`, TypLBrace},
			{`"foo] = 1;`, TypQuotedKey},
			{``, TypError},
		}},

		{`--`, []Token{
			{`--`, TypIgnored},
		}},

		{`json  =  1;`, []Token{
			{`json`, TypBare},
			{`=`, TypEquals},
			{`1`, TypNumber},
			{`;`, TypSemi},
		}},

		{`json=1;`, []Token{
			{`json`, TypBare},
			{`=`, TypEquals},
			{`1`, TypNumber},
			{`;`, TypSemi},
		}},
	}

	for _, c := range cases {
		l := newLexer(c.in)
		have := l.lex()

		if len(have) != len(c.want) {
			t.Logf("Input: %#v", c.in)
			t.Logf("Want: %#v", c.want)
			t.Logf("Have: %#v", have)
			t.Fatalf("want %d token.tokens, have %d", len(c.want), len(have))
		}

		for i := range have {
			if have[i] != c.want[i] {
				t.Logf("Input: %#v", c.in)
				t.Logf("Want: %#v", c.want)
				t.Logf("Have: %#v", have)
				t.Errorf("Want `%#v` in position %d, have `%#v`", c.want[i], i, have[i])
			}
		}
	}
}

func TestTokensSimple(t *testing.T) {
	in := `json.contact["e-mail"][0] = "mail@tomnomnom.com";`
	want := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					"mail@tomnomnom.com",
				},
			},
		},
	}

	l := newLexer(in)
	tokens := l.lex()
	have, err := ungronTokens(tokens)
	if err != nil {
		t.Fatalf("failed to ungron statement: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)

	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("Have and want datastructures are unequal")
	}
}

func TestTokensInvalid(t *testing.T) {
	cases := []struct {
		in []Token
	}{
		{[]Token{{``, TypError}}},                           // Error token.token
		{[]Token{{`foo`, TypString}}},                       // Invalid value
		{[]Token{{`"foo`, TypQuotedKey}, {"1", TypNumber}}}, // Invalid quoted key
		{[]Token{{`foo`, TypNumericKey}, {"1", TypNumber}}}, // Invalid numeric key
		{[]Token{{``, -255}, {"1", TypNumber}}},             // Invalid token.token type
	}

	for _, c := range cases {
		_, err := ungronTokens(c.in)
		if err == nil {
			t.Errorf("want non-nil error for %#v; have nil", c.in)
		}
	}
}

func TestMerge(t *testing.T) {
	a := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					0: "mail@tomnomnom.com",
				},
			},
		},
	}

	b := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					1: "test@tomnomnom.com",
					3: "foo@tomnomnom.com",
				},
				"twitter": "@TomNomNom",
			},
		},
	}

	want := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					0: "mail@tomnomnom.com",
					1: "test@tomnomnom.com",
					3: "foo@tomnomnom.com",
				},
				"twitter": "@TomNomNom",
			},
		},
	}

	t.Logf("A: %#v", a)
	t.Logf("B: %#v", b)
	have, err := recursiveMerge(a, b)
	if err != nil {
		t.Fatalf("failed to merge datastructures: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)
	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("Have and want datastructures are unequal")
	}
}
