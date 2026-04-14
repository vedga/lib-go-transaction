package transaction

import (
	"errors"
	"fmt"
)

type (
	// RetryTaskError indicate same task must retry their execution at the same state
	RetryTaskError struct {
		retries uint
	}
)

var (
	// ErrRetryTask indicate task must be retried
	ErrRetryTask = errors.New("retry task")
)

// NewRetryTaskError return new error implementation
func NewRetryTaskError(retries uint) error {
	return &RetryTaskError{
		retries: retries,
	}
}

// Error implement error interface
func (i *RetryTaskError) Error() string {
	return fmt.Sprintf("retries exceeded (max %d)", i.retries)
}

// Unwrap error to basic for direct comparison
func (i *RetryTaskError) Unwrap() error {
	return ErrRetryTask
}
