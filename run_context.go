package transaction

import (
	"context"

	"github.com/vedga/lib-go-transaction/data"
)

type (
	runContextKeyType int

	RunCtx struct {
		Rollback bool
		Attempt  uint
	}
)

const (
	runContextKey runContextKeyType = iota
)

// RunContext return task execution context if any
func RunContext(ctx context.Context) *RunCtx {
	if runCtx, e := data.As[*RunCtx](ctx.Value(runContextKey)); e == nil {
		return runCtx
	}

	return nil
}

func withRunContext(ctx context.Context, runCtx *RunCtx) context.Context {
	return context.WithValue(ctx, runContextKey, runCtx)
}
