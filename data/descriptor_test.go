package data

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescriptor(t *testing.T) {
	t.Parallel()

	type (
		typeA struct {
			value int
		}
		typeB struct{}

		args struct {
			builder func() (*Descriptor, error)
		}
	)

	const (
		kindA = "A"
	)

	var (
		errSetup = errors.New("setup error")
	)

	tests := []struct {
		name      string
		args      args
		wantError error
		want      *Descriptor
	}{
		{
			name: "NewDescriptor",
			args: args{
				builder: func() (*Descriptor, error) {
					return NewDescriptor[typeA](kindA)
				},
			},
			wantError: nil,
			want: &Descriptor{
				kind:  kindA,
				value: &typeA{},
			},
		},
		{
			name: "NewDescriptor with valid setup",
			args: args{
				builder: func() (*Descriptor, error) {
					return NewDescriptor[typeA](kindA, NewSetup[typeA](func(o *typeA) error {
						o.value = 1
						return nil
					}))
				},
			},
			wantError: nil,
			want: &Descriptor{
				kind: kindA,
				value: &typeA{
					value: 1,
				},
			},
		},
		{
			name: "NewDescriptor with setup error",
			args: args{
				builder: func() (*Descriptor, error) {
					return NewDescriptor[typeA](kindA, NewSetup[typeA](func(o *typeA) error {
						o.value = 1
						return errSetup
					}))
				},
			},
			wantError: errSetup,
			want:      nil,
		},
		{
			name: "NewDescriptor with invalid setup",
			args: args{
				builder: func() (*Descriptor, error) {
					return NewDescriptor[typeA](kindA, NewSetup[typeB](func(o *typeB) error {
						_ = o
						return nil
					}))
				},
			},
			wantError: ErrInvalidSetup,
			want:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, e := tt.args.builder()
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.wantError)
			})

			assert.Equal(t, tt.want, got)
		})
	}
}
