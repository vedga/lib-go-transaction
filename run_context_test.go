package transaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunContext(t *testing.T) {
	//	t.Parallel()

	type (
		args struct {
			ctx context.Context
		}
	)
	tests := []struct {
		name         string
		args         args
		want         bool
		wantRollback bool
		wantAttempt  uint
	}{
		{
			name: "With value in the context present (0) and rollback indicator is true",
			args: args{
				ctx: func() context.Context {
					return withRunContext(context.Background(), &runContextImplementation{
						rollback: true,
					})
				}(),
			},
			want:         true,
			wantRollback: true,
			wantAttempt:  0,
		},
		{
			name: "With value in the context present (1)",
			args: args{
				ctx: func() context.Context {
					return withRunContext(context.Background(), &runContextImplementation{
						attempt: 1,
					})
				}(),
			},
			want:         true,
			wantRollback: false,
			wantAttempt:  1,
		},
		{
			name: "With no Attempt value in the context",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				}(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//			t.Parallel()

			got := RunContext(tt.args.ctx)
			assert.Equal(t, tt.want, got != nil)

			if got != nil {
				assert.Equal(t, tt.wantRollback, got.Rollback())
				assert.Equal(t, tt.wantAttempt, got.Attempt())
			}
		})
	}
}
