package data

import "io"

type (
	descriptorCodec struct {
		write func(w io.Writer) error
		read  func(r io.Reader) error
	}
)

func newDescriptorCodec(write func(w io.Writer) error, read func(r io.Reader) error) Serializable {
	return &descriptorCodec{
		write: write,
		read:  read,
	}
}

// Write Descriptor to io.Writer
func (i *descriptorCodec) Write(w io.Writer) error {
	return i.write(w)
}

// Read Descriptor from io.Reader
func (i *descriptorCodec) Read(r io.Reader) error {
	return i.read(r)
}
