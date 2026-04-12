package transaction

import (
	"context"
	"fmt"
	"io"

	"github.com/vedga/lib-go-transaction/data"
)

type (
	// Task interface declaration
	Task interface {
		Run(ctx context.Context, tx Transaction) error
	}

	// TaskManager implementation
	TaskManager struct {
		dataManager *data.Manager
	}
)

// NewTaskManager return task manager implementation
func NewTaskManager(taskProducers data.Producers, options ...data.Option) *TaskManager {
	// Validate producers
	for _, producer := range taskProducers {
		descriptor, e := producer()
		if e != nil {
			panic(`Task producer is not usable: ` + e.Error())
		}

		if _, e = data.DescriptorValue[Task](descriptor); e != nil {
			panic(`Non task producer used in task manager: ` + e.Error())
		}
	}

	return &TaskManager{
		dataManager: data.NewManager(taskProducers, options...),
	}
}

// Write task to io.Writer
func (i *TaskManager) Write(w io.Writer, descriptor *data.Descriptor) error {
	return i.dataManager.Write(w, descriptor)
}

// Read task from io.Reader
func (i *TaskManager) Read(r io.Reader) (*data.Descriptor, error) {
	o, e := i.dataManager.Read(r)
	if e != nil {
		return nil, e
	}

	// Verify type
	if _, e = data.DescriptorValue[Task](o); e != nil {
		return nil, fmt.Errorf("read invalid task: %w", e)
	}

	return o, nil
}

// NewContainer return task exchange container
func (i *TaskManager) NewContainer(descriptor *data.Descriptor) (*data.Container, error) {
	return i.dataManager.NewContainer(descriptor)
}

// New return new task descriptor
func (i *TaskManager) New(kind string, setup ...data.Setup) (*data.Descriptor, error) {
	return i.dataManager.New(kind, setup...)
}
