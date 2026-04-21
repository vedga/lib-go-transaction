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
		Encode() (data.Bytes, error)
		AddTask(kind string, setup ...data.Setup) error
		AddRollbackTask(kind string, setup ...data.Setup) error
		QueueTask(kind string, task Task) error
		QueueRollbackTask(kind string, task Task) error
		SetRollback() error
		NewTask(kind string, setup ...data.Setup) (Task, error)
		NextAttempt(maxRetries uint) error
	}

	// implementation of the transaction task
	implementation struct {
		manager *Manager
		// ID of the transaction
		ID string
		// RollbackIndicator if true currant action is transaction rollback
		RollbackIndicator bool
		// Attempt of task execution
		Attempt uint
		// PendingTasks contain tasks sequence for execute transaction
		PendingTasks *deque.Deque[data.Bytes]
		// RollbackStack contain transaction rollback sequence
		RollbackStack *deque.Deque[data.Bytes]
	}
)

func withConstructor(ID string) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.ID = ID
		o.PendingTasks = deque.New[data.Bytes](0)
		o.RollbackStack = deque.New[data.Bytes](0)

		return nil
	})
}

func withTransactionManager(manager *Manager) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.manager = manager

		return nil
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

	if _, task := i.nextTask(); task != nil {
		// Task supported by this implementation, reset attempt counter because transaction may be backup in the
		// task if outbox pattern is used.
		i.Attempt = 0

		// Execute task
		return task.Run(ctx, tx)
	}

	return nil
}

// Encode transaction context to the byte sequence
func (i *implementation) Encode() (data.Bytes, error) {
	return i.manager.Encode(txKind, i)
}

func (i *implementation) nextTask() (string, Task) {
	var q *deque.Deque[data.Bytes]
	if i.RollbackIndicator {
		q = i.RollbackStack
	} else {
		q = i.PendingTasks
	}

	if encoded, present := q.PopFront(); present {
		// Not all tasks complete
		// Note: task removed from current transaction at this point
		if kind, task, e := i.manager.DecodeTask(encoded); e == nil && task != nil {
			// Task supported by this implementation
			return kind, task
		}
	}

	return ``, nil
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
	task, e := i.NewTask(kind, setup...)
	if e != nil {
		return fmt.Errorf(`create new task error: %w`, e)
	}

	// Queue task for execution
	return i.QueueTask(kind, task)
}

// AddRollbackTask add rollback task to the transaction
func (i *implementation) AddRollbackTask(kind string, setup ...data.Setup) error {
	task, e := i.NewTask(kind, setup...)
	if e != nil {
		return fmt.Errorf(`create new rollback task error: %w`, e)
	}

	// Queue task for execution
	return i.QueueRollbackTask(kind, task)
}

// QueueTask task for execution
func (i *implementation) QueueTask(kind string, task Task) error {
	encodedTask, e := i.manager.EncodeTask(kind, task)
	if e != nil {
		return e
	}

	// Normal task order
	i.PendingTasks.PushBack(encodedTask)

	return nil
}

// QueueRollbackTask for possible rollback
func (i *implementation) QueueRollbackTask(kind string, task Task) error {
	encodedTask, e := i.manager.EncodeTask(kind, task)
	if e != nil {
		return e
	}

	// Reverse task order
	i.RollbackStack.PushFront(encodedTask)

	return nil
}

// NewTask return new task context at data_old exchange format
func (i *implementation) NewTask(kind string, setup ...data.Setup) (Task, error) {
	return i.manager.NewTask(kind, setup...)
}

// NextAttempt check if next retry attempt is possible
// Note: This operation increase internal retry counter
func (i *implementation) NextAttempt(maxRetries uint) error {
	i.Attempt++

	if i.Attempt > maxRetries {
		// Retry limit exceed
		return ErrRetryLimitExceeded
	}

	// Can retry transaction
	return nil
}
