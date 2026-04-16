package data_old

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_Data(t *testing.T) {
	t.Parallel()

	const (
		kind = "kind"
	)

	type (
		value struct {
			Int    int
			String string
			UInt   uint
		}

		setup struct {
			producers Producers
			options   []Option
		}
		args struct {
			descriptor *Descriptor
		}
	)
	tests := []struct {
		name       string
		setup      setup
		args       args
		writeError error
		readError  error
	}{
		{
			name: "write",
			setup: setup{
				producers: Producers{
					func(setup ...Setup) (*Descriptor, error) {
						return NewDescriptor[value](kind, setup...)
					},
				},
			},
			args: args{
				descriptor: &Descriptor{
					kind: kind,
					value: &value{
						Int:    -1,
						String: "hello",
						UInt:   1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := NewManager(tt.setup.producers, tt.setup.options...)

			raw, e := Backup(i.Coder(tt.args.descriptor))
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.writeError)
			})

			var got Descriptor
			e = Restore(i.Coder(&got), raw)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.readError)
			})

			assert.Equal(t, tt.args.descriptor, &got)
		})
	}
}
