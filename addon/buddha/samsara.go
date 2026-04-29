package buddha

import (
	"context"
	"errors"
	"sync"

	transaction "github.com/vedga/lib-go-transaction"
)

type (
	// Samsara is implementation of the wheel of samsara
	Samsara struct {
		once                 *sync.Once
		transition           chan struct{}
		wg                   *sync.WaitGroup
		rebirth              func(ctx context.Context, transaction transaction.Transaction) error
		onUnrecoverableError func(ctx context.Context, transaction transaction.Transaction, e error)
	}
)

// New return implementation of the wheel of samsara
func New(rebirth func(ctx context.Context, transaction transaction.Transaction) error) *Samsara {
	return &Samsara{
		once:                 new(sync.Once),
		transition:           make(chan struct{}),
		wg:                   new(sync.WaitGroup),
		rebirth:              rebirth,
		onUnrecoverableError: func(_ context.Context, _ transaction.Transaction, _ error) {},
	}
}

// Close is implementation of the io.Closer interface
func (i *Samsara) Close() error {
	// For safety operation use only once method execution
	i.once.Do(func() {
		// Global shutdown signal
		close(i.transition)

		// Wait until all operation complete
		i.wg.Wait()
	})

	return nil
}

// Rebirth stream of consciousness
func (i *Samsara) Rebirth(_ context.Context, tx transaction.Transaction) error {
	ctx := context.Background()

	txCtx, txCancel := context.WithCancel(ctx)

	// Stream of consciousness monitoring
	go func() {
		select {
		case <-i.transition:
			// Global transition event, kill stream of consciousness in current life
			txCancel()
		case <-txCtx.Done():
			// Nirvana!
		}
	}()

	i.wg.Add(1)
	go func() {
		// Note:
		// This goroutine can be started before previous method Run() in the aggregates implementation is complete.
		// Golang detect possible race condition, therefore we must protect transaction Run() method with mutex.
		defer func() {
			// Processing operation complete
			i.wg.Done()
			txCancel()
		}()

		// Life is life! (c) Opus
		if e := tx.Run(txCtx, transaction.TxKind, nil); e != nil {
			if errors.Is(e, transaction.ErrNoAvailableTasks) {
				// Task operation complete
				return
			}

			// Some unrecoverable error
			i.onUnrecoverableError(ctx, tx, e)

			return
		}

		// Rebirth transaction
		if e := i.rebirth(ctx, tx); e != nil {
			// Some unrecoverable error
			i.onUnrecoverableError(ctx, tx, e)

			return
		}
	}()

	return nil
}
