package data

import (
	"io"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE

type (
	// Raw data format
	Raw = []byte

	// Codec interface allow save objects of any or limited types to the io.Writer
	// and then later restore it from io.Reader
	Codec interface {
		Write(w io.Writer, o any) error
		Read(r io.Reader, o any) error
	}

	// Serializable interface allow object to save and load by readers
	Serializable interface {
		Write(w io.Writer, codec Codec) error
		Read(r io.Reader, codec Codec) error
		Kind() string
	}

	// Setup function declaration
	Setup func(o any) error

	// Producer function for serializable objects
	Producer func(setup ...Setup) (Serializable, error)
)
