package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/vedga/lib-go-transaction/data"
	"github.com/vedga/lib-go-transaction/deque"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE
type (
	// Transaction interface declaration
	Transaction interface {
		Task
		AddTask(kind string, setup ...data.Setup) error
		AddRollbackTask(kind string, setup ...data.Setup) error
		QueueTask(container *data.Container)
		QueueRollbackTask(container *data.Container)
		SetRollback() error
		NewTask(kind string, setup ...data.Setup) (*data.Container, error)
		Backup() (data.Raw, error)
		NextAttempt(retryTaskError *RetryTaskError) error
	}

	implementation struct {
		manager *Manager
		// ID of the transaction
		ID string
		// RollbackIndicator if true currant action is transaction rollback
		RollbackIndicator bool
		// Attempt of task execution
		Attempt uint
		// PendingTasks contain tasks sequence for execute transaction
		PendingTasks *deque.Deque[*data.Container]
		// RollbackStack contain transaction rollback sequence
		RollbackStack *deque.Deque[*data.Container]
	}
)

func withConstructor(ID string) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.ID = ID
		o.PendingTasks = deque.New[*data.Container](0)
		o.RollbackStack = deque.New[*data.Container](0)

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
			o.ID = i.ID
			o.RollbackIndicator = i.RollbackIndicator
			o.PendingTasks = deque.Clone(i.PendingTasks)
			o.RollbackStack = deque.Clone(i.RollbackStack)

			return nil
		}

		return errors.New("only same transaction type supported yet")
	})
}

// Run transaction
// Return values:
// nil - no errors (task processed, current task not supported, transaction complete or being complete)
// not nil - last task execution status
// Predefined error ErrRetryTask indicate than transaction must be retried after some time. Suggested implementation
// is backup transaction before calling Run() method and if got ErrRetryTask error restore original transaction from
// the backup.
func (i *implementation) Run(ctx context.Context, tx Transaction) error {
	if tx != nil {
		return errors.New("nested transactions are not supported")
	}

	if task := i.nextTask(); task != nil {
		// Task supported by this implementation
		if e := task.Run(ctx, tx); e != nil {
			return e
		}

		// Reset retry attempt
		i.Attempt = 0
	}

	return nil
}

func (i *implementation) nextTask() Task {
	var q *deque.Deque[*data.Container]
	if i.RollbackIndicator {
		q = i.RollbackStack
	} else {
		q = i.PendingTasks
	}

	if taskContainer, present := q.PopFront(); present {
		// Not all tasks complete
		// Note: task removed from current transaction at this point
		if task, e := i.manager.GetTask(taskContainer); e == nil && task != nil {
			// Task supported by this implementation
			return task
		}
	}

	return nil
}

// SetRollback transaction indicator
func (i *implementation) SetRollback() error {
	if i.RollbackIndicator {
		// Already at rollback state
		return errors.New("invalid transaction state")
	}

	i.RollbackIndicator = true

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

// AddRollbackTask add rollback task to the transaction
func (i *implementation) AddRollbackTask(kind string, setup ...data.Setup) error {
	container, e := i.NewTask(kind, setup...)
	if e != nil {
		return fmt.Errorf(`create new rollback task error: %w`, e)
	}

	// Queue task for execution
	i.QueueRollbackTask(container)

	return nil
}

// QueueTask task for execution
func (i *implementation) QueueTask(container *data.Container) {
	// Normal task order
	i.PendingTasks.PushBack(container)
}

// QueueRollbackTask for possible rollback
func (i *implementation) QueueRollbackTask(container *data.Container) {
	// Reverse task order
	i.RollbackStack.PushFront(container)
}

// NewTask return new task context at data exchange format
func (i *implementation) NewTask(kind string, setup ...data.Setup) (*data.Container, error) {
	return i.manager.NewTask(kind, setup...)
}

// Backup transaction
func (i *implementation) Backup() (data.Raw, error) {
	return i.manager.Backup(i)
}

// NextAttempt check if next retry attempt is possible
// Note: This operation increase internal retry counter
func (i *implementation) NextAttempt(retryTaskError *RetryTaskError) error {
	i.Attempt++

	if i.Attempt > retryTaskError.maxRetries {
		// Retry limit exceed
		return ErrRetryLimitExceeded
	}

	// Can retry transaction
	return nil
}
