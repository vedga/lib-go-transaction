package transaction

import (
	"bytes"
	"fmt"
	"io"

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
	descriptor, e := i.txManager.Read(bytes.NewReader(backup))
	if e != nil {
		return nil, fmt.Errorf(`restore transaction descriptor error: %w`, e)
	}

	return data.DescriptorValue[Transaction](descriptor)
}

// Write transaction to io.Writer
func (i *Manager) Write(w io.Writer, tx Transaction) error {
	// Create new data descriptor with latest transaction manager version
	descriptor, e := i.newDescriptor(tx)
	if e != nil {
		return e
	}

	return i.txManager.Write(w, descriptor)
}

func (i *Manager) newDescriptor(tx Transaction) (*data.Descriptor, error) {
	descriptor, e := i.txManager.New(kind, withClone(tx))
	if e != nil {
		return nil, fmt.Errorf(`create transaction descriptor error: %w`, e)
	}

	return descriptor, nil
}

// Read transaction from io.Reader
func (i *Manager) Read(r io.Reader) (Transaction, error) {
	descriptor, e := i.txManager.Read(r)
	if e != nil {
		return nil, fmt.Errorf(`read transaction descriptor error: %w`, e)
	}

	return data.DescriptorValue[Transaction](descriptor)
}

// New return new transaction
func (i *Manager) New(setup ...data.Setup) Transaction {
	descriptor, e := i.txManager.New(kind, setup...)
	if e != nil {
		panic(fmt.Errorf(`create transaction descriptor error: %w`, e))
	}

	var tx Transaction
	if tx, e = data.DescriptorValue[Transaction](descriptor); e != nil {
		panic(fmt.Errorf(`unexpected transaction descriptor error: %w`, e))
	}

	return tx
}

// WriteTask task to io.Writer
func (i *Manager) WriteTask(w io.Writer, taskDescriptor *data.Descriptor) error {
	return i.taskManager.Write(w, taskDescriptor)
}

// ReadTask return Task from io.Reader
func (i *Manager) ReadTask(r io.Reader) (*data.Descriptor, error) {
	return i.taskManager.Read(r)
}

// NewTask return new task container
func (i *Manager) NewTask(kind string, setup ...data.Setup) (*data.Descriptor, error) {
	return i.taskManager.New(kind, setup...)
}
