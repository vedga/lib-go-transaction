package data

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type (
	// Serializable interface allow object to save and load by readers
	Serializable interface {
		Write(w io.Writer) error
		Read(r io.Reader) error
	}

	// Producer function for serializable objects
	Producer func(setup ...Setup) (*Descriptor, error)

	// Producers declare producers map
	Producers []Producer

	// Coder function
	Coder func(o any) Serializable

	// Raw data format
	Raw = []byte

	// Container for exchange information
	Container struct {
		Kind    string `json:"kind"`
		Payload Raw    `json:"payload"`
	}

	// Manager implementation
	Manager struct {
		producersMap map[string]Producer
		outerCoder   Coder
		innerCoder   Coder
	}

	// Option is Manager configuration modifier
	Option func(*Manager)
)

var (
	// ErrNotSupported indicate container content not supported
	ErrNotSupported = errors.New("container is not supported")
)

// NewManager return data manager implementation
func NewManager(producers Producers, options ...Option) *Manager {
	producersMap := make(map[string]Producer, len(producers))
	for _, descriptorProducer := range producers {
		descriptor, e := descriptorProducer()
		if e != nil {
			panic(`Invalid data producer: ` + e.Error())
		}

		kind := descriptor.kind
		if _, dup := producersMap[kind]; dup {
			panic(`data ` + kind + ` is duplicated.`)
		}

		producersMap[kind] = descriptorProducer
	}

	i := &Manager{
		producersMap: producersMap,
		outerCoder:   jsonCoder,
		innerCoder:   jsonCoder,
	}

	// Apply options
	for _, opt := range options {
		opt(i)
	}

	return i

}

// WithOuterCoder apply outer coder
func WithOuterCoder(coder Coder) Option {
	return func(i *Manager) {
		i.outerCoder = coder
	}
}

// WithInnerCoder apply inner coder
func WithInnerCoder(coder Coder) Option {
	return func(i *Manager) {
		i.innerCoder = coder
	}
}

// Write data to io.Writer
func (i *Manager) Write(w io.Writer, descriptor *Descriptor) error {
	c, e := i.NewContainer(descriptor)
	if e != nil {
		return fmt.Errorf(`data exchange container build error: %w`, e)
	}

	// Encode container
	return i.outerCoder(c).Write(w)

	/*
		kind := descriptor.kind

		// May be removing this check for improve performance?
		if _, e := i.New(kind); e != nil {
			return e
		}

		b := &bytes.Buffer{}
		if e := i.innerCoder(descriptor.value).Write(b); e != nil {
			return fmt.Errorf(`encode payload error: %w`, e)
		}

		return i.outerCoder(&Container{
			Kind:    kind,
			Payload: b.Bytes(),
		}).Write(w)
	*/
}

// NewContainer return data exchange container
func (i *Manager) NewContainer(descriptor *Descriptor) (*Container, error) {
	kind := descriptor.kind

	// May be removing this check for improve performance?
	if _, e := i.New(kind); e != nil {
		return nil, e
	}

	b := &bytes.Buffer{}
	if e := i.innerCoder(descriptor.value).Write(b); e != nil {
		return nil, fmt.Errorf(`encode payload error: %w`, e)
	}

	// Return data container
	return &Container{
		Kind:    kind,
		Payload: b.Bytes(),
	}, nil
}

// Read data from io.Reader
func (i *Manager) Read(r io.Reader) (*Descriptor, error) {
	c := new(Container)

	if e := i.outerCoder(c).Read(r); e != nil {
		return nil, fmt.Errorf(`read container error: %w`, e)
	}

	// When read data options not used
	o, e := i.New(c.Kind)
	if e != nil {
		return nil, fmt.Errorf(`data can't be read: %w`, e)
	}

	b := bytes.NewBuffer(c.Payload)

	if e = i.innerCoder(o.value).Read(b); e != nil {
		return nil, fmt.Errorf(`decode payload error: %w`, e)
	}

	return o, nil
}

// New return data descriptor
func (i *Manager) New(kind string, setup ...Setup) (*Descriptor, error) {
	producer := i.getProducer(kind)
	if producer == nil {
		return nil, ErrNotSupported
	}

	return producer(setup...)
}

// getProducer return specified producer
func (i *Manager) getProducer(kind string) Producer {
	return i.producersMap[kind]
}
