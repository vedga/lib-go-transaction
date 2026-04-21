package data

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	t.Parallel()

	type (
		typeA struct{}
		typeB struct {
			FieldInt int
		}
		typeC struct {
			fieldInt    int
			FieldString string
		}

		args struct {
			options []Option
			kind    string
			content any
			setup   []Setup
		}
	)

	const (
		kindA   = `A`
		kindB   = `B`
		kindC   = `C`
		kindInt = `int`
	)

	tests := []struct {
		name      string
		args      args
		wantError error
		want      any
	}{
		// TODO: All other case tests: unsupported type etc...
		{
			name: "success",
			args: args{
				options: []Option{
					WithProducer(kindInt, NewProducer[int]()),
					WithProducer(kindA, NewProducer[typeA]()),
					WithProducer(kindB, NewProducer[typeB]()),
					WithProducer(kindC, NewProducer[typeC]()),
				},
				setup: []Setup{},
				kind:  kindB,
				content: func() any {
					o, _ := NewProducer[typeB]()(NewSetup(func(o *typeB) error {
						o.FieldInt = 42
						return nil
					}))
					v, _ := As[*typeB](o)
					v.FieldInt = -1234
					return o
				}(),
			},
			want: func() any {
				o, _ := NewProducer[typeB]()()
				v, _ := As[*typeB](o)
				v.FieldInt = -1234
				return o
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := NewManager(tt.args.options...)

			buf, e := i.Encode(tt.args.kind, tt.args.content)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.wantError)
			})
			if e != nil {
				return
			}

			var (
				kind string
				got  any
			)
			kind, got, e = i.Decode(buf, tt.args.setup...)
			assert.NoError(t, e)

			assert.Equal(t, tt.args.kind, kind)
			assert.Equal(t, tt.want, got)
		})
	}
}
