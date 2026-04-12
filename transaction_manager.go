package transaction

import (
	"fmt"
	"io"

	"github.com/vedga/lib-go-transaction/data"
)

type (
	// Manager implementation
	Manager struct {
		txManager   *TaskManager
		taskManager *TaskManager
	}
)

const (
	kind = `1.0`
)

// NewManager return new transaction manager implementation
func NewManager(taskProducers data.Producers, options ...data.Option) *Manager {
	i := &Manager{
		taskManager: NewTaskManager(taskProducers, options...),
	}

	// Task manager for transaction itself
	i.txManager = NewTaskManager(data.Producers{
		func(setup ...data.Setup) (*data.Descriptor, error) {
			// Set manager before execute all other setup commands
			return data.NewDescriptor[implementation](kind,
				append([]data.Setup{
					withTransactionManager(i),
				}, setup...)...)
		},
	})

	return i
}

// Write transaction to io.Writer
func (i *Manager) Write(w io.Writer, tx Transaction) error {
	// Create new data descriptor with latest transaction manager version
	descriptor, e := i.txManager.New(kind, withClone(tx))
	if e != nil {
		return fmt.Errorf(`create transaction descriptor error: %w`, e)
	}

	return i.txManager.Write(w, descriptor)
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
func (i *Manager) New() Transaction {
	descriptor, e := i.txManager.New(kind)
	if e != nil {
		panic(fmt.Errorf(`create transaction descriptor error: %w`, e))
	}

	var tx Transaction
	if tx, e = data.DescriptorValue[Transaction](descriptor); e != nil {
		panic(fmt.Errorf(`unexpected transaction descriptor error: %w`, e))
	}

	return tx
}

// NewTask return new task container
func (i *Manager) NewTask(kind string, setup ...data.Setup) (*data.Container, error) {
	descriptor, e := i.taskManager.New(kind, setup...)
	if e != nil {
		return nil, fmt.Errorf(`create task descriptor error: %w`, e)
	}

	return i.taskManager.NewContainer(descriptor)
}
