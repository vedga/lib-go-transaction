package data

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataReadWrite(t *testing.T) {
	t.Parallel()

	type (
		args struct {
			producer Producer
			setup    []Setup
		}

		typeA struct{}
		typeB struct {
			FieldInt int
		}
		typeC struct {
			fieldInt    int
			FieldString string
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
		wantErr   error
		wantKind  string
		updater   func(o Serializable) error
		expected  func() any
		converter func(t *testing.T, o Serializable) any
	}{
		{
			name: "Int value",
			args: args{
				producer: NewProducer[int](kindInt),
				setup: []Setup{
					func(o any) error {
						v := o.(*int)
						*v = -123

						return nil
					},
				},
			},
			wantErr:  nil,
			wantKind: kindInt,
			updater: func(o Serializable) error {
				v, e := Ref[int](o)
				if e != nil {
					return e
				}

				*v++

				return nil
			},
			expected: func() any {
				return &[]int{-122}[0]
			},
			converter: func(t *testing.T, o Serializable) any {
				v, e := Ref[int](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Structure with exportable and non exportable fields with setup and update",
			args: args{
				producer: NewProducer[typeC](kindC),
				setup: []Setup{
					func(o any) error {
						v := o.(*typeC)
						v.FieldString = kindC

						return nil
					},
				},
			},
			wantErr:  nil,
			wantKind: kindC,
			updater: func(o Serializable) error {
				v, e := Ref[typeC](o)
				if e != nil {
					return e
				}

				v.fieldInt = 55
				v.FieldString = kindA + kindB + kindC

				return nil
			},
			expected: func() any {
				return &typeC{
					FieldString: kindA + kindB + kindC,
				}
			},
			converter: func(t *testing.T, o Serializable) any {
				v, e := Ref[typeC](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Structure with exportable fields with setup",
			args: args{
				producer: NewProducer[typeB](kindB),
				setup: []Setup{
					func(o any) error {
						v := o.(*typeB)
						v.FieldInt = 42

						return nil
					},
				},
			},
			wantErr:  nil,
			wantKind: kindB,
			updater: func(o Serializable) error {
				_, e := Ref[typeB](o)

				return e
			},
			expected: func() any {
				return &typeB{
					FieldInt: 42,
				}
			},
			converter: func(t *testing.T, o Serializable) any {
				v, e := Ref[typeB](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Empty structure w/o setup",
			args: args{
				producer: NewProducer[typeA](kindA),
			},
			wantErr:  nil,
			wantKind: kindA,
			updater: func(o Serializable) error {
				_, e := Ref[typeA](o)

				return e
			},
			expected: func() any {
				return &typeA{}
			},
			converter: func(t *testing.T, o Serializable) any {
				v, e := Ref[typeA](o)
				require.NoError(t, e)

				return v
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check producer with setup
			i, e := tt.args.producer(tt.args.setup...)
			require.Condition(t, func() bool {
				return errors.Is(e, tt.wantErr)
			})
			if e != nil {
				return
			}

			require.Equal(t, tt.wantKind, i.Kind())

			// Update content
			e = tt.updater(i)
			require.NoError(t, e)

			// Check write operation
			buf := new(bytes.Buffer)
			e = i.Write(buf)
			require.NoError(t, e)

			// Produce entity w/o setup
			i, e = tt.args.producer()
			require.NoError(t, e)

			// Check read operation
			e = i.Read(buf)
			require.NoError(t, e)

			expected := tt.expected()

			got := tt.converter(t, i)

			require.Equal(t, expected, got)
		})
	}
}
