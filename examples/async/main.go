package main

import (
	"context"

	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
)

type (
	dataA struct {
		// ValueA started from upper case therefor it serialized as "v" JSON field
		ValueA string `json:"v"`
	}

	// taskA is task A context
	taskA struct {
		// dataManager started from lower case letter therefore it is not serialized and must be initialized via constructor
		dataManager *data.Manager
		// CtxValueA started from upper case letter therefore it serialized as "ctx_value_a" JSON field
		CtxValueA int `json:"ctx_value_a"`
		// CtxValueB started from upper case letter therefore it serialized as "CtxValueB" JSON field
		CtxValueB string
	}

	// taskB is task B context
	taskB struct {
		dataManager *data.Manager
		CtxValueA   int
	}
)

const (
	kindDataA   = "dataA"
	kindDataInt = "dataInt"
	kindTaskA   = "taskA"
	kindTaskB   = "taskB"
)

// Run is implementation of the transaction.Task interface for task A
func (i *taskA) Run(ctx context.Context, kind string, tx transaction.Transaction) error {
	o, e := i.dataManager.New(kindDataInt, func(o any) error {
		p, e := data.As[*int](o)
		if e != nil {
			return e
		}

		// Setup int value
		*p = 1234

		return nil
	})

	// Encode modified data
	var encodedDataB data.Bytes
	encodedDataB, e = i.dataManager.Encode(kindDataInt, o)

	// Allocate data of type dataA. Also, possible pass initially data setup if necessary.
	o, e = i.dataManager.New(kindDataA)
	// Check error
	_ = e

	var p *dataA
	p, e = data.As[*dataA](o)
	// Check error
	_ = e

	// Setup data values
	p.ValueA = `Some value`

	// Encode modified data
	var encodedDataA data.Bytes
	encodedDataA, e = i.dataManager.Encode(kindDataA, o)
	// Check error

	// Push to the data stack
	tx.PushData(encodedDataA, encodedDataB)

	return nil
}

// Run is implementation of the transaction.Task interface for task B
func (i *taskB) Run(ctx context.Context, kind string, tx transaction.Transaction) error {
	encodedData, present := tx.PopData()
	if present {
		// Must be encodedDataB as used stack order
		dataKind, o, e := i.dataManager.Decode(encodedData)
		// Check error
		_ = e
		// Use for type check if required
		_ = dataKind

		var d *dataA
		d, e = data.As[*dataA](o)
		// Check error
		_ = e
		// Use data
		_ = d
	}

	// Use outbox pattern: encode next transaction and write it to the some storage
	txEncoded, e := tx.Encode()
	// Check error
	_ = e
	// Write encoded transaction to the storage
	_ = txEncoded

	// Indicate this task use outbox pattern
	return transaction.ErrOutboxPattern
}

// OperationTaskA return task A with custom initially context
func OperationTaskA(txManager *transaction.Manager, valueA int, valueB string) (transaction.Task, error) {
	return txManager.NewTask(kindTaskA, func(o any) error {
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
	return txManager.NewTask(kindTaskB, func(o any) error {
		p, e := data.As[*taskB](o)
		if e != nil {
			return e
		}

		p.CtxValueA = valueA

		return nil
	})
}

func main() {
	// Create data manager and register supported data types
	dataManager := data.NewManager(
		data.WithProducer(kindDataA, data.NewProducer[dataA]()),
		data.WithProducer(kindDataInt, data.NewProducer[int]()),
	)

	// Create transaction manager
	txManager := transaction.NewManager(
		transaction.WithTxTaskProducer(kindTaskA, func(setup ...data.Setup) (transaction.Task, error) {
			producer := data.NewProducer[taskA]()

			task, e := producer(append([]data.Setup{
				// taskA registration
				data.NewSetup[taskA](func(o *taskA) error {
					// Special constructor for task A if required
					o.dataManager = dataManager

					return nil
				}),
			}, setup...)...)
			if e != nil {
				return nil, e
			}

			return data.As[transaction.Task](task)
		}),
		transaction.WithTxTaskProducer(kindTaskB, func(setup ...data.Setup) (transaction.Task, error) {
			producer := data.NewProducer[taskB]()

			task, e := producer(append([]data.Setup{
				// taskB registration
				data.NewSetup[taskB](func(o *taskB) error {
					// Special constructor for task B if required
					o.dataManager = dataManager

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
	e = tx.QueueTask(kindTaskA, task)
	// Check error
	_ = e

	// Prepare task B with parameters
	task, e = OperationTaskB(txManager, 789)
	// Check error
	_ = e

	var encodedTask data.Bytes
	encodedTask, e = txManager.EncodeTask(kindTaskB, task)
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
