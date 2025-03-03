package util

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("test")

func TestRetry(t *testing.T) {
	t.Run("halt", func(t *testing.T) {
		var runs int
		err := Retry(t.Context(), 10, 0, func(_ uint) error {
			runs++
			return HaltRetries(errTest)
		})
		require.ErrorIs(t, err, errTest)
		assert.Equal(t, 1, runs)
	})

	t.Run("max", func(t *testing.T) {
		var runs int
		err := Retry(t.Context(), 10, 0, func(_ uint) error {
			runs++
			return errTest
		})
		require.Error(t, err)
		assert.Equal(t, 10, runs)
	})

	t.Run("pass on first run", func(t *testing.T) {
		var runs int
		err := Retry(t.Context(), 10, 0, func(_ uint) error {
			runs++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, runs)
	})

	t.Run("pass on fifth run", func(t *testing.T) {
		var runs int
		err := Retry(t.Context(), 10, 0, func(_ uint) error {
			runs++
			if runs < 5 {
				return errTest
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 5, runs)
	})

	t.Run("sleep backoff", func(t *testing.T) {
		var runs int
		start := time.Now()
		err := Retry(t.Context(), 10, time.Millisecond, func(_ uint) error {
			runs++
			if runs < 5 {
				return errTest
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 5, runs)
		assert.Greater(t, time.Since(start), 15*time.Millisecond)
	})
}
