package google

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

// Quick sum-up of cached resources:
// Compute: compute instances
// Cloud-dns: dns_managed_zones
// Cloud-sql: sql_databases_instance (except the on-prem type ones)
// Storage: storage_buckets

//compute instances
func cacheComputeInstances(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := g.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = computeInstance(ctx, g, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get compute instance")
		}

		err = g.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getComputeInstances(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheComputeInstances(ctx, g, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Cloud dns

func cacheDNSManagedZones(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := g.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = dnsManagedZone(ctx, g, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get dns managed zone")
		}

		err = g.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getDNSManagedZones(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheDNSManagedZones(ctx, g, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// storage

func cacheStorageBuckets(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := g.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}
		rs, err = storageBucket(ctx, g, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get storage buckets")
		}
		err = g.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getStorageBuckets(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheStorageBuckets(ctx, g, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// sql
func cacheSQLDatabaseInstances(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := g.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}
		rs, err = sqlDatabaseInstance(ctx, g, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get sql database instances")
		}
		err = g.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getSQLDatabaseInstances(ctx context.Context, g *google, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheSQLDatabaseInstances(ctx, g, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}
