package transaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunContext(t *testing.T) {
	t.Parallel()

	type (
		args struct {
			ctx context.Context
		}
	)
	tests := []struct {
		name string
		args args
		want *RunCtx
	}{
		{
			name: "With value in the context present (0) and rollback indicator is true",
			args: args{
				ctx: func() context.Context {
					return withRunContext(context.Background(), &RunCtx{
						Rollback: true,
					})
				}(),
			},
			want: &RunCtx{
				Rollback: true,
			},
		},
		{
			name: "With value in the context present (1)",
			args: args{
				ctx: func() context.Context {
					return withRunContext(context.Background(), &RunCtx{
						Attempt: 1,
					})
				}(),
			},
			want: &RunCtx{
				Attempt: 1,
			},
		},
		{
			name: "With no Attempt value in the context",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := RunContext(tt.args.ctx)

			assert.Equal(t, tt.want, got)
		})
	}
}
