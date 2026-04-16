package transaction

import (
	"context"

	"github.com/vedga/lib-go-transaction/data_old"
)

//go:generate mockgen -destination=mock/$GOFILE -source $GOFILE

type (
	// Task interface declaration
	Task interface {
		Run(ctx context.Context, tx Transaction) error
	}

	// TaskManager implementation
	TaskManager struct {
		dataManager *data_old.Manager
	}
)

// NewTaskManager return task manager implementation
func NewTaskManager(taskProducers data_old.Producers, options ...data_old.Option) *TaskManager {
	// Validate producers
	for _, producer := range taskProducers {
		descriptor, e := producer()
		if e != nil {
			panic(`Task producer is not usable: ` + e.Error())
		}

		if _, e = data_old.DescriptorValue[Task](descriptor); e != nil {
			panic(`Non task producer used in task manager: ` + e.Error())
		}
	}

	return &TaskManager{
		dataManager: data_old.NewManager(taskProducers, options...),
	}
}

// Coder return implementation for specified task data_old Descriptor
func (i *TaskManager) Coder(descriptor *data_old.Descriptor) data_old.Serializable {
	return i.dataManager.Coder(descriptor)
}

// New return new task descriptor
func (i *TaskManager) New(kind string, setup ...data_old.Setup) (*data_old.Descriptor, error) {
	return i.dataManager.New(kind, setup...)
}
