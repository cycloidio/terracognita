package errcode

import "errors"

// List of all the error Codes used
var (
	ErrResourceNotRead       = errors.New("the resource did not return an ID")
	ErrResourceDoNotMatchTag = errors.New("the resource does not match the required tags")

	ErrCacheKeyNotFound        = errors.New("the key used to search was not found")
	ErrCacheKeyAlreadyExisting = errors.New("the key already exists on the cache")
)
