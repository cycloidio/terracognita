package azurerm

import (
	"context"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/pkg/errors"
)

// Quick sum-up of cached resources:
//   Network -> virtual_networks, security_group, route_tables, virtual_hub, load_balancer
//   Compute -> virtual_machines, virtual_machine_scale_sets
//   Logic -> logic_app_worfklows
//   Container-registry -> container_registry
//   Container-service(k8s)-> kubernetes_cluster
//   Storage -> storage_accounts
//   Databases -> mariadb_server, postregresql_server, mysql_server,mssql_server
//   Redis -> redis_caches
//   DNS -> dns_zones
//   Private DNS -> dns_zones
//   Application Insights -> application_insights
//   Log analytics -> log_analytics_workspace

//Network

func cacheVirtualNetworks(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = virtualNetworks(ctx, a, ar, rt, filters)
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
func getVirtualNetworkNames(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualNetworks(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheSecurityGroups(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = networkSecurityGroups(ctx, a, ar, rt, filters)
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
func getSecurityGroups(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheSecurityGroups(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheRouteTables(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = routeTables(ctx, a, ar, rt, filters)
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
func getRouteTables(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheRouteTables(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheVirtualHubs(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = virtualHubs(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get virtualHubs")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getVirtualHub(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualHubs(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheLbs(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = lbs(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get load balancers")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getLbs(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheLbs(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Compute

func cacheVirtualMachines(ctx context.Context, a *azurerm, ar *AzureReader, rtList []string, filters *filter.Filter) ([]provider.Resource, error) {
	var resources []provider.Resource
	for _, rt := range rtList {
		rs, err := a.cache.Get(rt)
		if err != nil {
			if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
				return nil, errors.WithStack(err)
			}

			rs, err = virtualMachines(ctx, a, ar, rt, filters)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get virtual machines")
			}

			err = a.cache.Set(rt, rs)
			if err != nil {
				return nil, err
			}
			resources = append(resources, rs...)
		}
	}
	return resources, nil
}
func getVirtualMachineNames(ctx context.Context, a *azurerm, ar *AzureReader, rt []string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualMachines(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheVirtualMachineScaleSets(ctx context.Context, a *azurerm, ar *AzureReader, rtList []string, filters *filter.Filter) ([]provider.Resource, error) {
	var resources []provider.Resource

	for _, rt := range rtList {
		rs, err := a.cache.Get(rt)
		if err != nil {
			if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
				return nil, errors.WithStack(err)
			}
			rs, err = virtualMachineScaleSets(ctx, a, ar, rt, filters)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get virtual machines scale sets")
			}

			err = a.cache.Set(rt, rs)
			if err != nil {
				return nil, err
			}
		}
	}
	return resources, nil
}
func getVirtualMachineScaleSetNames(ctx context.Context, a *azurerm, ar *AzureReader, rtList []string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheVirtualMachineScaleSets(ctx, a, ar, rtList, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Logic

func cacheWorkflows(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = logicAppWorkflows(ctx, a, ar, rt, filters)
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
func getWorkflowNames(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheWorkflows(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Container registry

func cacheContainerRegistries(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = containerRegistries(ctx, a, ar, rt, filters)
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
func getContainerRegistries(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheContainerRegistries(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Container service (k8s)

func cacheKubernetesClusters(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = kubernetesClusters(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get KubernetesCluster")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getKubernetesClusters(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheKubernetesClusters(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

//Storage

func cacheStorageAccounts(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = storageAccounts(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get storageAccounts")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getStorageAccounts(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheStorageAccounts(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Database

func cacheMariadbServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = mariadbServers(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get MariaDB Servers")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getMariadbServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheMariadbServers(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheMysqlServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = mysqlServers(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get MySQL Servers")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getMysqlServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheMysqlServers(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cachePostgresqlServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = postgresqlServers(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get PostgreSQL Servers")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getPostgresqlServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cachePostgresqlServers(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

func cacheMsSQLServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = mssqlServers(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get SQL Servers")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}
func getMsSQLServers(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheMsSQLServers(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Redis
func cacheRedisCaches(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = redisCaches(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get Redis Caches")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getRedisCaches(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheRedisCaches(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// DNS

func cacheDNSZones(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = dnsZones(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get DNS Zones")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getDNSZones(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheDNSZones(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Private DNS

func cachePrivateDNSZones(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = privateDNSZones(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get Private DNS Zones")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getPrivateDNSZones(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cachePrivateDNSZones(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

//Application Insights
func cacheApplicationInsights(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = applicationInsights(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get application insigths")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getApplicationInsightsComponents(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cacheApplicationInsights(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}

// Log Analytics
func cachelogAnalyticsWorkspaces(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]provider.Resource, error) {
	rs, err := a.cache.Get(rt)
	if err != nil {
		if errors.Cause(err) != errcode.ErrCacheKeyNotFound {
			return nil, errors.WithStack(err)
		}

		rs, err = logAnalyticsWorkspaces(ctx, a, ar, rt, filters)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get log analytics workspace")
		}

		err = a.cache.Set(rt, rs)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func getLogAnalyticsWorkspaces(ctx context.Context, a *azurerm, ar *AzureReader, rt string, filters *filter.Filter) ([]string, error) {
	rs, err := cachelogAnalyticsWorkspaces(ctx, a, ar, rt, filters)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs))
	for _, i := range rs {
		names = append(names, i.Data().Get("name").(string))
	}

	return names, nil
}
