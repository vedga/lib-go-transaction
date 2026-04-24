package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
	mock "github.com/vedga/lib-go-transaction/mock"
)

func TestManager_Transaction(t *testing.T) {
	t.Parallel()

	type (
		taskA struct {
			*mock.MockTask
			Int int
		}
		taskB struct {
			*mock.MockTask
			String string
		}

		args struct {
			options []transaction.Option
			setup   func(tx transaction.Transaction)
		}
	)

	const (
		kindA = `taskA`
		kindB = `taskB`
	)

	tests := []struct {
		name       string
		args       args
		writeError error
		readError  error
	}{
		{
			name: "Write and read transaction",
			args: args{
				options: []transaction.Option{
					// Task A
					transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskA]()

						task, e := producer(setup...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
					// Task B
					transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskB]()

						task, e := producer(setup...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
				},
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

			i := transaction.NewManager(tt.args.options...)

			tx := i.New()

			tt.args.setup(tx)

			backup, e := tx.Encode()
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.writeError)
			})

			var got transaction.Transaction
			_, got, e = i.Decode(backup)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.readError)
			})

			assert.Equal(t, tx, got)

			// Backup transaction
			backup, e = got.Encode()
			assert.NoError(t, e)

			// Modify transaction
			e = got.MarkRollback("custom cause")
			assert.NoError(t, e)

			assert.NotEqual(t, tx, got)

			// Restore transaction
			_, got, e = i.Decode(backup)
			assert.NoError(t, e)

			assert.Equal(t, tx, got)
		})
	}
}
