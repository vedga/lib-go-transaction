package data

import "io"

type (
	// Raw data format
	Raw = []byte

	// Serializable interface allow object to save and load by readers
	Serializable interface {
		Kind() string
		Write(w io.Writer) error
		Read(r io.Reader) error
	}

	// Setup function declaration
	Setup func(o any) error

	// Producer function for serializable objects
	Producer func(setup ...Setup) (Serializable, error)
)
