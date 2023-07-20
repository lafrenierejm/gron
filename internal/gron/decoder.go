package gron

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
)

// an ActionFn represents a main action of the program, it accepts
// an input, output and a bitfield of options; returning an exit
// code and any error that occurred
type ActionFn func(io.Reader, io.Writer, int) (int, error)

type Decoder interface {
	Decode(interface{}) error
}

func MakeDecoder(r io.Reader, optYAML int) Decoder {
	if optYAML > 0 {
		return yaml.NewDecoder(r)
	} else {
		d := json.NewDecoder(r)
		d.UseNumber()
		return d
	}
}
