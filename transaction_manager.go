package transaction

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/vedga/lib-go-transaction/data"
)

type (
	// Manager implementation
	Manager struct {
		newID       func() string
		taskOptions []data.Option
		txManager   *TaskManager
		taskManager *TaskManager
	}

	// Option is Manager configuration modifier
	Option func(*Manager)
)

const (
	// TxKind is default transaction container type
	TxKind = `1.0`
)

// NewManager return new transaction manager implementation
func NewManager(options ...Option) *Manager {
	i := &Manager{
		newID: func() string {
			return uuid.New().String()
		},
	}

	// Apply options
	for _, opt := range options {
		opt(i)
	}

	// Task manager for transaction itself
	i.txManager = NewTaskManager(
		WithTaskProducer(TxKind, func(setup ...data.Setup) (Task, error) {
			producer := data.NewProducer[implementation]()

			tx, e := producer(
				append([]data.Setup{
					withConstructor(i.newID()),
					withTransactionManager(i),
				}, setup...)...,
			)
			if e != nil {
				return nil, e
			}

			return data.As[Task](tx)
		}),
	)

	i.taskManager = NewTaskManager(i.taskOptions...)

	return i
}

// WithTxIDProducer apply custom transaction ID producer
func WithTxIDProducer(producer func() string) Option {
	return func(i *Manager) {
		i.newID = producer
	}
}

// WithTxTaskProducer add task producer
func WithTxTaskProducer(kind string, producer TaskProducer) Option {
	return func(i *Manager) {
		i.taskOptions = append(i.taskOptions, WithTaskProducer(kind, producer))
	}
}

// RestoreCheckRetry restore transaction and check if retry limit exceed
func (i *Manager) RestoreCheckRetry(backup data.Bytes, retryTaskError *RetryTaskError, setup ...data.Setup) (Transaction, error) {
	_, tx, e := i.Decode(backup, setup...)
	if e != nil {
		return nil, e
	}

	// Check retry attempt.
	// Note: this operation increase transaction internal retry counter
	if e = tx.NextAttempt(retryTaskError.maxRetries); e != nil {
		return nil, fmt.Errorf(`%w: max %d`, e, retryTaskError.maxRetries)
	}

	return tx, nil
}

// Encode transaction context to the byte sequence
func (i *Manager) Encode(kind string, tx Transaction) (data.Bytes, error) {
	return i.txManager.EncodeTask(kind, tx)
}

// Decode bytes sequence to the transaction context
func (i *Manager) Decode(source data.Bytes, setup ...data.Setup) (string, Transaction, error) {
	kind, task, e := i.txManager.DecodeTask(source, setup...)
	if e != nil {
		return kind, nil, e
	}

	var tx Transaction
	if tx, e = data.As[Transaction](task); e != nil {
		return kind, nil, e
	}

	return kind, tx, nil
}

// New return new transaction
func (i *Manager) New(setup ...data.Setup) (tx Transaction) {
	task, e := i.txManager.NewTask(TxKind, setup...)
	if e != nil {
		panic(fmt.Errorf(`unexpected transaction builder error: %v`, e))
	}

	if tx, e = data.As[Transaction](task); e != nil {
		panic(fmt.Errorf(`unexpected transaction producer error: %v`, e))
	}

	return tx
}

// EncodeTask encode task context to the byte sequence
func (i *Manager) EncodeTask(kind string, task Task) (data.Bytes, error) {
	return i.taskManager.EncodeTask(kind, task)
}

// DecodeTask bytes sequence to the task context
func (i *Manager) DecodeTask(source data.Bytes, setup ...data.Setup) (string, Task, error) {
	return i.taskManager.DecodeTask(source, setup...)
}

// NewTask return new task container
func (i *Manager) NewTask(kind string, setup ...data.Setup) (Task, error) {
	return i.taskManager.NewTask(kind, setup...)
}
