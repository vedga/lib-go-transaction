package transaction

import (
	"errors"
	"fmt"
)

type (
	// RetryTaskError indicate same task must retry their execution at the same state
	RetryTaskError struct {
		maxRetries uint
	}
)

var (
	// ErrRetryTask indicate task must be retried
	ErrRetryTask = errors.New("retry task")
	// ErrRetryLimitExceeded indicate retry limit exceed
	ErrRetryLimitExceeded = errors.New("retry limit exceeded")
)

// NewRetryTaskError return new error implementation
func NewRetryTaskError(maxRetries uint) error {
	return &RetryTaskError{
		maxRetries: maxRetries,
	}
}

// Error implement error interface
func (i *RetryTaskError) Error() string {
	return fmt.Sprintf("retries exceeded (max %d)", i.maxRetries)
}

// Unwrap error to basic for direct comparison
func (i *RetryTaskError) Unwrap() error {
	return ErrRetryTask
}
