package util_test

import (
	"testing"
	"time"

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
}
