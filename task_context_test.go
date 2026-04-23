package transaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttempt(t *testing.T) {
	type (
		args struct {
			ctx context.Context
		}
	)
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "With value in the context present",
			args: args{
				ctx: func() context.Context {
					return withTaskContext(context.Background(), 1)
				}(),
			},
			want: 1,
		},
		{
			name: "With no Attempt value in the context",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				}(),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Attempt(tt.args.ctx)

			assert.Equal(t, tt.want, got)
		})
	}
}
