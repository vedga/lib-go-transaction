package data

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataReadWrite(t *testing.T) {
	// TODO: Rewrite data test
	t.Parallel()

	type (
		args struct {
			producer Producer
			codec    Codec
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
		updater   func(o any) error
		expected  func() any
		converter func(t *testing.T, o any) any
	}{
		{
			name: "Int value",
			args: args{
				producer: NewProducer[int](),
				codec:    NewCodecJSON(),
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
			updater: func(o any) error {
				v, e := As[*int](o)
				if e != nil {
					return e
				}

				*v++

				return nil
			},
			expected: func() any {
				return &[]int{-122}[0]
			},
			converter: func(t *testing.T, o any) any {
				v, e := As[*int](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Structure with exportable and non exportable fields with setup and update",
			args: args{
				producer: NewProducer[typeC](),
				codec:    NewCodecJSON(),
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
			updater: func(o any) error {
				v, e := As[typeC](o)
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
			converter: func(t *testing.T, o any) any {
				v, e := As[typeC](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Structure with exportable fields with setup",
			args: args{
				producer: NewProducer[typeB](),
				codec:    NewCodecJSON(),
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
			updater: func(o any) error {
				_, e := As[typeB](o)

				return e
			},
			expected: func() any {
				return &typeB{
					FieldInt: 42,
				}
			},
			converter: func(t *testing.T, o any) any {
				v, e := As[typeB](o)
				require.NoError(t, e)

				return v
			},
		},
		//
		{
			name: "Empty structure w/o setup",
			args: args{
				producer: NewProducer[typeA](),
				codec:    NewCodecJSON(),
			},
			wantErr:  nil,
			wantKind: kindA,
			updater: func(o any) error {
				_, e := As[typeA](o)

				return e
			},
			expected: func() any {
				return &typeA{}
			},
			converter: func(t *testing.T, o any) any {
				v, e := As[typeA](o)
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

			//			require.Equal(t, tt.wantKind, i.Kind())

			// Update content
			e = tt.updater(i)
			require.NoError(t, e)

			/*
				// Check write operation
				buf := new(bytes.Buffer)
				e = i.Write(buf, tt.args.codec)
				require.NoError(t, e)

				// Produce entity w/o setup
				i, e = tt.args.producer()
				require.NoError(t, e)

				// Check read operation
				e = i.Read(buf, tt.args.codec)
				require.NoError(t, e)

				expected := tt.expected()

				got := tt.converter(t, i)

				require.Equal(t, expected, got)
			*/
		})
	}
}
