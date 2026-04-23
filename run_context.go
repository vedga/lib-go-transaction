package transaction

import (
	"context"

	"github.com/vedga/lib-go-transaction/data"
)

type (
	runContextKeyType int

	runContextImplementation struct {
		rollback bool
		attempt  uint
	}

	// RunCtx interface for access to the running task context
	RunCtx interface {
		Rollback() bool
		Attempt() uint
	}
)

const (
	runContextKey runContextKeyType = iota
)

// Rollback indicate task Run() method perform transaction rollback
func (i *runContextImplementation) Rollback() bool {
	return i.rollback
}

// Attempt indicate task Run() method running attempt
func (i *runContextImplementation) Attempt() uint {
	return i.attempt
}

// RunContext return task execution context if any
func RunContext(ctx context.Context) RunCtx {
	if runCtx, e := data.As[RunCtx](ctx.Value(runContextKey)); e == nil {
		return runCtx
	}

	return nil
}

func withRunContext(ctx context.Context, runCtx *runContextImplementation) context.Context {
	return context.WithValue(ctx, runContextKey, runCtx)
}
