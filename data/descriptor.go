package data

import (
	"encoding/json"
	"errors"
	"fmt"
)

type (
	// Raw data format
	Raw = []byte

	// container representation
	container struct {
		Kind    string `json:"kind"`
		Payload Raw    `json:"payload"`
	}

	// Descriptor implementation
	Descriptor struct {
		kind  string
		value any
	}

	// Setup function declaration
	Setup func(o any) error
)

var (
	// ErrInvalidSetup indicate invalid setup attempt
	ErrInvalidSetup = errors.New("invalid setup")
)

// NewDescriptor return new data descriptor
func NewDescriptor[T any](kind string, setup ...Setup) (*Descriptor, error) {
	v := new(T)

	for _, fn := range setup {
		if e := fn(v); e != nil {
			return nil, e
		}
	}

	return &Descriptor{
		kind:  kind,
		value: v,
	}, nil
}

// NewSetup return setup implementation for specified type
func NewSetup[T any](setup func(*T) error) Setup {
	return func(v any) error {
		o, e := Value[*T](v)
		if e == nil {
			return setup(o)
		}

		return e
	}
}

// DescriptorValue return specified data type from data descriptor
func DescriptorValue[T any](descriptor *Descriptor) (T, error) {
	return Value[T](descriptor.value)
}

// Value return value of required type or error
func Value[T any](v any) (T, error) {
	if value, valid := v.(T); valid {
		return value, nil
	}

	var none T
	return none, ErrInvalidSetup
}

// MarshalJSON implementation of json.Marshaler interface
func (i *Descriptor) MarshalJSON() ([]byte, error) {
	c := container{
		Kind: i.kind,
	}

	var e error
	if c.Payload, e = Backup(jsonCoder(i.value)); e != nil {
		return nil, fmt.Errorf(`payload encoding error: %w`, e)
	}

	return json.Marshal(c)
}

// UnmarshalJSON implementation json.Unmarshaler interface
func (i *Descriptor) UnmarshalJSON(raw []byte) error {
	c := new(container)

	if e := json.Unmarshal(raw, c); e != nil {
		return fmt.Errorf(`payload decoding error: %w`, e)
	}

	return Restore(jsonCoder(i.value), c.Payload)
}
