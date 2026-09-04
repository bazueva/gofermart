package retrier

import (
	"context"
	"time"
)

// ErrorClassifier — интерфейс для проверки того, нужно ли повторять ошибку.
type ErrorClassifier interface {
	IsRetriable(err error) bool
}

func Do(
	ctx context.Context,
	maxAttempts int,
	initialBackoff time.Duration,
	classifier func(error) bool,
	fn func() error,
) error {
	backoff := initialBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if !classifier(err) || attempt == maxAttempts {
			return err
		}

		select {
		case <-time.After(backoff):
			backoff *= 2
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
