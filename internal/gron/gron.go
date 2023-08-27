package gron

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
)

// Exit codes
const (
	exitOK = iota
	exitOpenFile
	exitReadInput
	exitFormStatements
	exitFetchURL
	exitParseStatements
	exitJSONEncode
)

// Gron is the default action. Given JSON as the input it returns a list
// of assignment statements. Possible options are optNoSort and optMonochrome
func Gron(r io.Reader, w io.Writer, conv StatementConv, inYaml bool, sortOutput bool, outJson bool) (int, error) {
	var err error

	ss, err := StatementsFromJSON(MakeDecoder(r, inYaml, sortOutput), Statement{{"json", TypBare}})
	if err != nil {
		goto out
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	if sortOutput {
		sort.Sort(ss)
	}

	for _, s := range ss {
		if outJson {
			s, err = s.Jsonify()
			if err != nil {
				goto out
			}
		}
		fmt.Fprintln(w, conv(s))
	}

out:
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %s", err)
	}
	return exitOK, nil
}

// GronStream is like the gron action, but it treats the input as one
// JSON object per line. There's a bit of code duplication from the
// gron action, but it'd be fairly messy to combine the two actions
func GronStream(
	r io.Reader,
	w io.Writer,
	conv StatementConv,
	inYaml bool,
	outSort bool,
	outJson bool,
) (int, error) {
	var err error
	errstr := "failed to form statements"
	var i int
	var sc *bufio.Scanner
	var buf []byte

	// Helper function to make the prefix statements for each line
	makePrefix := func(index int) Statement {
		return Statement{
			{"json", TypBare},
			{"[", TypLBrace},
			{fmt.Sprintf("%d", index), TypNumericKey},
			{"]", TypRBrace},
		}
	}

	// The first line of output needs to establish that the top-level
	// thing is actually an array...
	top := Statement{
		{"json", TypBare},
		{"=", TypEquals},
		{"[]", TypEmptyArray},
		{";", TypSemi},
	}

	if outJson {
		top, err = top.Jsonify()
		if err != nil {
			goto out
		}
	}

	fmt.Fprintln(w, conv(top))

	// Read the input line by line
	sc = bufio.NewScanner(r)
	buf = make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	i = 0
	for sc.Scan() {

		line := bytes.NewBuffer(sc.Bytes())

		var ss Statements
		ss, err = StatementsFromJSON(MakeDecoder(line, inYaml, outSort), makePrefix(i))
		i++
		if err != nil {
			goto out
		}

		// Go's maps do not have well-defined ordering, but we want a consistent
		// output for a given input, so we must sort the statements
		if outSort {
			sort.Sort(ss)
		}

		for _, s := range ss {
			if outJson {
				s, err = s.Jsonify()
				if err != nil {
					goto out
				}

			}
			fmt.Fprintln(w, conv(s))
		}
	}
	if err = sc.Err(); err != nil {
		errstr = "error reading multiline input: %s"
	}

out:
	if err != nil {
		return exitFormStatements, fmt.Errorf(errstr+": %s", err)
	}
	return exitOK, nil
}
