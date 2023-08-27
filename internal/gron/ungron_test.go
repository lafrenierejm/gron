package gron

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	json "github.com/virtuald/go-ordered-json"
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
	want := json.OrderedObject{
		{
			Key: "json",
			Value: json.OrderedObject{
				{
					Key: "contact",
					Value: json.OrderedObject{
						{Key: "e-mail", Value: []interface{}{"mail@tomnomnom.com"}},
					},
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
	a := json.OrderedObject{
		json.Member{
			Key: "json", Value: json.OrderedObject{
				json.Member{
					Key: "contact", Value: json.OrderedObject{
						json.Member{
							Key: "e-mail", Value: json.OrderedObject{
								json.Member{
									Key:   "0",
									Value: "mail@tomnomnom.com",
								},
							},
						},
					},
				},
			},
		},
	}

	b := json.OrderedObject{
		json.Member{
			Key: "json", Value: json.OrderedObject{
				json.Member{
					Key: "contact", Value: json.OrderedObject{
						json.Member{
							Key: "e-mail", Value: json.OrderedObject{
								{
									Key:   "1",
									Value: "test@tomnomnom.com",
								},
								{
									Key:   "3",
									Value: "foo@tomnomnom.com",
								},
							},
						},
						json.Member{
							Key:   "twitter",
							Value: "@TomNomNom",
						},
					},
				},
			},
		},
	}

	want := json.OrderedObject{
		{
			Key: "json", Value: json.OrderedObject{
				{
					Key: "contact", Value: json.OrderedObject{
						{
							Key: "e-mail", Value: json.OrderedObject{
								{
									Key:   "0",
									Value: "mail@tomnomnom.com",
								},
								{
									Key:   "1",
									Value: "test@tomnomnom.com",
								},
								{
									Key:   "3",
									Value: "foo@tomnomnom.com",
								},
							},
						},
						{
							Key:   "twitter",
							Value: "@TomNomNom",
						},
					},
				},
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
	if !reflect.DeepEqual(have, want) {
		t.Errorf("Have and want datastructures are unequal")
	}
}

func TestUngron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.gron", "testdata/one.json"},
		{"testdata/two.gron", "testdata/two.json"},
		{"testdata/three.gron", "testdata/three.json"},
		{"testdata/grep-separators.gron", "testdata/grep-separators.json"},
		{"testdata/github.sorted.gron", "testdata/github.json"},
		// {"testdata/large-line.gron", "testdata/large-line.json", true},
		{"testdata/duplicate-numeric.gron", "testdata/duplicate-numeric.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := Ungron(in, out, false, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}

func TestUngronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.jgron", "testdata/one.json"},
		{"testdata/two.jgron", "testdata/two.json"},
		{"testdata/three.jgron", "testdata/three.json"},
		{"testdata/github.jgron", "testdata/github.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := Ungron(in, out, true, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}
