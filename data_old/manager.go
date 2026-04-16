package data_old

import (
	"errors"
	"fmt"
	"io"
)

type (
	// Producer function for serializable objects
	Producer func(setup ...Setup) (*Descriptor, error)

	// Producers declare producers map
	Producers []Producer

	// Coder function
	Coder func(o any) Serializable

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

// NewManager return data_old manager implementation
func NewManager(producers Producers, options ...Option) *Manager {
	producersMap := make(map[string]Producer, len(producers))
	for _, descriptorProducer := range producers {
		descriptor, e := descriptorProducer()
		if e != nil {
			panic(`Invalid data_old producer: ` + e.Error())
		}

		kind := descriptor.kind
		if _, dup := producersMap[kind]; dup {
			panic(`data_old ` + kind + ` is duplicated.`)
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

// Coder return implementation for specified data_old Descriptor
func (i *Manager) Coder(descriptor *Descriptor) Serializable {
	return newDescriptorCodec(
		func(w io.Writer) error {
			c := &container{
				Kind: descriptor.kind,
			}

			var e error
			if c.Payload, e = Backup(i.innerCoder(descriptor.value)); e != nil {
				return fmt.Errorf(`encode container payload error: %w`, e)
			}

			// Encode container
			if e = i.outerCoder(c).Write(w); e != nil {
				return fmt.Errorf(`encode container error: %w`, e)
			}

			return nil
		},
		func(r io.Reader) error {
			c := new(container)

			if e := i.outerCoder(c).Read(r); e != nil {
				return fmt.Errorf(`decode container error: %w`, e)
			}

			// Note: Setup not applied if descriptor restored from io.Reader
			newDescriptor, e := i.New(c.Kind)
			if e != nil {
				return e
			}

			if e = Restore(i.innerCoder(newDescriptor.value), c.Payload); e != nil {
				return fmt.Errorf(`decode container payload error: %w`, e)
			}

			descriptor.kind = newDescriptor.kind
			descriptor.value = newDescriptor.value

			return nil
		},
	)
}

// New return data_old descriptor by kind. For data_old entity applied passed optional setup.
func (i *Manager) New(kind string, setup ...Setup) (*Descriptor, error) {
	if producer, found := i.producersMap[kind]; found {
		return producer(setup...)
	}

	return nil, ErrNotSupported
}
