package data

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	codecJSON struct {
	}
)

// NewCodecJSON return new JSON codec implementation
func NewCodecJSON() Codec {
	return &codecJSON{}
}

// Write is implementation of Codec interface
func (i *codecJSON) Write(w io.Writer, o any) error {
	encoder := json.NewEncoder(w)

	if e := encoder.Encode(o); e != nil {
		return fmt.Errorf(`json encode error: %w`, e)
	}

	return nil
}

// Read is implementation of Codec interface
func (i *codecJSON) Read(r io.Reader, o any) error {
	decoder := json.NewDecoder(r)

	decoder.DisallowUnknownFields()

	if e := decoder.Decode(o); e != nil {
		return fmt.Errorf(`json decode error: %w`, e)
	}

	return nil
}
