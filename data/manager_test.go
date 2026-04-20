package data

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
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
			content Serializable
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
		want      Serializable
	}{
		// TODO: All other case tests: unsupported type etc...
		{
			name: "success",
			args: args{
				options: []Option{
					WithProducer(NewProducer[int](kindInt)),
					WithProducer(NewProducer[typeA](kindA)),
					WithProducer(NewProducer[typeB](kindB)),
					WithProducer(NewProducer[typeC](kindC)),
				},
				content: func() Serializable {
					o, _ := NewProducer[typeB](kindB)()
					v, _ := Ref[typeB](o)
					v.FieldInt = -1234
					return o
				}(),
			},
			want: func() Serializable {
				o, _ := NewProducer[typeB](kindB)()
				v, _ := Ref[typeB](o)
				v.FieldInt = -1234
				return o
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := NewManager(tt.args.options...)

			buf := NewBytesReaderWriter(nil)
			e := i.Write(buf, tt.args.content)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.wantError)
			})
			if e != nil {
				return
			}

			var got Serializable
			got, e = i.Read(buf, tt.args.setup...)
			assert.NoError(t, e)

			assert.Equal(t, tt.want, got)
		})
	}
}
