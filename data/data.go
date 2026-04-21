package data

import (
	"errors"
)

var (
	// ErrInvalidTransformation indicate invalid value transformation
	ErrInvalidTransformation = errors.New("invalid data transformation")
)

// NewProducer return new producer of entity type T
func NewProducer[T any]() Producer {
	return func(setup ...Setup) (any, error) {
		o := new(T)

		for _, fn := range setup {
			if e := fn(o); e != nil {
				return nil, e
			}
		}

		return o, nil
	}
}

// NewSetup return setup implementation for specified type
func NewSetup[T any](setup func(*T) error) Setup {
	return func(v any) error {
		if o, compatible := v.(*T); compatible {
			return setup(o)
		}

		return ErrInvalidTransformation
	}
}

// As return pointer to the value of required type or error
func As[T any](v any) (T, error) {
	if i, valid := v.(T); valid {
		return i, nil
	}

	var none T
	return none, ErrInvalidTransformation
}
