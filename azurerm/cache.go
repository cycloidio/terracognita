package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
	"github.com/pkg/errors"
)

func cacheVirtualNetworks(ctx context.Context, a *azurerm, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = virtualNetworks(ctx, a, rt, tags)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get virtual networks")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getVirtualNetworkNames(ctx context.Context, a *azurerm, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheVirtualNetworks(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}
