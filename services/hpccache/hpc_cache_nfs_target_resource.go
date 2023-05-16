package hpccache

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storagecache/mgmt/2021-09-01/storagecache"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/hpccache/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/hpccache/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceHPCCacheNFSTarget() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceHPCCacheNFSTargetCreateOrUpdate,
		Update: resourceHPCCacheNFSTargetCreateOrUpdate,
		Read:   resourceHPCCacheNFSTargetRead,
		Delete: resourceHPCCacheNFSTargetDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StorageTargetID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageTargetName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"cache_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"namespace_junction": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				// Confirmed with service team that they have a mac of 10 that is enforced by the backend.
				MaxItems: 10,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"namespace_path": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.CacheNamespacePath,
						},
						"nfs_export": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.CacheNFSExport,
						},
						"target_path": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validate.CacheNFSTargetPath,
						},

						"access_policy_name": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "default",
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"target_host_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// TODO: use SDK enums once following issue is addressed
			// https://github.com/Azure/azure-rest-api-specs/issues/13839
			"usage_model": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"READ_HEAVY_INFREQ",
					"READ_HEAVY_CHECK_180",
					"WRITE_WORKLOAD_15",
					"WRITE_AROUND",
					"WRITE_WORKLOAD_CHECK_30",
					"WRITE_WORKLOAD_CHECK_60",
					"WRITE_WORKLOAD_CLOUDWS",
				}, false),
			},
		},
	}
}

func resourceHPCCacheNFSTargetCreateOrUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HPCCache.StorageTargetsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure HPC Cache NFS Target creation.")
	id := parse.NewStorageTargetID(subscriptionId, d.Get("resource_group_name").(string), d.Get("cache_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		resp, err := client.Get(ctx, id.ResourceGroup, id.CacheName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("checking for existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(resp.Response) {
			return tf.ImportAsExistsError("azurerm_hpc_cache_nfs_target", *resp.ID)
		}
	}

	// Construct parameters
	param := &storagecache.StorageTarget{
		StorageTargetProperties: &storagecache.StorageTargetProperties{
			Junctions:  expandNamespaceJunctions(d.Get("namespace_junction").(*pluginsdk.Set).List()),
			TargetType: storagecache.StorageTargetTypeNfs3,
			Nfs3: &storagecache.Nfs3Target{
				Target:     utils.String(d.Get("target_host_name").(string)),
				UsageModel: utils.String(d.Get("usage_model").(string)),
			},
		},
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.CacheName, id.Name, param)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceHPCCacheNFSTargetRead(d, meta)
}

func resourceHPCCacheNFSTargetRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HPCCache.StorageTargetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageTargetID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.CacheName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] HPC Cache NFS Target %q was not found (Resource Group %q, Cache %q) - removing from state!", id.Name, id.ResourceGroup, id.CacheName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving HPC Cache NFS Target %q (Resource Group %q, Cache %q): %+v", id.Name, id.ResourceGroup, id.CacheName, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("cache_name", id.CacheName)

	if props := resp.StorageTargetProperties; props != nil {
		if props.TargetType != storagecache.StorageTargetTypeNfs3 {
			return fmt.Errorf("The type of this HPC Cache Target %q (Resource Group %q, Cahe %q) is not a NFS Target", id.Name, id.ResourceGroup, id.CacheName)
		}
		if nfs3 := props.Nfs3; nfs3 != nil {
			d.Set("target_host_name", nfs3.Target)
			d.Set("usage_model", nfs3.UsageModel)
		}
		if err := d.Set("namespace_junction", flattenNamespaceJunctions(props.Junctions)); err != nil {
			return fmt.Errorf(`Error setting "namespace_junction" %q (Resource Group %q, Cahe %q): %+v`, id.Name, id.ResourceGroup, id.CacheName, err)
		}
	}

	return nil
}

func resourceHPCCacheNFSTargetDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).HPCCache.StorageTargetsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageTargetID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.CacheName, id.Name, "")
	if err != nil {
		return fmt.Errorf("deleting HPC Cache NFS Target %q (Resource Group %q, Cahe %q): %+v", id.Name, id.ResourceGroup, id.CacheName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of HPC Cache NFS Target %q (Resource Group %q, Cahe %q): %+v", id.Name, id.ResourceGroup, id.CacheName, err)
	}

	return nil
}

func expandNamespaceJunctions(input []interface{}) *[]storagecache.NamespaceJunction {
	result := make([]storagecache.NamespaceJunction, 0)

	for _, v := range input {
		b := v.(map[string]interface{})
		result = append(result, storagecache.NamespaceJunction{
			NamespacePath:   utils.String(b["namespace_path"].(string)),
			NfsExport:       utils.String(b["nfs_export"].(string)),
			TargetPath:      utils.String(b["target_path"].(string)),
			NfsAccessPolicy: utils.String(b["access_policy_name"].(string)),
		})
	}

	return &result
}

func flattenNamespaceJunctions(input *[]storagecache.NamespaceJunction) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make([]interface{}, 0)

	for _, e := range *input {
		namespacePath := ""
		if v := e.NamespacePath; v != nil {
			namespacePath = *v
		}

		nfsExport := ""
		if v := e.NfsExport; v != nil {
			nfsExport = *v
		}

		targetPath := ""
		if v := e.TargetPath; v != nil {
			targetPath = *v
		}

		accessPolicy := ""
		if v := e.NfsAccessPolicy; v != nil {
			accessPolicy = *e.NfsAccessPolicy
		}

		output = append(output, map[string]interface{}{
			"namespace_path":     namespacePath,
			"nfs_export":         nfsExport,
			"target_path":        targetPath,
			"access_policy_name": accessPolicy,
		})
	}

	return output
}
