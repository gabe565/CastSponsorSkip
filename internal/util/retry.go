package util

import (
	"context"
	"time"
)

func Retry(ctx context.Context, attempts uint, sleep time.Duration, fn func(try uint) error) error {
	var err error
	for i := uint(0); i < attempts; i += 1 {
		if err = fn(i); err == nil {
			return nil
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
