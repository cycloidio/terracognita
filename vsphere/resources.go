package vsphere

import (
	"context"
	"errors"

	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix vsphere_ -transform snake -linecomment
const (
	_ ResourceType = iota

	// Host and Cluster Management
	computeCluster // compute_cluster
	resourcePool   // resource_pool

	// Inventory
	datacenter
	folder

	// Storage
	datastoreCluster // datastore_cluster

	// Virtual Machine
	virtualMachine // virtual_machine
)

type rtFn func(ctx context.Context, vs *vsphere, vm *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		computeCluster:   getComputeClusters,
		resourcePool:     getResourcePools,
		datacenter:       getDatacenters,
		folder:           getFolders,
		datastoreCluster: getDatastoreClusters,
		virtualMachine:   getVirtualMachines,
	}
)

func getDatastoreClusters(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.DatastoreClusterList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getDistributedPortGroups(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	v, err := r.CreateContainerView(ctx, r.Common.Client().ServiceContent.RootFolder, []string{"DistributedVirtualPortgroup"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all DistributedVirtualPortgroups
	var dvps []mo.DistributedVirtualPortgroup
	err = v.Retrieve(ctx, []string{"DistributedVirtualPortgroup"}, []string{}, &dvps)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, dvp := range dvps {
		r := provider.NewResource(dvp.ManagedEntity.ExtensibleManagedObject.Self.String(), resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getFolders(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.FolderList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getDatacenters(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.DatacenterList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getVirtualMachines(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.VirtualMachineList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getResourcePools(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.ResourcePoolList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}

func getComputeClusters(ctx context.Context, vs *vsphere, r *reader, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	vms, err := r.Finder.ComputeResourceList(ctx, "/...")
	if err != nil {
		var nferr *find.NotFoundError
		if errors.As(err, &nferr) {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]provider.Resource, 0, len(vms))
	for _, vm := range vms {
		r := provider.NewResource(vm.InventoryPath, resourceType, vs)
		resources = append(resources, r)
	}

	return resources, nil
}
