package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data_old"
	mock "github.com/vedga/lib-go-transaction/mock"
	"go.uber.org/mock/gomock"
)

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

	unsupportedTask, initError := transaction.NewManager(data_old.Producers{
		// taskA
		func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
			return data_old.NewDescriptor[taskUnsupported](kindUnsupported, setup...)
		},
	}).NewTask(kindUnsupported)
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
						data_old.Producers{
							// taskA
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskA](kindA,
									append(
										[]data_old.Setup{
											// Имплементация taskA
											data_old.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskB](kindB,
									append(
										[]data_old.Setup{
											// Имплементация taskB
											data_old.NewSetup[taskB](func(o *taskB) error {
												o.MockTask = tb
												return nil
											}),
										},
										setup...,
									)...)
							},
						},
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
							).
							Return(transaction.NewRetryTaskError(1)),
						// kindA - second attempt
						ta.EXPECT().
							Run(
								gomock.Any(),
								gomock.Any(),
							).
							Return(nil),
						// kindB
						tb.EXPECT().
							Run(
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
					nil,
					nil,
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
						data_old.Producers{
							// taskA
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskA](kindA,
									append(
										[]data_old.Setup{
											// Имплементация taskA
											data_old.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskB](kindB,
									append(
										[]data_old.Setup{
											// Имплементация taskB
											data_old.NewSetup[taskB](func(o *taskB) error {
												o.MockTask = tb
												return nil
											}),
										},
										setup...,
									)...)
							},
						},
					)

					tx := m.New()

					tx.QueueTask(unsupportedTask)

					gomock.InOrder()

					return m, tx
				},
				statuses: []error{
					// No errors indicate transaction finished or can't be processed
					nil,
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
						data_old.Producers{
							// taskA
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskA](kindA,
									append(
										[]data_old.Setup{
											// Имплементация taskA
											data_old.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
								return data_old.NewDescriptor[taskB](kindB,
									append(
										[]data_old.Setup{
											// Имплементация taskB
											data_old.NewSetup[taskB](func(o *taskB) error {
												o.MockTask = tb
												return nil
											}),
										},
										setup...,
									)...)
							},
						},
					)

					tx := m.New()

					gomock.InOrder()

					return m, tx
				},
				statuses: []error{
					// No errors indicate transaction finished or can't be processed
					nil,
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
				backup, e := tx.Backup()
				assert.NoError(t, e)

				e = tx.Run(context.Background(), nil)

				assert.Condition(t, func() bool {
					return errors.Is(e, wantError)
				})

				var retryIndicator *transaction.RetryTaskError
				if errors.As(e, &retryIndicator) {
					// Restore transaction from backup
					tx, e = m.RestoreCheckRetry(backup, retryIndicator)
					assert.NoError(t, e)

					// Required task retry execution
					e = tx.Run(context.Background(), nil)
					// After retry, we expect successful task execution
					assert.NoError(t, e)
				}
			}
		})
	}
}
