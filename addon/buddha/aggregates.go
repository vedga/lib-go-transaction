package buddha

import (
	"context"
	"fmt"

	transaction "github.com/vedga/lib-go-transaction"
	"github.com/vedga/lib-go-transaction/data"
)

type (
	aggregates struct {
		transaction.Task
		samsara        *Samsara
		BirthIndicator bool `json:"bi"`
	}
)

// WithTxTaskProducer return task aggregates implementation
func WithTxTaskProducer(kind string, samsara *Samsara, producer transaction.TaskProducer) transaction.Option {
	return transaction.WithTxTaskProducer(kind, func(setup ...data.Setup) (transaction.Task, error) {
		task, e := producer(setup...)
		if e != nil {
			return nil, e
		}

		i := &aggregates{
			Task:    task,
			samsara: samsara,
		}

		return data.As[transaction.Task](i)
	})
}

// Run is implementation of transaction.Task interface
// In this point we have dilemma: real task can run in background long time (until it finished or killed by Samsara),
// but transaction.Transaction Run() method must be finished immoderately for detach transaction from main pipeline.
// We can make independed copy of original transaction.Transaction or protect Run() method in the transaction by mutex.
// Note:
// Current implementation use transaction clone.
func (i *aggregates) Run(ctx context.Context, taskKind string, tx transaction.Transaction) error {
	if i.BirthIndicator {
		// Execute business task
		return i.Task.Run(ctx, taskKind, tx)
	}

	// Prepare task registration
	i.BirthIndicator = true

	// Clone transaction because we make independent transaction execution path
	tx = tx.Clone()

	// Re-queue original task to the transaction
	if e := tx.QueueTask(taskKind, i); e != nil {
		return fmt.Errorf("queueing original task error: %w", e)
	}

	if e := i.samsara.Rebirth(ctx, tx); e != nil {
		return fmt.Errorf("rebirthing original task error: %w", e)
	}

	// Take control of the transaction. Original transaction don't handled by caller now.
	return transaction.ErrNoAvailableTasks
}
