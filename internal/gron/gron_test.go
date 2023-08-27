package gron

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGronUnsorted(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.gron"},
		{"testdata/two.json", "testdata/two.gron"},
		{"testdata/three.json", "testdata/three.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := Gron(in, out, StatementToString, false, false, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestGronSorted(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		// {"testdata/one.json", "testdata/one.sorted.gron"},
		{"testdata/two.json", "testdata/two.sorted.gron"},
		{"testdata/three.json", "testdata/three.sorted.gron"},
		{"testdata/github.json", "testdata/github.sorted.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := Gron(in, out, StatementToString, false, true, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.gron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := GronStream(in, out, StatementToString, false, true, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestLargeGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/long-stream.json", "testdata/long-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := GronStream(in, out, StatementToString, false, true, false)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestGronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
		sort    bool
	}{
		{"testdata/one.json", "testdata/one.jgron", false},
		{"testdata/two.json", "testdata/two.jgron", false},
		{"testdata/three.json", "testdata/three.jgron", false},
		{"testdata/github.json", "testdata/github.jgron", true},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := Gron(in, out, StatementToString, false, c.sort, true)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestGronStreamJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.jgron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.jgron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := GronStream(in, out, StatementToString, false, true, true)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}
}

func BenchmarkBigJSON(b *testing.B) {
	in, err := os.Open("testdata/big.json")
	if err != nil {
		b.Fatalf("failed to open test data file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		out := &bytes.Buffer{}
		_, err = in.Seek(0, 0)
		if err != nil {
			b.Fatalf("failed to rewind input: %s", err)
		}

		_, err := Gron(in, out, StatementToString, false, true, false)
		if err != nil {
			b.Fatalf("failed to gron: %s", err)
		}
	}
}
