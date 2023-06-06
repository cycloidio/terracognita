package azurerm

import (
	"context"
	"fmt"
	"strings"

	autorestAzure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfazurerm "github.com/hashicorp/terraform-provider-azurerm/provider"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/cycloidio/terracognita/cache"
	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/log"
	"github.com/cycloidio/terracognita/provider"
)

// version of the Terraform provider, this is automatically changed with the 'make update-terraform-provider'
const version = "3.20.0"

// skippableCodes is a list of codes
// which won't make Terracognita failed
// but they will be printed on the output
// they are based on the err.Code() content
// of the Azure error
var skippableCodes = map[string]struct{}{
	"ParentResourceNotFound": struct{}{},
}

type azurerm struct {
	tfAzureRMClient interface{}
	tfProvider      *schema.Provider
	azurerReaders   []*AzureReader

	configuraiton map[string]interface{}

	cache cache.Cache
}

// NewProvider returns a AzureRM Provider
func NewProvider(ctx context.Context, clientID, clientSecret, environment string, resourceGroupNames []string, subscriptionID, tenantID string) (provider.Provider, error) {
	readers := make([]*AzureReader, 0, len(resourceGroupNames))
	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading Azure reader")
	for _, rgn := range resourceGroupNames {
		reader, err := NewAzureReader(ctx, clientID, clientSecret, environment, rgn, subscriptionID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("could not initialize AzureReader: %s", err)
		}
		readers = append(readers, reader)
	}

	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading TF provider")
	tfp := tfazurerm.AzureProvider()

	rawCfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"client_id":       clientID,
		"client_secret":   clientSecret,
		"environment":     environment,
		"subscription_id": subscriptionID,
		"tenant_id":       tenantID,
	})

	log.Get().Log("func", "azurerm.NewProvider", "msg", "loading TF client")
	if diags := tfp.Configure(ctx, rawCfg); diags.HasError() {
		return nil, fmt.Errorf("could not initialize 'terraform/azurerm.Provider.Configure()' because: %s", diags[0].Summary)
	}

	return &azurerm{
		tfAzureRMClient: tfp.Meta(),
		tfProvider:      tfp,
		azurerReaders:   readers,
		cache:           cache.New(),
		configuraiton: map[string]interface{}{
			"environment": environment,
		},
	}, nil
}
func (a *azurerm) HasResourceType(t string) bool {
	_, err := ResourceTypeString(t)
	return err == nil
}

func (a *azurerm) Region() string                        { return a.azurerReaders[0].GetLocation() }
func (a *azurerm) String() string                        { return "azurerm" }
func (a *azurerm) TagKey() string                        { return "tags" }
func (a *azurerm) Source() string                        { return "hashicorp/azurerm" }
func (a *azurerm) Version() string                       { return version }
func (a *azurerm) Configuration() map[string]interface{} { return a.configuraiton }

func (a *azurerm) ResourceTypes() []string {
	return ResourceTypeStrings()
}

func (a *azurerm) Resources(ctx context.Context, t string, f *filter.Filter) ([]provider.Resource, error) {
	rt, err := ResourceTypeString(t)
	if err != nil {
		return nil, err
	}

	rfn, ok := resources[rt]
	if !ok {
		return nil, errors.Errorf("the resource %q it's not implemented", t)
	}

	resources := make([]provider.Resource, 0, 0)
	for _, ar := range a.azurerReaders {
		nres, err := rfn(ctx, a, ar, t, f)
		if err != nil {
			// we filter the error from Azure and return a custom error
			// type if it's an error that we want to skip
			// Remove all wrap layer to get the right type
			unwrapErr := err
			for errors.Unwrap(unwrapErr) != nil {
				unwrapErr = errors.Unwrap(unwrapErr)
			}
			if reqErr, ok := unwrapErr.(*autorestAzure.RequestError); ok {
				if _, ok := skippableCodes[reqErr.ServiceError.Code]; ok {
					return nil, fmt.Errorf("%w: %v", errcode.ErrProviderAPI, reqErr)
				}
			}

			return nil, errors.Wrapf(err, "error while reading from resource %q", t)
		}
		for _, r := range nres {
			// As we are already doing the filter in the reader
			// we can set all the resources to ignore further
			// Tag fitlers
			r.SetIgnoreTagFilter(true)
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (a *azurerm) TFClient() interface{} {
	return a.tfAzureRMClient
}

func (a *azurerm) TFProvider() *schema.Provider {
	return a.tfProvider
}

func (a *azurerm) FixResource(t string, v cty.Value) (cty.Value, error) {
	var err error
	switch t {
	case "azurerm_virtual_machine":
		// We should never set the managed_disk_id if create_option is FromImage
		// the unsetManageDiskID is the list of all the indexed of the storage_account_type that
		// have the create_option as FromImage
		var unsetManageDiskID = make(map[int]struct{})
		err = cty.Walk(v, func(path cty.Path, val cty.Value) (bool, error) {
			if len(path) == 3 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {

					switch gas.Name {
					case "storage_os_disk":
						if path[2].(cty.GetAttrStep).Name == "create_option" {
							var co string
							err := gocty.FromCtyValue(val, &co)
							if err != nil {
								return false, errors.Wrapf(err, "failed to convert CTY value to GO type")
							}
							if co == "FromImage" {
								var idx int
								err := gocty.FromCtyValue(path[1].(cty.IndexStep).Key, &idx)
								if err != nil {
									return false, errors.Wrapf(err, "failed to convert CTY value to GO type")
								}
								unsetManageDiskID[idx] = struct{}{}
							}
						}
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
		}
		v, err = cty.Transform(v, func(path cty.Path, v cty.Value) (cty.Value, error) {
			if len(path) == 3 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "os_profile":
						switch path[2].(cty.GetAttrStep).Name {
						case "admin_password":
							// The value can't be retrieved. Terraform Azure provider set "ignored-as-imported" which is not a valid password.
							// In this case, set the default password which respects Azure constraints
							return cty.StringVal("Ignored-as-!mport3d"), nil
						}
					case "storage_os_disk":
						var idx int
						err := gocty.FromCtyValue(path[1].(cty.IndexStep).Key, &idx)
						if err != nil {
							return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						if _, ok := unsetManageDiskID[idx]; ok && path[2].(cty.GetAttrStep).Name == "managed_disk_id" {
							return cty.NullVal(cty.String), nil
						}
					case "storage_data_disk":
						switch path[2].(cty.GetAttrStep).Name {
						case "managed_disk_id":
							// Since we manage extra disk with the resource itself, id should never be set
							return cty.NullVal(cty.String), nil
						case "create_option":
							// Since we manage extra disk with the resource itself, id should always be set to Empty
							return cty.StringVal("Empty"), nil
						}
					}
				}
			}
			return v, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
		}
	case "azurerm_managed_disk":
		// Are set to default value which is good but only with type not UltraSSD or PremiumV2, else terraform will raise an error
		var unsetDiskReadWrite bool
		err = cty.Walk(v, func(path cty.Path, v cty.Value) (bool, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "storage_account_type":
						var sat string
						err := gocty.FromCtyValue(v, &sat)
						if err != nil {
							return false, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						if sat != "UltraSSD" && sat != "PremiumV2" {
							unsetDiskReadWrite = true
						}
					}
				}
			}
			return true, nil
		})
		if err != nil {
			return cty.NullVal(cty.EmptyObject), errors.Wrapf(err, "failed to convert CTY value to GO type")
		}
		// Once we know we have to unset the disk default values we Transform the config
		if unsetDiskReadWrite {
			v, err = cty.Transform(v, func(path cty.Path, v cty.Value) (cty.Value, error) {
				if len(path) > 0 {
					if gas, ok := path[0].(cty.GetAttrStep); ok {
						switch gas.Name {
						case "disk_iops_read_write", "disk_mbps_read_write":
							return cty.NullVal(cty.Number), nil
						}
					}
				}
				return v, nil
			})
			if err != nil {
				return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
			}
		}

	case "azurerm_virtual_machine_data_disk_attachment":
		v, err = cty.Transform(v, func(path cty.Path, v cty.Value) (cty.Value, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "create_option":
						// For some reason this values is Empty sometime.
						// When importing we only list attached disk. So it should always be set to Attach
						return cty.StringVal("Attach"), nil
					}
				}
			}
			return v, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
		}
	case "azurerm_windows_virtual_machine":
		v, err = cty.Transform(v, func(path cty.Path, v cty.Value) (cty.Value, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "admin_password":
						// The value can't be retrieved. Terraform Azure provider set "ignored-as-imported" which is not a valid password.
						// In this case, set the default password which respects Azure constraints
						return cty.StringVal("Ignored-as-!mport3d"), nil
					case "platform_fault_domain":
						// By default this attribute is set, but this is not a valid value with terraform apply.
						// In this case, we shouldn't write platform_fault_domain.
						var pfd int
						err := gocty.FromCtyValue(v, &pfd)
						if err != nil {
							return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						if pfd == -1 {
							return cty.NullVal(cty.Number), nil
						}
					case "availability_set_id":
						// For some reason this values has the end part of the ID capitalized which
						// makes it fail when trying to create a link between resources
						var asi string
						err := gocty.FromCtyValue(v, &asi)
						if err != nil {
							return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
						}
						sasi := strings.Split(asi, "/")
						sasi[len(sasi)-1] = strings.ToLower(sasi[len(sasi)-1])
						return cty.StringVal(strings.Join(sasi, "/")), nil
					}
				}
			}
			return v, nil
		})
	case "azurerm_network_security_group":
		v, err = cty.Transform(v, func(path cty.Path, v cty.Value) (cty.Value, error) {
			if len(path) > 0 {
				if gas, ok := path[0].(cty.GetAttrStep); ok {
					switch gas.Name {
					case "security_rule":
						if len(path) < 3 {
							return v, nil
						}
						switch path[2].(cty.GetAttrStep).Name {
						case "protocol":
							// For some reason the Protocol is set like: TCP, but the valid value is Tcp
							var sp string
							err := gocty.FromCtyValue(v, &sp)
							if err != nil {
								return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
							}
							return cty.StringVal(cases.Title(language.English).String(strings.ToLower(sp))), nil
						}
					}
				}
			}
			return v, nil
		})
		if err != nil {
			return v, errors.Wrapf(err, "failed to convert CTY value to GO type")
		}

	}
	return v, nil
}
func (a *azurerm) FilterByTags(tags interface{}) error { return nil }
