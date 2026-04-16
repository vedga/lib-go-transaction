package transaction

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/vedga/lib-go-transaction/data_old"
)

type (
	// Manager implementation
	Manager struct {
		newID       func() string
		dataOptions []data_old.Option
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
func NewManager(taskProducers data_old.Producers, options ...Option) *Manager {
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
	i.txManager = NewTaskManager(data_old.Producers{
		func(setup ...data_old.Setup) (*data_old.Descriptor, error) {
			// Set manager before execute all other setup commands
			return data_old.NewDescriptor[implementation](kind,
				append([]data_old.Setup{
					withConstructor(i.newID()),
					withTransactionManager(i),
				}, setup...)...)
		},
	})

	return i
}

// WithOuterCoder apply outer coder
func WithOuterCoder(coder data_old.Coder) Option {
	return func(i *Manager) {
		i.dataOptions = append(i.dataOptions, data_old.WithOuterCoder(coder))
	}
}

// WithInnerCoder apply inner coder
func WithInnerCoder(coder data_old.Coder) Option {
	return func(i *Manager) {
		i.dataOptions = append(i.dataOptions, data_old.WithInnerCoder(coder))
	}
}

// WithTxIDProducer apply custom transaction ID producer
func WithTxIDProducer(producer func() string) Option {
	return func(i *Manager) {
		i.newID = producer
	}
}

// Encode transaction encoding
func (i *Manager) Encode(txDescriptor *data_old.Descriptor) (data_old.Raw, error) {
	return data_old.Backup(i.Coder(txDescriptor))
}

// RestoreCheckRetry restore transaction and check if retry limit exceed
func (i *Manager) RestoreCheckRetry(backup data_old.Raw, retryTaskError *RetryTaskError) (Transaction, error) {
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
func (i *Manager) Restore(backup data_old.Raw) (Transaction, error) {
	var descriptor data_old.Descriptor
	if e := data_old.Restore(i.Coder(&descriptor), backup); e != nil {
		return nil, e
	}

	return data_old.DescriptorValue[Transaction](&descriptor)
}

// Coder return implementation for specified transaction data_old Descriptor
func (i *Manager) Coder(descriptor *data_old.Descriptor) data_old.Serializable {
	return i.txManager.Coder(descriptor)
}

// New return new transaction
func (i *Manager) New(setup ...data_old.Setup) Transaction {
	descriptor := i.NewTxDescriptor(setup...)

	tx, e := data_old.DescriptorValue[Transaction](descriptor)
	if e != nil {
		panic(fmt.Errorf(`unexpected transaction descriptor error: %w`, e))
	}

	return tx
}

// NewTxDescriptor return new transaction descriptor
func (i *Manager) NewTxDescriptor(setup ...data_old.Setup) *data_old.Descriptor {
	descriptor, e := i.txManager.New(kind, setup...)
	if e != nil {
		panic(fmt.Errorf(`create transaction descriptor error: %w`, e))
	}

	return descriptor
}

// EncodeTask perform task encoding
func (i *Manager) EncodeTask(taskDescriptor *data_old.Descriptor) (data_old.Raw, error) {
	return data_old.Backup(i.TaskCoder(taskDescriptor))
}

// RestoreTask task from backup
func (i *Manager) RestoreTask(backup data_old.Raw) (Task, error) {
	var descriptor data_old.Descriptor
	if e := data_old.Restore(i.TaskCoder(&descriptor), backup); e != nil {
		return nil, e
	}

	return data_old.DescriptorValue[Task](&descriptor)
}

// TaskCoder return implementation for specified task data_old Descriptor
func (i *Manager) TaskCoder(descriptor *data_old.Descriptor) data_old.Serializable {
	return i.taskManager.Coder(descriptor)
}

// NewTask return new task container
func (i *Manager) NewTask(kind string, setup ...data_old.Setup) (*data_old.Descriptor, error) {
	return i.taskManager.New(kind, setup...)
}
