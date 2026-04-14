package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_Backup(t *testing.T) {
	t.Parallel()

	type (
		args struct {
			container *Container
		}
	)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Container backup and restore",
			args: args{
				container: &Container{
					Kind:    "Test container",
					Payload: []byte(`{"foo":"bar"}`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			backup, e := tt.args.container.Backup()
			assert.NoError(t, e)

			var got *Container
			got, e = Restore(backup)
			assert.NoError(t, e)

			assert.Equal(t, tt.args.container, got)
		})
	}
}
