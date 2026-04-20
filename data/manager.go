package data

import (
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
func WithProducer(producer Producer) Option {
	return func(i *Manager) {
		// Try to create entity w/o special data setup options.
		o, e := producer()
		if e != nil {
			panic(fmt.Sprintf("data producer instantiation error: %v", e))
		}

		// Check for duplicate producer instantiation
		kind := o.Kind()
		if _, dup := i.producers[kind]; dup {
			panic(fmt.Sprintf("data producer already defined: %v", kind))
		}

		i.producers[kind] = producer
	}
}

// Write object to the io.Writer
func (i *Manager) Write(w io.Writer, o Serializable) error {
	b := NewBytesReaderWriter(nil)

	// Encode data
	if e := o.Write(b, i.innerCodec); e != nil {
		return fmt.Errorf("data encode error: %w", e)
	}

	// Write container
	if e := i.outerCodec.Write(w, &container{
		Kind: o.Kind(),
		Data: b.Bytes(),
	}); e != nil {
		return fmt.Errorf("container encode error: %w", e)
	}

	return nil
}

// Read object from the io.Reader
func (i *Manager) Read(r io.Reader, setup ...Setup) (Serializable, error) {
	// Read container
	c := new(container)
	if e := i.outerCodec.Read(r, c); e != nil {
		return nil, fmt.Errorf("container decode error: %w", e)
	}

	o, e := i.New(c.Kind, setup...)
	if e != nil {
		return nil, e
	}

	if e = o.Read(NewBytesReaderWriter(c.Data), i.innerCodec); e != nil {
		return nil, fmt.Errorf("container decode error: %w", e)
	}

	return o, nil
}

// New return new initialized data for specified kind type
func (i *Manager) New(kind string, setup ...Setup) (Serializable, error) {
	producer, supported := i.producers[kind]
	if !supported {
		return nil, ErrUnsupportedData
	}

	return producer(setup...)
}
