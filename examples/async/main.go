package main

import (
	"context"

	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
)

type (
	taskA struct {
		CtxValueA int
		CtxValueB string
	}

	taskB struct {
		CtxValueA int
	}
)

const (
	kindA = "taskA"
	kindB = "taskB"
)

// Run is implementation of the transaction.Task interface for task A
func (i *taskA) Run(ctx context.Context, kind string, tx transaction.Transaction) error {
	return nil
}

// OperationTaskA return task A with custom initially context
func OperationTaskA(txManager *transaction.Manager, valueA int, valueB string) (transaction.Task, error) {
	return txManager.NewTask(kindA, func(o any) error {
		p, e := data.As[*taskA](o)
		if e != nil {
			return e
		}

		p.CtxValueA = valueA
		p.CtxValueB = valueB

		return nil
	})
}

// OperationTaskB return task B with custom initially context
func OperationTaskB(txManager *transaction.Manager, valueA int) (transaction.Task, error) {
	return txManager.NewTask(kindB, func(o any) error {
		p, e := data.As[*taskB](o)
		if e != nil {
			return e
		}

		p.CtxValueA = valueA

		return nil
	})
}

// Run is implementation of the transaction.Task interface for task B
func (i *taskB) Run(ctx context.Context, kind string, tx transaction.Transaction) error {
	return nil
}

func main() {
	// Create transaction manager
	txManager := transaction.NewManager(
		transaction.WithTxTaskProducer(kindA, func(setup ...data.Setup) (transaction.Task, error) {
			producer := data.NewProducer[taskA]()

			task, e := producer(append([]data.Setup{
				// taskA registration
				data.NewSetup[taskA](func(o *taskA) error {
					// Special constructor for task A if required
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
				// taskB registration
				data.NewSetup[taskB](func(o *taskB) error {
					// Special constructor for task B if required
					return nil
				}),
			}, setup...)...)
			if e != nil {
				return nil, e
			}

			return data.As[transaction.Task](task)
		}),
	)

	// Create new transaction
	tx := txManager.New()

	// Prepare task A with parameters
	task, e := OperationTaskA(txManager, 1234, `String value`)
	// Check error
	_ = e

	// Direct add task A to the transaction
	e = tx.QueueTask(kindA, task)
	// Check error
	_ = e

	// Prepare task B with parameters
	task, e = OperationTaskB(txManager, 789)
	// Check error
	_ = e

	var encodedTask data.Bytes
	encodedTask, e = txManager.EncodeTask(kindB, task)
	// Check error
	_ = e

	// At this point you may send encodedTask to other service for building transaction on remote side
	// Remote transaction builder can simply add task to the transaction task queue
	e = tx.QueueEncodedTask(encodedTask)
	// Check error
	_ = e

	var encodedTx data.Bytes
	encodedTx, e = tx.Encode()
	// Check error
	_ = e

	// Send encodedTx via common message bus...

	// At receiver side we can try to execute transaction by following code
	var newTx transaction.Transaction
	newTx, e = txManager.Run(context.Background(), encodedTx)
	// Check error, acknowledge message bus to processing message operation complete
	_ = e

	if newTx != nil {
		// Encode new transaction
		encodedTx, e = newTx.Encode()
		// Check error
		_ = e
		// Send new transaction via message bus
		_ = encodedTx
	}
}
