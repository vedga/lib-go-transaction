package data

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	payload struct {
		o any
	}
)

func jsonCoder(o any) Serializable {
	return &payload{
		o: o,
	}
}

// Write payload to io.Writer
func (i *payload) Write(w io.Writer) error {
	encoder := json.NewEncoder(w)

	if e := encoder.Encode(i.o); e != nil {
		return fmt.Errorf(`encode payload error: %w`, e)
	}

	return nil
}

// Read payload from io.Reader
func (i *payload) Read(r io.Reader) error {
	decoder := json.NewDecoder(r)

	decoder.DisallowUnknownFields()

	if e := decoder.Decode(i.o); e != nil {
		return fmt.Errorf(`decode payload error: %w`, e)
	}

	return nil
}
