package buddha

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
	mock "github.com/vedga/lib-go-transaction/mock"
	"go.uber.org/mock/gomock"
)

func TestBudda(t *testing.T) {
	t.Parallel()

	type (
		taskA struct {
			*mock.MockTask
			Field int
		}

		args struct {
			rebirth func(ctx context.Context, transaction transaction.Transaction) error
		}
	)

	const (
		kindA = `taskA`
	)

	tests := []struct {
		name          string
		args          args
		simulatorA    func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction)
		simulatorB    func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction)
		wantMigration bool
	}{
		{
			name: "Task is not complete before shutdown",
			args: args{},
			simulatorA: func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
					// Initialization difference:
					// instead of transaction.WithTxTaskProducer(...) use buddha.WithTxTaskProducer(...)
					WithTxTaskProducer(kindA, samsara, func(setup ...data.Setup) (transaction.Task, error) {
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
				)

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				gomock.InOrder(
					// Задача запускается на Node A
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							assert.Equal(t, uint(0), tx.Attempt())
							assert.Equal(t, false, tx.Rollback())

							// Only for test purpose!
							wg.Done()

							// For test purpose stop until context cancelled
							<-ctx.Done()

							// Task not complete, migrate request
							return transaction.ErrMigrate
						}),
				)

				return manager, tx
			},
			simulatorB: func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
					// Initialization difference:
					// instead of transaction.WithTxTaskProducer(...) use buddha.WithTxTaskProducer(...)
					WithTxTaskProducer(kindA, samsara, func(setup ...data.Setup) (transaction.Task, error) {
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
				)

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				gomock.InOrder(
					// Задача запускается на Node A
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							assert.Equal(t, uint(0), tx.Attempt())
							assert.Equal(t, false, tx.Rollback())

							// Task finished in the node B
							wg.Done()

							return nil
						}),
				)

				return manager, tx
			},
			wantMigration: true,
		},
		//
		{
			name: "Task complete before shutdown",
			args: args{},
			simulatorA: func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction) {
				ta := mock.NewMockTask(mc)

				manager := transaction.NewManager(
					// Task A
					// Initialization difference:
					// instead of transaction.WithTxTaskProducer(...) use buddha.WithTxTaskProducer(...)
					WithTxTaskProducer(kindA, samsara, func(setup ...data.Setup) (transaction.Task, error) {
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
				)

				tx := manager.New()

				e := tx.AddTask(kindA)
				assert.NoError(t, e)

				gomock.InOrder(
					// Задача запускается на Node A
					ta.
						EXPECT().
						Run(
							gomock.Any(),
							gomock.Any(),
							gomock.Any(),
						).
						DoAndReturn(func(ctx context.Context, _ string, tx transaction.Transaction) error {
							assert.Equal(t, uint(0), tx.Attempt())
							assert.Equal(t, false, tx.Rollback())

							// Only for test purpose!
							wg.Done()

							// For test purpose stop until context cancelled
							<-ctx.Done()

							return nil
						}),
				)

				return manager, tx
			},
			simulatorB: func(t *testing.T, mc *gomock.Controller, samsara *Samsara, wg *sync.WaitGroup) (*transaction.Manager, transaction.Transaction) {
				assert.Fail(t, "should not have been called")
				return nil, nil
			},
			wantMigration: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := gomock.NewController(t)
			defer mc.Finish()

			wgMigrationExpected := new(sync.WaitGroup)

			if tt.wantMigration {
				// Migration expected in this test case
				wgMigrationExpected.Add(1)
			}

			var migrationData data.Bytes

			samsaraNodeA := New(func(_ context.Context, tx transaction.Transaction) error {
				// Prepare migration data
				var e error
				migrationData, e = tx.Encode()
				assert.NoError(t, e)

				// Migration complete
				wgMigrationExpected.Done()

				return nil
			})

			wgTaskStarted := new(sync.WaitGroup)

			managerNodeA, txNodeA := tt.simulatorA(t, mc, samsaraNodeA, wgTaskStarted)

			encodedA, e := txNodeA.Encode()
			assert.NoError(t, e)

			// Wait until task start running. Only for test case!
			wgTaskStarted.Add(1)

			txNodeA, e = managerNodeA.Run(context.TODO(), encodedA)
			assert.NoError(t, e)
			assert.Nil(t, txNodeA)

			// Wait until task start running. Only for test case!
			wgTaskStarted.Wait()

			// Simulate shutdown
			e = samsaraNodeA.Close()
			assert.NoError(t, e)

			// Wait until migration complete if necessary
			wgTaskStarted.Wait()

			if !tt.wantMigration {
				// No migration
				assert.Empty(t, migrationData)
				return
			}

			// Check presence migration data
			assert.NotEmpty(t, migrationData)

			// At this point we can send migrationData via message bus to the other node.

			samsaraNodeB := New(func(_ context.Context, _ transaction.Transaction) error {
				return nil
			})

			wgTaskNodeB := new(sync.WaitGroup)

			// Wait task completion in the node B
			wgTaskNodeB.Add(1)

			managerNodeB, txNodeB := tt.simulatorB(t, mc, samsaraNodeB, wgTaskNodeB)

			var encodedB data.Bytes
			encodedB, e = txNodeB.Encode()
			assert.NoError(t, e)

			txNodeB, e = managerNodeB.Run(context.TODO(), encodedB)
			assert.NoError(t, e)
			assert.Nil(t, txNodeB)

			// Wait until task complete at node B
			wgTaskNodeB.Wait()
		})
	}
}
