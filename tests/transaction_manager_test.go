package tests

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
	mock "github.com/vedga/lib-go-transaction/mock"
)

func TestManager_Transaction(t *testing.T) {
	t.Parallel()

	const (
		kindA = `taskA`
		kindB = `taskB`
	)

	type (
		taskA struct {
			*mock.MockTask
			Int int
		}
		taskB struct {
			*mock.MockTask
			String string
		}

		setup struct {
			taskProducers data.Producers
			options       []transaction.Option
		}
		args struct {
			setup func(tx transaction.Transaction)
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
			name: "Write and read transaction",
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
				setup: func(tx transaction.Transaction) {
					_ = tx.AddTask(kindA)
					_ = tx.AddTask(kindB)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := transaction.NewManager(tt.setup.taskProducers, tt.setup.options...)

			tx := i.New()

			tt.args.setup(tx)

			b := &bytes.Buffer{}
			e := i.Write(b, tx)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.writeError)
			})

			var got transaction.Transaction
			got, e = i.Read(b)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.readError)
			})

			assert.Equal(t, tx, got)

			// Backup transaction
			var backup data.Raw
			backup, e = got.Backup()
			assert.NoError(t, e)

			// Modify transaction
			e = got.SetRollback()
			assert.NoError(t, e)

			assert.NotEqual(t, tx, got)

			// Restore transaction
			got, e = i.Restore(backup)
			assert.NoError(t, e)

			assert.Equal(t, tx, got)
		})
	}
}
