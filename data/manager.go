package data

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type (
	// Bytes are alias for byte sequence
	Bytes = []byte

	// Manager implementation
	Manager struct {
		producers  map[string]Producer
		outerCodec Codec
		innerCodec Codec
	}

	container struct {
		Kind string `json:"k"`
		Data Bytes  `json:"d"`
	}

	// Option is Manager configuration type
	Option func(*Manager)
)

var (
	// ErrUnsupportedData indicate data isn't supported
	ErrUnsupportedData = errors.New("data type not supported")
)

// NewManager return new data manager implementation
// Note:
// Attempt to use incompatible data producers cause panic because case indicate program design error.
func NewManager(options ...Option) *Manager {
	i := &Manager{
		producers:  make(map[string]Producer),
		outerCodec: NewCodecJSON(),
		innerCodec: NewCodecJSON(),
	}

	// Apply options
	for _, option := range options {
		option(i)
	}

	return i
}

// WithProducer add data producer
func WithProducer(kind string, producer Producer) Option {
	return func(i *Manager) {
		if _, dup := i.producers[kind]; dup {
			panic(fmt.Sprintf("data producer already defined: %v", kind))
		}

		i.producers[kind] = producer
	}
}

// Encode object to the byte sequence
func (i *Manager) Encode(kind string, o any) (Bytes, error) {
	b := new(bytes.Buffer)
	if e := i.Write(b, kind, o); e != nil {
		return nil, e
	}

	return b.Bytes(), nil
}

// Write object to the io.Writer with type kind
func (i *Manager) Write(w io.Writer, kind string, o any) error {
	encoded, e := Encode(i.innerCodec, o)
	if e != nil {
		return fmt.Errorf("data encode error: %w", e)
	}

	// Write container
	if e := i.outerCodec.Write(w, &container{
		Kind: kind,
		Data: encoded,
	}); e != nil {
		return fmt.Errorf("container encode error: %w", e)
	}

	return nil
}

// Decode bytes sequence to the object
func (i *Manager) Decode(source Bytes, setup ...Setup) (string, any, error) {
	return i.Read(bytes.NewReader(source), setup...)
}

// Read object from the io.Reader
func (i *Manager) Read(r io.Reader, setup ...Setup) (string, any, error) {
	// Read container
	c := new(container)
	if e := i.outerCodec.Read(r, c); e != nil {
		return ``, nil, fmt.Errorf("container decode error: %w", e)
	}

	kind := c.Kind

	o, e := i.New(c.Kind, setup...)
	if e != nil {
		return kind, nil, e
	}

	if e = Decode(i.innerCodec, c.Data, o); e != nil {
		return kind, nil, fmt.Errorf("container decode error: %w", e)
	}

	return kind, o, nil
}

// New return new initialized data for specified kind type
func (i *Manager) New(kind string, setup ...Setup) (any, error) {
	producer, supported := i.producers[kind]
	if !supported {
		return nil, ErrUnsupportedData
	}

	return producer(setup...)
}
