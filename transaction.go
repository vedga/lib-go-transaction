package transaction

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/vedga/lib-go-transaction/data"
	"github.com/vedga/lib-go-transaction/deque"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE
type (
	// Transaction interface declaration
	Transaction interface {
		Task
		ID() string
		Encode() (data.Bytes, error)
		AddTask(kind string, setup ...data.Setup) error
		AddRollbackTask(kind string, setup ...data.Setup) error
		QueueTask(kind string, task Task) error
		QueueEncodedTask(encodedTask data.Bytes) error
		QueueRollbackTask(kind string, task Task) error
		PushData(items ...data.Bytes)
		PopData() (data.Bytes, bool)
		DataCount() int
		ClearData()
		Rollback() error
		NewTask(kind string, setup ...data.Setup) (Task, error)
		NextAttempt(maxRetries uint) error
	}

	// implementation of the transaction task
	implementation struct {
		manager           *Manager
		TxID              string                   `json:"id"`
		RollbackIndicator bool                     `json:"ri"`
		TaskAttempt       uint                     `json:"ta"`
		PendingTasks      *deque.Deque[data.Bytes] `json:"tq"`
		RollbackStack     *deque.Deque[data.Bytes] `json:"rs"`
		DataStack         *deque.Deque[data.Bytes] `json:"ds"`
	}
)

var (
	// ErrNoAvailableTasks indicate no available tasks inside transaction
	ErrNoAvailableTasks = errors.New("no available tasks")
	// ErrOutboxPattern indicate using outbox pattern
	ErrOutboxPattern = errors.New("outbox pattern")
)

func withConstructor(txID string) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.PendingTasks = deque.New[data.Bytes](0)
		o.RollbackStack = deque.New[data.Bytes](0)
		o.DataStack = deque.New[data.Bytes](0)

		// Also setup transaction ID
		setup := WithTransactionID(txID)

		return setup(o)
	})
}

func withTransactionManager(manager *Manager) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.manager = manager

		return nil
	})
}

// WithTransactionID set transaction ID
func WithTransactionID(txID string) data.Setup {
	return data.NewSetup[implementation](func(o *implementation) error {
		o.TxID = txID

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
func (i *implementation) Run(ctx context.Context, txKind string, tx Transaction) error {
	if !strings.EqualFold(txKind, TxKind) {
		return errors.New("unsupported transaction")
	}

	if tx != nil {
		return errors.New("nested transactions are not supported")
	}

	if taskKind, task := i.nextTask(); task != nil {
		// Real attempt number passed via execution context
		taskCtx := withTaskContext(ctx, i.TaskAttempt)

		// Task supported by this implementation, reset attempt counter because transaction may be backup in the
		// task if outbox pattern is used.
		i.TaskAttempt = 0

		// Execute task
		e := task.Run(taskCtx, taskKind, i)
		if e != nil {
			// Some error occurred
			if errors.Is(e, ErrOutboxPattern) {
				// Using outbox pattern indicate no more tasks in this tran
				return ErrNoAvailableTasks
			}

			return e
		}

		if i.taskQueue().Size() > 0 {
			// Transaction is not complete
			return nil
		}

		// Transaction operation complete
	}

	// No available tasks in this transaction
	return ErrNoAvailableTasks
}

// ID return transaction ID
func (i *implementation) ID() string {
	return i.TxID
}

// Attempt return task execution attempt
func (i *implementation) Attempt() uint {
	return i.TaskAttempt
}

// Encode transaction context to the byte sequence
func (i *implementation) Encode() (data.Bytes, error) {
	return i.manager.Encode(TxKind, i)
}

func (i *implementation) nextTask() (string, Task) {
	q := i.taskQueue()

	if encoded, present := q.PopFront(); present {
		// Not all tasks complete
		// Note: task removed from current transaction at this point
		if kind, task, e := i.manager.DecodeTask(encoded); e == nil && task != nil {
			// Task supported by this implementation
			return kind, task
		}
	}

	// No more tasks or task type isn't supported
	return ``, nil
}

func (i *implementation) taskQueue() *deque.Deque[data.Bytes] {
	if i.RollbackIndicator {
		return i.RollbackStack
	}

	return i.PendingTasks
}

// Rollback transaction
func (i *implementation) Rollback() error {
	if i.RollbackIndicator {
		// Already at rollback state
		return errors.New("invalid transaction state")
	}

	i.RollbackIndicator = true

	// Data stack not used in rollback mode
	i.ClearData()

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

	return i.QueueEncodedTask(encodedTask)
}

// QueueEncodedTask add encoded task to the transaction
func (i *implementation) QueueEncodedTask(encodedTask data.Bytes) error {
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

// PushData push custom data to the data stack
// Note:
// If using multiple items then it
func (i *implementation) PushData(items ...data.Bytes) {
	// Reverse multiple items for stack ordering
	slices.Reverse(items)

	i.DataStack.PushFront(items...)
}

// PopData return most custom data value if any
func (i *implementation) PopData() (data.Bytes, bool) {
	return i.DataStack.PopFront()
}

// DataCount return number of items in data stack
func (i *implementation) DataCount() int {
	return i.DataStack.Size()
}

// ClearData clear data stack
func (i *implementation) ClearData() {
	i.DataStack.Clear()
}

// NewTask return new task context at data_old exchange format
func (i *implementation) NewTask(kind string, setup ...data.Setup) (Task, error) {
	return i.manager.NewTask(kind, setup...)
}

// NextAttempt check if next retry attempt is possible
// Note: This operation increase internal retry counter
func (i *implementation) NextAttempt(maxRetries uint) error {
	i.TaskAttempt++

	if i.TaskAttempt > maxRetries {
		// Retry limit exceed
		return ErrRetryLimitExceeded
	}

	// Can retry transaction
	return nil
}
