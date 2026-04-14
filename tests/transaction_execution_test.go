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

	unsupportedTask, initError := transaction.NewManager(data.Producers{
		// taskA
		func(setup ...data.Setup) (*data.Descriptor, error) {
			return data.NewDescriptor[taskUnsupported](kindUnsupported, setup...)
		},
	}).NewTask(kindUnsupported)
	assert.NoError(t, initError)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Execute all tasks",
			args: args{
				newTransaction: func(t *testing.T, mc *gomock.Controller) (*transaction.Manager, transaction.Transaction) {
					ta := mock.NewMockTask(mc)
					tb := mock.NewMockTask(mc)

					// Создаем менеджер с двумя задачами
					m := transaction.NewManager(
						data.Producers{
							// taskA
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskA](kindA,
									append(
										[]data.Setup{
											// Имплементация taskA
											data.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskB](kindB,
									append(
										[]data.Setup{
											// Имплементация taskB
											data.NewSetup[taskB](func(o *taskB) error {
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
							Return(transaction.ErrRetryTask),
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
						data.Producers{
							// taskA
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskA](kindA,
									append(
										[]data.Setup{
											// Имплементация taskA
											data.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskB](kindB,
									append(
										[]data.Setup{
											// Имплементация taskB
											data.NewSetup[taskB](func(o *taskB) error {
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
						data.Producers{
							// taskA
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskA](kindA,
									append(
										[]data.Setup{
											// Имплементация taskA
											data.NewSetup[taskA](func(o *taskA) error {
												o.MockTask = ta
												return nil
											}),
										},
										setup...,
									)...)
							},
							// taskB
							func(setup ...data.Setup) (*data.Descriptor, error) {
								return data.NewDescriptor[taskB](kindB,
									append(
										[]data.Setup{
											// Имплементация taskB
											data.NewSetup[taskB](func(o *taskB) error {
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

				if errors.Is(e, transaction.ErrRetryTask) {
					// Restore transaction from backup
					tx, e = m.Restore(backup)
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
