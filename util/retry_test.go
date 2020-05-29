package util_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cycloidio/terracognita/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	throttlingErr = "Throttling"
)

func TestRetry(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var count int
		fn := func() error {
			count++
			return nil
		}

		err := util.Retry(fn, 3, 0*time.Second)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("SuccessOnTimes", func(t *testing.T) {
		var count int
		fn := func() error {
			count++
			if count == 1 {
				return awserr.New(throttlingErr, "message", nil)
			}
			return nil
		}

		err := util.Retry(fn, 3, 0*time.Second)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
	t.Run("Error", func(t *testing.T) {
		var count int
		fn := func() error {
			count++
			return awserr.New(throttlingErr, "message", nil)
		}

		err := util.Retry(fn, 3, 0*time.Second)
		require.Equal(t, err, awserr.New(throttlingErr, "message", nil))
		assert.Equal(t, 3, count)
	})
	t.Run("NoRetrySTD", func(t *testing.T) {
		var count int
		fn := func() error {
			count++
			return fmt.Errorf("some std error")
		}

		err := util.Retry(fn, 3, 0*time.Second)
		require.Equal(t, err, fmt.Errorf("some std error"))
		assert.Equal(t, 1, count)
	})
	t.Run("NoRetry|pk/errors", func(t *testing.T) {
		var count int
		eerr := errors.New("some custom error")
		fn := func() error {
			count++
			return eerr
		}

		err := util.Retry(fn, 3, 0*time.Second)
		require.Equal(t, eerr, err)
		assert.Equal(t, 1, count)
	})
}
