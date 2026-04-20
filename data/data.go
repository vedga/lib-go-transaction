package data

import (
	"errors"
	"io"
)

type (
	implementation[T any] struct {
		kind string
		o    *T
	}
)

var (
	// ErrInvalidTransformation indicate invalid value transformation
	ErrInvalidTransformation = errors.New("invalid data transformation")
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

	return nil, ErrInvalidTransformation
}

// Kind implementation Serializable interface
func (i *implementation[T]) Kind() string {
	return i.kind
}

// Write implementation Serializable interface
// This method write only value content from the current place, kind not used.
func (i *implementation[T]) Write(w io.Writer, codec Codec) error {
	return codec.Write(w, i.o)
}

// Read implementation Serializable interface
// This method read only value content to the current place, kind value stay unchanged.
func (i *implementation[T]) Read(r io.Reader, codec Codec) error {
	return codec.Read(r, i.o)
}
