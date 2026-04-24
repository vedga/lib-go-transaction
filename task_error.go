package transaction

import (
	"errors"
	"fmt"
)

type (
	// RetryTaskError indicate same task must retry their execution at the same state
	// TODO: Rename to ErrTaskRetry or ErrRetryTask?
	RetryTaskError struct {
		maxRetries uint
	}
)

var (
	// ErrRetryTask indicate task must be retried
	// TODO: Remove when TestTransactionDataStack() come unnecessary
	ErrRetryTask = errors.New("retry task")
	// ErrRetryLimitExceeded indicate retry limit exceed
	ErrRetryLimitExceeded = errors.New("retry limit exceeded")
	// ErrMigrate indicate task require migrate to other execution point with current state
	ErrMigrate = errors.New("migrate")
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
// TODO: Remove when TestTransactionDataStack() come unnecessary
func (i *RetryTaskError) Unwrap() error {
	return ErrRetryTask
}
