package aws

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

func cacheAPIGatewayRestApis(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {

	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = apiGatewayRestApis(ctx, a, rt, filters)
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

func getAPIGatewayRestApis(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheAPIGatewayRestApis(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	ids := make([]string, 0, len(rs))
	for _, i := range rs {
		ids = append(ids, i.ID())
	}

	return ids, nil
}

func cacheLoadBalancersV2(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both aws_alb and aws_lb defined, keep only aws_alb
	if filters.IsIncluded("aws_alb", "aws_lb") && (!filters.IsExcluded("aws_alb") && rt == "aws_lb") {
		return nil, nil
	}

	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = albs(ctx, a, rt, filters)
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

func getLoadBalancersV2Arns(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheLoadBalancersV2(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.ID())
	}

	return names, nil
}

func cacheLoadBalancersV2Listeners(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both defined, keep only aws_alb_listener
	if filters.IsIncluded("aws_alb_listener", "aws_lb_listener") && (!filters.IsExcluded("aws_alb_listener") && rt == "aws_lb_listener") {
		return nil, nil
	}

	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = albListeners(ctx, a, rt, filters)
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

func getLoadBalancersV2ListenersArns(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheLoadBalancersV2Listeners(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.ID())
	}

	return names, nil
}

func cacheIAMGroups(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamGroups(ctx, a, rt, filters)
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

func getIAMGroupNames(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheIAMGroups(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.ID())
	}

	return names, nil
}

func cacheIAMRoles(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamRoles(ctx, a, rt, filters)
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

func getIAMRoleNames(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheIAMRoles(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.ID())
	}

	return names, nil
}

func cacheIAMUsers(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = iamUsers(ctx, a, rt, filters)
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

func getIAMUserNames(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheIAMUsers(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.ID())
	}

	return names, nil
}

func cacheRoute53Zones(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = route53Zones(ctx, a, rt, filters)
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

func getRoute53ZoneIDs(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheRoute53Zones(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	ids := make([]string, 0, len(rs))
	for _, i := range rs {
		ids = append(ids, i.ID())
	}

	return ids, nil
}

func cacheSESDomainIdentities(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = sesDomainIdentities(ctx, a, rt, filters)
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

func getSESDomainIdentityDomains(ctx context.Context, a *aws, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheSESDomainIdentities(ctx, a, rt, filters)
	if err != nil {
		return nil, err
	}

	// Get the actual needed value
	// TODO cach this result too
	domains := make([]string, 0, len(rs))
	for _, i := range rs {
		domains = append(domains, i.ID())
	}

	return domains, nil
}
