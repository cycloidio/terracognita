package cache

import (
	"github.com/cycloidio/terraforming/errcode"
	"github.com/cycloidio/terraforming/provider"
)

// Cache implements a simple cache of provider.Resource
// it's not concurrently safe
type Cache interface {
	// Set set's the rs to the key
	// if an already existing key
	// was there, it'll return an error
	Set(key string, rs []*provider.Resource) error

	// Get get's the values of the key
	// if the key is not found an error
	// is returned
	Get(key string) ([]*provider.Resource, error)
}

type cache struct {
	data map[string][]*provider.Resource
}

func New() Cache {
	return &cache{
		data: make(map[string][]*provider.Resource),
	}
}

func (c *cache) Set(key string, rs []*provider.Resource) error {
	_, ok := c.data[key]
	if ok {
		return errcode.ErrCacheKeyAlreadyExisting
	}
	c.data[key] = rs
	return nil
}

func (c *cache) Get(key string) ([]*provider.Resource, error) {
	rs, ok := c.data[key]
	if !ok {
		return nil, errcode.ErrCacheKeyNotFound
	}

	return rs, nil
}
