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
		dataOptions []data.Option
		txManager   *TaskManager
		taskManager *TaskManager
	}

	// Option is Manager configuration modifier
	Option func(*Manager)
)

const (
	kind = `1.0`
)

// NewManager return new transaction manager implementation
func NewManager(taskProducers data.Producers, options ...Option) *Manager {
	i := &Manager{
		newID: func() string {
			return uuid.New().String()
		},
	}

	// Apply options
	for _, opt := range options {
		opt(i)
	}

	i.taskManager = NewTaskManager(taskProducers, i.dataOptions...)

	// Task manager for transaction itself
	i.txManager = NewTaskManager(data.Producers{
		func(setup ...data.Setup) (*data.Descriptor, error) {
			// Set manager before execute all other setup commands
			return data.NewDescriptor[implementation](kind,
				append([]data.Setup{
					withConstructor(i.newID()),
					withTransactionManager(i),
				}, setup...)...)
		},
	})

	return i
}

// WithOuterCoder apply outer coder
func WithOuterCoder(coder data.Coder) Option {
	return func(i *Manager) {
		i.dataOptions = append(i.dataOptions, data.WithOuterCoder(coder))
	}
}

// WithInnerCoder apply inner coder
func WithInnerCoder(coder data.Coder) Option {
	return func(i *Manager) {
		i.dataOptions = append(i.dataOptions, data.WithInnerCoder(coder))
	}
}

// WithTxIDProducer apply custom transaction ID producer
func WithTxIDProducer(producer func() string) Option {
	return func(i *Manager) {
		i.newID = producer
	}
}

// Encode transaction encoding
func (i *Manager) Encode(txDescriptor *data.Descriptor) (data.Raw, error) {
	return data.Backup(i.Coder(txDescriptor))
}

// RestoreCheckRetry restore transaction and check if retry limit exceed
func (i *Manager) RestoreCheckRetry(backup data.Raw, retryTaskError *RetryTaskError) (Transaction, error) {
	tx, e := i.Restore(backup)
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

// Restore transaction from backup
func (i *Manager) Restore(backup data.Raw) (Transaction, error) {
	var descriptor data.Descriptor
	if e := data.Restore(i.Coder(&descriptor), backup); e != nil {
		return nil, e
	}

	return data.DescriptorValue[Transaction](&descriptor)
}

// Coder return implementation for specified transaction data Descriptor
func (i *Manager) Coder(descriptor *data.Descriptor) data.Serializable {
	return i.txManager.Coder(descriptor)
}

// New return new transaction
func (i *Manager) New(setup ...data.Setup) Transaction {
	descriptor := i.NewTxDescriptor(setup...)

	tx, e := data.DescriptorValue[Transaction](descriptor)
	if e != nil {
		panic(fmt.Errorf(`unexpected transaction descriptor error: %w`, e))
	}

	return tx
}

// NewTxDescriptor return new transaction descriptor
func (i *Manager) NewTxDescriptor(setup ...data.Setup) *data.Descriptor {
	descriptor, e := i.txManager.New(kind, setup...)
	if e != nil {
		panic(fmt.Errorf(`create transaction descriptor error: %w`, e))
	}

	return descriptor
}

// EncodeTask perform task encoding
func (i *Manager) EncodeTask(taskDescriptor *data.Descriptor) (data.Raw, error) {
	return data.Backup(i.TaskCoder(taskDescriptor))
}

// RestoreTask task from backup
func (i *Manager) RestoreTask(backup data.Raw) (Task, error) {
	var descriptor data.Descriptor
	if e := data.Restore(i.TaskCoder(&descriptor), backup); e != nil {
		return nil, e
	}

	return data.DescriptorValue[Task](&descriptor)
}

// TaskCoder return implementation for specified task data Descriptor
func (i *Manager) TaskCoder(descriptor *data.Descriptor) data.Serializable {
	return i.taskManager.Coder(descriptor)
}

// NewTask return new task container
func (i *Manager) NewTask(kind string, setup ...data.Setup) (*data.Descriptor, error) {
	return i.taskManager.New(kind, setup...)
}
