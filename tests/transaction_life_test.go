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

func TestTransactionLife(t *testing.T) {
	t.Parallel()

	type (
		taskA struct {
			*mock.MockTask
		}
		taskB struct {
			*mock.MockTask
		}
		taskC struct {
			*mock.MockTask
		}
		taskD struct {
			*mock.MockTask
		}
	)

	const (
		kindA = `taskA`
		kindB = `taskB`
		kindC = `taskC`
		kindD = `taskD`
	)

	unexpectedError := errors.New("unexpected error")

	tests := []struct {
		name      string
		simulator func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction)
		wantError error
	}{
		{
			name: "Execute in order A, B, 3 retry in task A, retry limit exceeded",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					// Task A, retry #1
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return transaction.NewRetryTaskError(3)
						}),
					// Task A, retry #2
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(1), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return transaction.NewRetryTaskError(3)
						}),
					// Task A, retry #3
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(2), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return transaction.NewRetryTaskError(3)
						}),
				)

				return manager, tx
			},
			wantError: transaction.ErrRetryLimitExceeded,
		},
		//
		{
			name: "Execute in order A, B, 2 retry in task A, no errors",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					// Task A, retry #1
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return transaction.NewRetryTaskError(3)
						}),
					// Task A, retry #2
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(1), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return transaction.NewRetryTaskError(3)
						}),
					// Task A, retry #3
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(2), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return nil
						}),
					// Task B
					tb.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return nil
						}),
				)

				return manager, tx
			},
			wantError: nil,
		},
		//
		{
			name: "Execute in order A, B w/o errors",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return nil
						}),
					tb.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							_ = tx

							return nil
						}),
				)

				return manager, tx
			},
			wantError: nil,
		},
		//
		{
			name: "Execute in order A, B with possible rollbacks but w/o errors",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindB)
							assert.NoError(t, e)

							return nil
						}),
					tb.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindA)
							assert.NoError(t, e)

							return nil
						}),
				)

				return manager, tx
			},
			wantError: nil,
		},
		//
		{
			name: "Execute in order A, B with rollback on task B, but it return unexpected error",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)
				tc := mock.NewMockTask(mc)
				td := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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
					// Task C
					transaction.WithTxTaskProducer(kindC, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskC]()

						task, e := producer(append([]data.Setup{
							// Имплементация taskA
							data.NewSetup[taskC](func(o *taskC) error {
								o.MockTask = tc
								return nil
							}),
						}, setup...)...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
					// Task D
					transaction.WithTxTaskProducer(kindD, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskD]()

						task, e := producer(append([]data.Setup{
							// Имплементация taskA
							data.NewSetup[taskD](func(o *taskD) error {
								o.MockTask = td
								return nil
							}),
						}, setup...)...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
				)

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindC)
							assert.NoError(t, e)

							return nil
						}),
					tb.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindD)
							assert.NoError(t, e)

							// Initiate rollback by task B
							e = tx.Rollback()
							assert.NoError(t, e)

							return unexpectedError
						}),
				)

				return manager, tx
			},
			wantError: unexpectedError,
		},
		//
		{
			name: "Execute in order A, B with rollback on task B w/o errors",
			simulator: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)
				tb := mock.NewMockTask(mc)
				tc := mock.NewMockTask(mc)
				td := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
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
					// Task B
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
					// Task C
					transaction.WithTxTaskProducer(kindC, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskC]()

						task, e := producer(append([]data.Setup{
							// Имплементация taskA
							data.NewSetup[taskC](func(o *taskC) error {
								o.MockTask = tc
								return nil
							}),
						}, setup...)...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
					// Task D
					transaction.WithTxTaskProducer(kindD, func(setup ...data.Setup) (transaction.Task, error) {
						producer := data.NewProducer[taskD]()

						task, e := producer(append([]data.Setup{
							// Имплементация taskA
							data.NewSetup[taskD](func(o *taskD) error {
								o.MockTask = td
								return nil
							}),
						}, setup...)...)
						if e != nil {
							return nil, e
						}

						return data.As[transaction.Task](task)
					}),
				)

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				e = tx.AddTask(kindB)
				assert.NoError(t, e)

				gomock.InOrder(
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindC)
							assert.NoError(t, e)

							return nil
						}),
					tb.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, false, runCtx.Rollback())

							e := tx.AddRollbackTask(kindD)
							assert.NoError(t, e)

							// Initiate rollback by task B
							e = tx.Rollback()
							assert.NoError(t, e)

							return nil
						}),
					// Rollback task D from stack
					td.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, true, runCtx.Rollback())

							_ = tx

							return nil
						}),
					// Rollback task C from stack
					tc.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							runCtx := transaction.RunContext(ctx)
							assert.NotNil(t, runCtx)
							assert.Equal(t, uint(0), runCtx.Attempt())
							assert.Equal(t, true, runCtx.Rollback())

							_ = tx

							return nil
						}),
				)

				return manager, tx
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := gomock.NewController(t)
			defer mc.Finish()

			manager, tx := tt.simulator(t, mc)

			ixID := tx.ID()

			for tx != nil {
				encodedTx, e := tx.Encode()
				assert.NoError(t, e)

				// Execute task
				tx, e = manager.Run(context.Background(), encodedTx)

				if e != nil {
					assert.Condition(t, func() bool {
						return errors.Is(e, tt.wantError)
					})

					// Test complete by expected error
					return
				}

				if tx != nil {
					// Transaction ID must be same
					assert.Equal(t, ixID, tx.ID())
				}
			}
		})
	}
}
