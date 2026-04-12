package transaction

import (
	"context"
	"errors"
	"fmt"

	clone "github.com/huandu/go-clone/generic"
	"github.com/vedga/lib-go-transaction/data"
	"github.com/vedga/lib-go-transaction/queue"
	"github.com/vedga/lib-go-transaction/stack"
)

type (
	// Transaction interface declaration
	Transaction interface {
		Task
		AddTask(kind string, setup ...data.Setup) error
		QueueTask(container *data.Container)
		NewTask(kind string, setup ...data.Setup) (*data.Container, error)
	}

	implementation struct {
		manager *Manager
		// Tasks sequence for execute transaction
		Pending queue.Queue[*data.Container]
		// Transaction rollback sequence
		Rollback stack.Stack[*data.Container]
	}
)

func withTransactionManager(manager *Manager) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.manager = manager

		return nil
	})
}

func withClone(tx Transaction) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		if i, theSame := tx.(*implementation); theSame {
			// Clone fields
			o.Pending = clone.Clone(i.Pending)
			o.Rollback = clone.Clone(i.Rollback)

			return nil
		}

		return errors.New("only same transaction type supported yet")
	})
}

// Run transaction
func (i *implementation) Run(ctx context.Context, tx Transaction) error {
	if tx != nil {
		return errors.New("nested transactions are not supported")
	}

	return nil
}

// AddTask add task to the transaction
func (i *implementation) AddTask(kind string, setup ...data.Setup) error {
	container, e := i.NewTask(kind, setup...)
	if e != nil {
		return fmt.Errorf(`create new task error: %w`, e)
	}

	// Queue task for execution
	i.QueueTask(container)

	return nil
}

// QueueTask task for execution
func (i *implementation) QueueTask(container *data.Container) {
	i.Pending.Enqueue(container)
}

// NewTask return new task context at data exchange format
func (i *implementation) NewTask(kind string, setup ...data.Setup) (*data.Container, error) {
	return i.manager.NewTask(kind, setup...)
}
