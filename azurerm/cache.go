package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

func cacheVirtualNetworks(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = virtualNetworks(ctx, a, rt, filters)
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

func getVirtualNetworkNames(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualNetworks(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheVirtualMachines(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = virtualMachines(ctx, a, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get virtual machines")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getVirtualMachineNames(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualMachines(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheWorkflows(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = logicAppWorkflows(ctx, a, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get workflows")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getWorkflowNames(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheWorkflows(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}
