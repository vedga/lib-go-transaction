package data

import (
	"bytes"
	"fmt"
)

type (
	// Raw data format
	Raw = []byte

	// Container for exchange information
	Container struct {
		Kind    string `json:"kind"`
		Payload Raw    `json:"payload"`
	}
)

// Restore container from backup
func Restore(raw Raw) (*Container, error) {
	c := new(Container)

	if e := jsonCoder(c).Read(bytes.NewBuffer(raw)); e != nil {
		return nil, fmt.Errorf(`read container error: %w`, e)
	}

	return c, nil
}

// Backup return container backup
func (i *Container) Backup() (Raw, error) {
	coder := jsonCoder(i)

	buf := new(bytes.Buffer)
	if e := coder.Write(buf); e != nil {
		return nil, fmt.Errorf(`container backup error: %w`, e)
	}

	return buf.Bytes(), nil
}
