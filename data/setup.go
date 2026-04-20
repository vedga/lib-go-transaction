package data

// NewSetup return setup implementation for specified type
func NewSetup[T any](setup func(*T) error) Setup {
	return func(v any) error {
		if o, compatible := v.(*T); compatible {
			return setup(o)
		}

		return ErrInvalidTransformation
	}
}
