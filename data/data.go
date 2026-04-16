package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type (
	implementation[T any] struct {
		kind string
		o    *T
	}
)

var (
	// ErrInvalidValue indicate invalid value transformation
	ErrInvalidValue = errors.New("invalid value")
)

// NewProducer return new producer of entity type T
func NewProducer[T any](kind string) Producer {
	return func(setup ...Setup) (Serializable, error) {
		o := new(T)

		for _, fn := range setup {
			if e := fn(o); e != nil {
				return nil, e
			}
		}

		return &implementation[T]{
			kind: kind,
			o:    o,
		}, nil
	}
}

// Ref return pointer to the value of required type or error
func Ref[T any](v Serializable) (*T, error) {
	if i, valid := v.(*implementation[T]); valid {
		return i.o, nil
	}

	return nil, ErrInvalidValue
}

// Kind implementation Serializable interface
func (i *implementation[T]) Kind() string {
	return i.kind
}

// Write implementation Serializable interface
func (i *implementation[T]) Write(w io.Writer) error {
	encoder := json.NewEncoder(w)

	if e := encoder.Encode(i.o); e != nil {
		return fmt.Errorf(`json encode error: %w`, e)
	}

	return nil

}

// Read implementation Serializable interface
func (i *implementation[T]) Read(r io.Reader) error {
	decoder := json.NewDecoder(r)

	decoder.DisallowUnknownFields()

	if e := decoder.Decode(i.o); e != nil {
		return fmt.Errorf(`json decode error: %w`, e)
	}

	return nil
}
