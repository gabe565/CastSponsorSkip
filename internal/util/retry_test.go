package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	t.Run("halt", func(t *testing.T) {
		targetErr := errors.New("test")
		var runs int
		if err := Retry(context.Background(), 10, 0, func(try uint) error {
			runs += 1
			return HaltRetries(targetErr)
		}); !assert.ErrorIs(t, targetErr, err) {
			return
		}
		assert.Equal(t, 1, runs)
	})

	t.Run("max", func(t *testing.T) {
		var runs int
		if err := Retry(context.Background(), 10, 0, func(try uint) error {
			runs += 1
			return errors.New("test")
		}); !assert.Error(t, err) {
			return
		}
		assert.Equal(t, 10, runs)
	})

	t.Run("pass on first run", func(t *testing.T) {
		var runs int
		if err := Retry(context.Background(), 10, 0, func(try uint) error {
			runs += 1
			return nil
		}); !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 1, runs)
	})

	t.Run("pass on fifth run", func(t *testing.T) {
		var runs int
		if err := Retry(context.Background(), 10, 0, func(try uint) error {
			runs += 1
			if runs < 5 {
				return errors.New("test")
			}
			return nil
		}); !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 5, runs)
	})

	t.Run("sleep backoff", func(t *testing.T) {
		var runs int
		start := time.Now()
		if err := Retry(context.Background(), 10, time.Millisecond, func(try uint) error {
			runs += 1
			if runs < 5 {
				return errors.New("test")
			}
			return nil
		}); !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 5, runs)
		assert.Greater(t, time.Since(start), 15*time.Millisecond)
	})
}
