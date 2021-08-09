package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

// Caches resources, that may be needed for other ones
// 1 - add cache method
// 2 - declare method to get the cached resources

//Caching VirtualNetworks
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

//Caching Virtual Machines
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

////Caching Workflows
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

// Caching Security Group
func cacheSecurityGroups(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = networkSecurityGroups(ctx, a, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get Security Groups")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getSecurityGroups(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheSecurityGroups(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Caching routeTables
func cacheRouteTables(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = routeTables(ctx, a, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get routeTables")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getRouteTables(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheRouteTables(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Caching container registries
func cacheContainerRegistries(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = containerRegistries(ctx, a, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get ContainerRegistries")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getContainerRegistries(ctx context.Context, a *azurerm, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheContainerRegistries(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}
