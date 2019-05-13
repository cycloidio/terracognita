package util

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

const (
	// AWSThrottlingCode it's the code used for AWS services
	// to anounce the API limit
	AWSThrottlingCode = "Throttling"

	timesDefault    = 3
	intervalDefault = 10 * time.Second
)

// RetryFn it's a type to represent the function wrapped for the
// Retry or RetryDefault methods
type RetryFn func() error

// Retry calls rfn and checks the errors, if it matches the error
// and if it does it tries 'times' withing the 'interval'
func Retry(rfn RetryFn, times int, interval time.Duration) error {
	err := rfn()
	times -= 1
	if err != nil {
		if times == 0 {
			return err
		}
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == AWSThrottlingCode {
				time.Sleep(interval)
				return Retry(rfn, times, interval)
			}
		}
	}

	return err
}

// RetryDefault calls Retry with the default parameters
func RetryDefault(rfn RetryFn) error {
	return Retry(rfn, timesDefault, intervalDefault)
}
