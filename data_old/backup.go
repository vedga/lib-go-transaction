package data_old

import (
	"bytes"
	"io"
)

type (
	// Serializable interface allow object to save and load by readers
	Serializable interface {
		Write(w io.Writer) error
		Read(r io.Reader) error
	}
)

// Backup Serializable object as Raw data_old type
func Backup(o Serializable) (Raw, error) {
	b := new(bytes.Buffer)
	if e := o.Write(b); e != nil {
		return nil, e
	}

	return b.Bytes(), nil
}

// Restore Serializable object from Raw data_old type
func Restore(o Serializable, s Raw) error {
	return o.Read(bytes.NewBuffer(s))
}
