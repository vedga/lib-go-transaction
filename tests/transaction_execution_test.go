package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
	mock "github.com/vedga/lib-go-transaction/mock"
	"go.uber.org/mock/gomock"
)

func TestTransactionProcessing(t *testing.T) {
	t.Parallel()

	type (
		taskA struct {
			*mock.MockTask
			Field int
		}
		taskB struct {
			*mock.MockTask
		}

		args struct {
			newTransaction func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes)
		}
	)

	const (
		kindA = `taskA`
		kindB = `taskB`
	)

	ctx := context.Background()

	errUnexpected := errors.New("unexpected error")

	// All transaction is unsupported with this manager
	unsupported := transaction.NewManager()

	tests := []struct {
		name    string
		args    args
		wantErr error
		wantTx  func(t *testing.T, m *transaction.Manager, txOriginal transaction.Transaction) transaction.Transaction
	}{
		{
			name: "Unexpected task error",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, tx transaction.Transaction) error {
							// Add rollback operations
							task, taskErr := tx.NewTask(kindA, data.NewSetup[taskA](func(o *taskA) error {
								// Additional setup for taskA
								o.Field = -1234
								return nil
							}))
							assert.NoError(t, taskErr)

							taskErr = tx.QueueRollbackTask(kindA, task)
							assert.NoError(t, taskErr)

							// Task return unexpected errors
							return errUnexpected
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: errUnexpected,
			wantTx: func(_ *testing.T, _ *transaction.Manager, _ transaction.Transaction) transaction.Transaction {
				return nil
			},
		},
		//
		{
			name: "Transaction outbox pattern",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, tx transaction.Transaction) error {
							// Add rollback operations
							task, taskErr := tx.NewTask(kindA, data.NewSetup[taskA](func(o *taskA) error {
								// Additional setup for taskA
								o.Field = -1234
								return nil
							}))
							assert.NoError(t, taskErr)

							taskErr = tx.QueueRollbackTask(kindA, task)
							assert.NoError(t, taskErr)

							var encodedTx data.Bytes
							encodedTx, taskErr = tx.Encode()
							assert.NoError(t, taskErr)

							// Verify encoded transaction to store at the database
							newTx := m.New(
								// Must be same transaction ID
								transaction.WithTransactionID(tx.ID()),
							)

							// Pending only task B
							taskErr = newTx.AddTask(kindB)
							assert.NoError(t, taskErr)

							// But new task must be in the rollback stack
							task, taskErr = m.NewTask(kindA, data.NewSetup[taskA](func(o *taskA) error {
								// Additional setup for taskA
								o.Field = -1234
								return nil
							}))
							assert.NoError(t, taskErr)
							taskErr = newTx.QueueRollbackTask(kindA, task)
							assert.NoError(t, taskErr)

							var outboxTx transaction.Transaction
							_, outboxTx, taskErr = m.Decode(encodedTx)
							assert.NoError(t, taskErr)

							// Check expected and outbox transactions
							assert.Equal(t, newTx, outboxTx)

							// Task use outbox pattern
							return transaction.ErrOutboxPattern
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: nil,
			wantTx: func(t *testing.T, m *transaction.Manager, txOriginal transaction.Transaction) transaction.Transaction {
				// No transaction to send because outbox pattern is used
				return nil
			},
		},
		//
		{
			name: "Execute last task in the transaction",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, _ transaction.Transaction) error {

							// Task return no errors
							return nil
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: nil,
			wantTx: func(t *testing.T, m *transaction.Manager, txOriginal transaction.Transaction) transaction.Transaction {
				return nil
			},
		},
		//
		{
			name: "Process first task with retry limit exceed.",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, _ transaction.Transaction) error {

							// Task return retry limit exceed
							return transaction.NewRetryTaskError(0)
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: transaction.ErrRetryLimitExceeded,
			wantTx: func(_ *testing.T, _ *transaction.Manager, _ transaction.Transaction) transaction.Transaction {
				// Retry limit exceed
				return nil
			},
		},
		//
		{
			name: "Process first task with retry not exceed. Return modified transaction.",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, _ transaction.Transaction) error {

							// Task return retry request
							return transaction.NewRetryTaskError(1)
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: nil,
			wantTx: func(t *testing.T, m *transaction.Manager, txOriginal transaction.Transaction) transaction.Transaction {
				tx := m.New(
					// Must be same transaction ID
					transaction.WithTransactionID(txOriginal.ID()),
				)

				e := tx.NextAttempt(1)

				// Pending task A
				e = tx.AddTask(kindA)
				assert.NoError(t, e)

				// Pending task B
				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				return tx
			},
		},
		//
		{
			name: "Process first task, it put one rollback operation. Return modified transaction.",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, data.Bytes) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						ta.EXPECT().Run(
							gomock.Any(),
							gomock.Eq(kindA),
							gomock.Any(),
						).DoAndReturn(func(_ context.Context, _ string, tx transaction.Transaction) error {
							// Add rollback operations
							task, taskErr := tx.NewTask(kindA, data.NewSetup[taskA](func(o *taskA) error {
								// Additional setup for taskA
								o.Field = -1234
								return nil
							}))
							assert.NoError(t, taskErr)

							taskErr = tx.QueueRollbackTask(kindA, task)
							assert.NoError(t, taskErr)

							// Task return no errors
							return nil
						}),
					)

					var encodedTx data.Bytes
					encodedTx, e = tx.Encode()
					assert.NoError(t, e)

					return m, encodedTx
				},
			},
			wantErr: nil,
			wantTx: func(t *testing.T, m *transaction.Manager, txOriginal transaction.Transaction) transaction.Transaction {
				tx := m.New(
					// Must be same transaction ID
					transaction.WithTransactionID(txOriginal.ID()),
				)

				// Pending only task B
				e := tx.AddTask(kindB)
				assert.NoError(t, e)

				// But new task must be in the rollback stack
				var task transaction.Task
				task, e = m.NewTask(kindA, data.NewSetup[taskA](func(o *taskA) error {
					// Additional setup for taskA
					o.Field = -1234
					return nil
				}))
				assert.NoError(t, e)
				e = tx.QueueRollbackTask(kindA, task)
				assert.NoError(t, e)

				return tx
			},
		},
		//
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := gomock.NewController(t)
			defer mc.Finish()

			// Simulate transaction
			m, encodedTx := tt.args.newTransaction(t, mc)

			// Pass to the manager which isn't support this task
			newTx, e := unsupported.Run(ctx, encodedTx)
			// Unsupported task: no errors and newTx is nil
			// ACK transaction but nothing resend
			assert.NoError(t, e)
			assert.Nil(t, newTx)

			// Process supported transaction
			newTx, e = m.Run(ctx, encodedTx)
			assert.Condition(t, func() bool {
				return errors.Is(e, tt.wantErr)
			})

			if e != nil {
				return
			}

			// Original transaction
			var tx transaction.Transaction
			_, tx, e = m.Decode(encodedTx)
			assert.NoError(t, e)

			// Check new transaction
			wantTx := tt.wantTx(t, m, tx)
			assert.Equal(t, wantTx, newTx)
		})
	}
}

func TestTransactionExecution(t *testing.T) {
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
		taskUnsupported struct {
			transaction.Task
		}

		args struct {
			newTransaction func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction)
			statuses       []error
		}
	)

	const (
		kindA           = `taskA`
		kindB           = `taskB`
		kindUnsupported = `unsupported`
	)

	unsupportedTask, initError := transaction.NewManager(
		transaction.WithTxTaskProducer(kindUnsupported, func(setup ...data.Setup) (transaction.Task, error) {
			producer := data.NewProducer[taskUnsupported]()

			task, e := producer(setup...)
			if e != nil {
				return nil, e
			}

			return data.As[transaction.Task](task)
		}),
	).NewTask(kindUnsupported)
	assert.NoError(t, initError)

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add task retry test case
		{
			name: "Execute all tasks",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.AddTask(kindA)
					assert.NoError(t, e)
					e = tx.AddTask(kindB)
					assert.NoError(t, e)

					gomock.InOrder(
						// kindA - first attempt
						ta.EXPECT().
							Run(
								gomock.Any(),
								gomock.Any(),
								gomock.Any(),
							).
							Return(transaction.NewRetryTaskError(1)),
						// kindA - second attempt
						ta.EXPECT().
							Run(
								gomock.Any(),
								gomock.Any(),
								gomock.Any(),
							).
							Return(nil),
						// kindB
						tb.EXPECT().
							Run(
								gomock.Any(),
								gomock.Any(),
								gomock.Any(),
							).
							Return(nil),
					)

					return m, tx
				},
				statuses: []error{
					// No errors indicate transaction finished or can't be processed
					transaction.ErrRetryTask,
					transaction.ErrNoAvailableTasks,
					transaction.ErrNoAvailableTasks,
				},
			},
		},
		//
		{
			name: "Unsupported transaction",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					e := tx.QueueTask(kindUnsupported, unsupportedTask)
					assert.NoError(t, e)

					gomock.InOrder()

					return m, tx
				},
				statuses: []error{
					// No errors indicate transaction finished or can't be processed
					transaction.ErrNoAvailableTasks,
				},
			},
		},
		//
		{
			name: "Empty transaction",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskA]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskA](func(o *taskA) error {
									o.MockTask = ta
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
						transaction.WithTxTaskProducer(kindB, func(setup ...data.Setup) (transaction.Task, error) {
							producer := data.NewProducer[taskB]()

							task, e := producer(append([]data.Setup{
								// Имплементация taskA
								data.NewSetup[taskB](func(o *taskB) error {
									o.MockTask = tb
									return nil
								}),
							}, setup...)...)
							if e != nil {
								return nil, e
							}

							return data.As[transaction.Task](task)
						}),
					)

					tx := m.New()

					gomock.InOrder()

					return m, tx
				},
				statuses: []error{
					// No errors indicate transaction finished or can't be processed
					transaction.ErrNoAvailableTasks,
				},
			},
		},
		//
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := gomock.NewController(t)
			defer mc.Finish()

			m, tx := tt.args.newTransaction(t, mc)

			for _, wantError := range tt.args.statuses {
				backup, e := tx.Encode()
				assert.NoError(t, e)

				e = tx.Run(context.Background(), transaction.TxKind, nil)

				assert.Condition(t, func() bool {
					return errors.Is(e, wantError)
				})

				var retryIndicator *transaction.RetryTaskError
				if errors.As(e, &retryIndicator) {
					// Restore transaction from backup
					tx, e = m.RestoreCheckRetry(backup, retryIndicator)
					assert.NoError(t, e)

					// Required task retry execution
					e = tx.Run(context.Background(), transaction.TxKind, nil)
					// After retry, we expect successful task execution
					assert.NoError(t, e)
				}
			}
		})
	}
}
