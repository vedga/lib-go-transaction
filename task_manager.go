package transaction

import (
	"context"

	"github.com/vedga/lib-go-transaction/data"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE

type (
	// Task interface declaration
	Task interface {
		Run(ctx context.Context, tx Transaction) error
	}

	// TaskManager implementation
	TaskManager struct {
		*data.Manager
	}

	// TaskProducer function
	TaskProducer func(setup ...data.Setup) (Task, error)
)

// NewTaskManager return task manager implementation
func NewTaskManager(options ...data.Option) *TaskManager {
	return &TaskManager{
		Manager: data.NewManager(options...),
	}
}

// WithTaskProducer add task producer
func WithTaskProducer(kind string, producer TaskProducer) data.Option {
	return data.WithProducer(kind, func(setup ...data.Setup) (any, error) {
		return producer(setup...)
	})
}

// EncodeTask encode task context to the byte sequence
func (i *TaskManager) EncodeTask(kind string, task Task) (data.Bytes, error) {
	return i.Encode(kind, task)
}

// DecodeTask bytes sequence to the task context
func (i *TaskManager) DecodeTask(source data.Bytes, setup ...data.Setup) (string, Task, error) {
	kind, o, e := i.Decode(source, setup...)
	if e != nil {
		return kind, nil, e
	}

	var task Task
	if task, e = data.As[Task](o); e != nil {
		return kind, nil, e
	}

	return kind, task, nil
}

// NewTask return new task context
func (i *TaskManager) NewTask(kind string, setup ...data.Setup) (Task, error) {
	o, e := i.New(kind, setup...)
	if e != nil {
		return nil, e
	}

	return data.As[Task](o)
}
