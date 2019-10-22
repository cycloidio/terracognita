package util

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/cycloidio/terracognita/log"
)

const (
	timesDefault    = 3
	intervalDefault = 30 * time.Second
)

// RetryFn it's a type to represent the function wrapped for the
// Retry or RetryDefault methods
type RetryFn func() error

// Retry calls rfn and checks the errors, if it matches the error
// and if it does it tries 'times' withing the 'interval'
func Retry(rfn RetryFn, times int, interval time.Duration) error {
	err := rfn()
	times--
	if err != nil {
		if times == 0 {
			return err
		}
		// If the error is from the stdlib we just continue with them
		// This is a fix because 'request.IsErrorRetryable(err)' will always
		// retry normal errors "just in case" and we do not want to retry errors
		// that we return
		if fmt.Sprintf("%T", err) == "*errors.errorString" {
			return err
		}
		if request.IsErrorRetryable(err) || request.IsErrorThrottle(err) || request.IsErrorExpiredCreds(err) {
			log.Get().Log("func", "utils.Retry", "msg", "waiting for Throttling error", "times-left", times)
			time.Sleep(interval)
			return Retry(rfn, times, interval)
		}
	}

	return err
}

// RetryDefault calls Retry with the default parameters
func RetryDefault(rfn RetryFn) error {
	return Retry(rfn, timesDefault, intervalDefault)
}
