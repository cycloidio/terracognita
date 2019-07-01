package aws

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
	"github.com/pkg/errors"
)

func cacheIAMGroups(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamGroups(ctx, a, rt, tags)
		if err != nil {
			return nil, err
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getIAMGroupNames(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheIAMGroups(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Id())
	}

	return names, nil
}

func cacheIAMRoles(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamRoles(ctx, a, rt, tags)
		if err != nil {
			return nil, err
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getIAMRoleNames(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheIAMRoles(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Id())
	}

	return names, nil
}

func cacheIAMUsers(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamUsers(ctx, a, rt, tags)
		if err != nil {
			return nil, err
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getIAMUserNames(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheIAMUsers(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Id())
	}

	return names, nil
}

func cacheRoute53Zones(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = route53Zones(ctx, a, rt, tags)
		if err != nil {
			return nil, err
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getRoute53ZoneIDs(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheRoute53Zones(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	ids := make([]string, 0, len(rs))
	for _, i := range rs {
		ids = append(ids, i.Data().Id())
	}

	return ids, nil
}

func cacheSESDomainIdentities(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = sesDomainIdentities(ctx, a, rt, tags)
		if err != nil {
			return nil, err
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getSESDomainIdentityDomains(ctx context.Context, a *aws, rt string, tags []tag.Tag) ([]string, error) {
	rs, err := cacheSESDomainIdentities(ctx, a, rt, tags)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	domains := make([]string, 0, len(rs))
	for _, i := range rs {
		domains = append(domains, i.Data().Id())
	}

	return domains, nil
}
