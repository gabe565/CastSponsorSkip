package util

import (
	"context"
	"errors"
	"time"
)

type HaltRetriesError struct {
	Err error
}

func (err HaltRetriesError) Error() string {
	return err.Err.Error()
}

func (err HaltRetriesError) Unwrap() error {
	return err.Err
}

func HaltRetries(err error) error {
	return HaltRetriesError{Err: err}
}

func Retry(ctx context.Context, attempts uint, sleep time.Duration, fn func(try uint) error) error {
	var err error
	for i := range attempts {
		if err = fn(i); err == nil {
			return nil
		}

		var haltRetriesErr HaltRetriesError
		if errors.As(err, &haltRetriesErr) {
			return haltRetriesErr.Err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
			sleep *= 2
		}
	}
	return err
}
