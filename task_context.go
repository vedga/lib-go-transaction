package transaction

import (
	"context"

	"github.com/vedga/lib-go-transaction/data"
)

type (
	taskContextKeyType int
)

const (
	taskContextKey taskContextKeyType = iota
)

// Attempt return task execution attempt number
func Attempt(ctx context.Context) uint {
	if v, e := data.As[uint](ctx.Value(taskContextKey)); e == nil {
		return v
	}

	return 0
}

func withTaskContext(ctx context.Context, attempt uint) context.Context {
	return context.WithValue(ctx, taskContextKey, attempt)
}
