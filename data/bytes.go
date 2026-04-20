package data

import "bytes"

// NewBytesReaderWriter return io.Reader and io.Writer for bytes sequence implementation
func NewBytesReaderWriter(source Bytes) *bytes.Buffer {
	return bytes.NewBuffer(source)
}
