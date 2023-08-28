package util

import (
	"context"
	"time"
)

type ErrHaltRetries struct {
	Err error
}

func (err ErrHaltRetries) Error() string {
	return err.Err.Error()
}

func (err ErrHaltRetries) Unwrap() error {
	return err.Err
}

func HaltRetries(err error) error {
	return ErrHaltRetries{Err: err}
}

func Retry(ctx context.Context, attempts uint, sleep time.Duration, fn func(try uint) error) error {
	var err error
	for i := uint(0); i < attempts; i += 1 {
		if err = fn(i); err == nil {
			return nil
		} else {
			switch err := err.(type) {
			case ErrHaltRetries:
				return err.Err
			}
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
