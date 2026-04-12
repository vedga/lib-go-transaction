package transaction

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vedga/lib-go-transaction/data"
)

type (
	taskA struct {
		Int int
	}
	taskB struct {
		String string
	}
)

func (i *taskA) Run(_ context.Context, _ Transaction) error {
	return nil
}

func (i *taskB) Run(_ context.Context, _ Transaction) error {
	return nil
}

func TestManager_Transaction(t *testing.T) {
	const (
		kindA = `taskA`
		kindB = `taskB`
	)

	type (
		setup struct {
			taskProducers data.Producers
			options       []data.Option
		}
		args struct {
			setup func(tx Transaction)
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
			name: "success",
			setup: setup{
				taskProducers: data.Producers{
					func(setup ...data.Setup) (*data.Descriptor, error) {
						return data.NewDescriptor[taskA](kindA, setup...)
					},
					func(setup ...data.Setup) (*data.Descriptor, error) {
						return data.NewDescriptor[taskB](kindB, setup...)
					},
				},
			},
			args: args{
				setup: func(tx Transaction) {
					_ = tx.AddTask(kindA)
					_ = tx.AddTask(kindB)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewManager(tt.setup.taskProducers, tt.setup.options...)

			tx := i.New()

			tt.args.setup(tx)

			b := &bytes.Buffer{}
			e := i.Write(b, tx)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.writeError)
			})

			var got Transaction
			got, e = i.Read(b)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.readError)
			})

			assert.Equal(t, tx, got)
		})
	}
}
