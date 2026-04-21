package data

import "bytes"

// Encode object to the byte sequence
func Encode(codec Codec, o any) (Bytes, error) {
	b := new(bytes.Buffer)
	if e := codec.Write(b, o); e != nil {
		return nil, e
	}

	return b.Bytes(), nil
}

// Decode byte sequence to the object of specified type
func Decode(codec Codec, source Bytes, o any) error {
	b := bytes.NewBuffer(source)
	if e := codec.Read(b, o); e != nil {
		return e
	}

	return nil
}
