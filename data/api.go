package data

import (
	"io"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE

type (
	// Codec interface allow save objects of any or limited types to the io.Writer
	// and then later restore it from io.Reader
	Codec interface {
		Write(w io.Writer, o any) error
		Read(r io.Reader, o any) error
	}

	// Setup function declaration
	Setup func(o any) error

	// Producer function for serializable objects
	Producer func(setup ...Setup) (any, error)
)
