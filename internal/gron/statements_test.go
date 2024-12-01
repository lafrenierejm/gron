package gron

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	json "github.com/virtuald/go-ordered-json"
)

func statementsFromStringSlice(strs []string) Statements {
	ss := make(Statements, len(strs))
	for i, str := range strs {
		ss[i] = StatementFromString(str)
	}
	return ss
}

func TestStatementsSimple(t *testing.T) {
	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"a_null": null,
		"an_arr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1,
		"id": 66912849,
		"": 2
	}`)

	ss, err := StatementsFromJSON(MakeDecoder(bytes.NewReader(j), false, false), Statement{{"json", TypBare}})
	if err != nil {
		t.Errorf("Want nil error from makeStatementsFromJSON() but got %s", err)
	}

	wants := statementsFromStringSlice([]string{
		`json = {};`,
		`json.dotted = "A dotted value";`,
		`json["a quoted"] = "value";`,
		`json.bool1 = true;`,
		`json.bool2 = false;`,
		`json.a_null = null;`,
		`json.an_arr = [];`,
		`json.an_arr[0] = 1;`,
		`json.an_arr[1] = 1.5;`,
		`json.anob = {};`,
		`json.anob.foo = "bar";`,
		`json["else"] = 1;`,
		`json.id = 66912849;`,
		`json[""] = 2;`,
	})

	t.Logf("Have: %#v", ss)
	for _, want := range wants {
		if !ss.Contains(want) {
			t.Errorf("Statement group should contain `%s` but doesn't", want)
		}
	}
}

func TestStatementsSimpleYaml(t *testing.T) {
	j := []byte(`'': 2
a quoted: value
an_arr:
- 1
- 1.5
anob:
  foo: bar
a_null:
bool1: true
bool2: false
dotted: A dotted value
else: 1
x: |
  y: "z"
id: 66912849`)

	ss, err := StatementsFromJSON(MakeDecoder(bytes.NewReader(j), true, false), Statement{{"yaml", TypBare}})
	if err != nil {
		t.Errorf("Want nil error from makeStatementsFromJSON() but got %s", err)
	}

	wants := statementsFromStringSlice([]string{
		`yaml = {};`,
		`yaml.dotted = "A dotted value";`,
		`yaml["a quoted"] = "value";`,
		`yaml.bool1 = true;`,
		`yaml.bool2 = false;`,
		`yaml.a_null = null;`,
		`yaml.an_arr = [];`,
		`yaml.an_arr[0] = 1;`,
		`yaml.an_arr[1] = 1.5;`,
		`yaml.anob = {};`,
		`yaml.anob.foo = "bar";`,
		`yaml["else"] = 1;`,
		`yaml.id = 66912849;`,
		`yaml[""] = 2;`,
		`yaml.x = "y: \"z\"\n";`,
	})

	t.Logf("Have: %#v", ss)
	for _, want := range wants {
		if !ss.Contains(want) {
			t.Errorf("Statement group should contain `%s` but doesn't", want)
		}
	}
}

func TestStatementsSorting(t *testing.T) {
	want := statementsFromStringSlice([]string{
		`json.a = true;`,
		`json.b = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.c[10] = true;`,
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	})

	have := statementsFromStringSlice([]string{
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.b = true;`,
		`json.c[10] = true;`,
		`json.c[21][11] = true;`,
		`json.a = true;`,
	})

	sort.Sort(have)

	for i := range want {
		if !reflect.DeepEqual(have[i], want[i]) {
			t.Errorf("Statements sorted incorrectly; want `%s` at index %d, have `%s`", want[i], i, have[i])
		}
	}
}

func BenchmarkStatementsLess(b *testing.B) {
	ss := statementsFromStringSlice([]string{
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	})

	for i := 0; i < b.N; i++ {
		_ = ss.Less(0, 1)
	}
}

func BenchmarkFill(b *testing.B) {
	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"a_null": null,
		"an_arr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1
	}`)

	var top interface{}
	err := json.Unmarshal(j, &top)
	if err != nil {
		b.Fatalf("Failed to unmarshal test file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		ss := make(Statements, 0)
		ss.fill(Statement{{"json", TypBare}}, top)
	}
}

func TestUngronStatementsSimple(t *testing.T) {
	in := statementsFromStringSlice([]string{
		`json.contact = {};`,
		`json.contact.twitter = "@TomNomNom";`,
		`json.contact["e-mail"][0] = "mail@tomnomnom.com";`,
		`json.contact["e-mail"][1] = "test@tomnomnom.com";`,
		`json.contact["e-mail"][3] = "foo@tomnomnom.com";`,
	})

	want := json.OrderedObject{
		{
			Key: "json", Value: json.OrderedObject{
				{
					Key: "contact", Value: json.OrderedObject{
						{
							Key:   "twitter",
							Value: "@TomNomNom",
						},
						{
							Key: "e-mail", Value: []interface{}{
								"mail@tomnomnom.com",
								"test@tomnomnom.com",
								nil,
								"foo@tomnomnom.com",
							},
						},
					},
				},
			},
		},
	}

	have, err := in.ToInterface()
	if err != nil {
		t.Fatalf("want nil error but have: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)

	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("have and want are not equal")
	}
}

func TestUngronStatementsInvalid(t *testing.T) {
	cases := []Statements{
		statementsFromStringSlice([]string{``}),
		statementsFromStringSlice([]string{`this isn't a statement at all`}),
		statementsFromStringSlice([]string{`json[0] = 1;`, `json.bar = 1;`}),
	}

	for _, c := range cases {
		_, err := c.ToInterface()
		if err == nil {
			t.Errorf("want non-nil error; have nil")
		}
	}
}

func TestStatement(t *testing.T) {
	s := Statement{
		Token{"json", TypBare},
		Token{".", TypDot},
		Token{"foo", TypBare},
		Token{"=", TypEquals},
		Token{"2", TypNumber},
		Token{";", TypSemi},
	}

	have := s.String()
	want := "json.foo = 2;"
	if have != want {
		t.Errorf("have: `%s` want: `%s`", have, want)
	}
}
