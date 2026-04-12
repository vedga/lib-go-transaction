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

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE
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
		// RollbackIndicator if true currant action is transaction rollback
		RollbackIndicator bool
		// PendingTasks contain tasks sequence for execute transaction
		PendingTasks *queue.Queue[*data.Container]
		// RollbackStack contain transaction rollback sequence
		RollbackStack *stack.Stack[*data.Container]
	}
)

var (
	// ErrContinueTransaction indicate transaction must be continued
	ErrContinueTransaction = errors.New("continue transaction")
	// ErrRetryTask indicate transaction must be retried with current top task
	ErrRetryTask = errors.New("retry task")
)

func withConstructor() data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.PendingTasks = queue.New[*data.Container](0)
		o.RollbackStack = stack.New[*data.Container](0)

		return nil
	})
}

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
			o.PendingTasks = clone.Clone(i.PendingTasks)
			o.RollbackStack = clone.Clone(i.RollbackStack)

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

	// Execute all possible tasks
	for taskContainer, present := i.PendingTasks.Peek(); present; taskContainer, present = i.PendingTasks.Peek() {
		task := i.manager.GetTask(taskContainer)
		if task == nil {
			// Most task can't be handled by this instance
			return ErrContinueTransaction
		}

		// Execute task
		e := task.Run(ctx, i)
		if errors.Is(e, ErrRetryTask) {
			// Retry transaction with current top task
			return ErrContinueTransaction
		}

		// TODO: Check if error occurred in transaction rollback state

		// Drop processed task
		_ = i.PendingTasks.Drop()

		// Any other errors cause transaction rollback
		if e != nil {
			// Rollback transaction
			i.Rollback()
		}
	}

	// All tasks in the transaction operation complete

	return nil
}

// Rollback transaction
func (i *implementation) Rollback() {
	if i.RollbackIndicator {
		// Already at rollback state
		return
	}

	i.RollbackIndicator = true

	// Create pending tasks queue
	i.PendingTasks = queue.New[*data.Container](i.RollbackStack.Size())

	// TODO: May be implement corresponded method in the queue package?
	for taskContainer, present := i.RollbackStack.Pop(); present; taskContainer, present = i.RollbackStack.Pop() {
		i.PendingTasks.Enqueue(taskContainer)
	}
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
	i.PendingTasks.Enqueue(container)
}

// NewTask return new task context at data exchange format
func (i *implementation) NewTask(kind string, setup ...data.Setup) (*data.Container, error) {
	return i.manager.NewTask(kind, setup...)
}
