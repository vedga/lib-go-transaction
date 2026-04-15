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

// Coder return implementation for specified task data Descriptor
func (i *TaskManager) Coder(descriptor *data.Descriptor) data.Serializable {
	return i.dataManager.Coder(descriptor)
}

// New return new task descriptor
func (i *TaskManager) New(kind string, setup ...data.Setup) (*data.Descriptor, error) {
	return i.dataManager.New(kind, setup...)
}
