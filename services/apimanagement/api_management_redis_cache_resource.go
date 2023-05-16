package apimanagement

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceApiManagementRedisCache() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementRedisCacheCreateUpdate,
		Read:   resourceApiManagementRedisCacheRead,
		Update: resourceApiManagementRedisCacheCreateUpdate,
		Delete: resourceApiManagementRedisCacheDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.RedisCacheID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApiManagementChildName,
			},

			"api_management_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApiManagementID,
			},

			"connection_string": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"redis_cache_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"cache_location": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Default:          "default",
				ValidateFunc:     validate.RedisCacheLocation,
				DiffSuppressFunc: location.DiffSuppressFunc,
			},
		},
	}
}

func resourceApiManagementRedisCacheCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	apimId, err := parse.ApiManagementID(d.Get("api_management_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewRedisCacheID(subscriptionId, apimId.ResourceGroup, apimId.ServiceName, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, apimId.ResourceGroup, apimId.ServiceName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for existing %q: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_api_management_redis_cache", id.ID())
		}
	}

	parameters := apimanagement.CacheContract{
		CacheContractProperties: &apimanagement.CacheContractProperties{
			ConnectionString: utils.String(d.Get("connection_string").(string)),
			UseFromLocation:  utils.String(location.Normalize(d.Get("cache_location").(string))),
		},
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		parameters.CacheContractProperties.Description = utils.String(v.(string))
	}

	// Remove the extra / in the ResourceID so the redis cache can be associated to the api mgmt service as expected,
	// otherwise, the resourceId behave like
	// "https://management.azure.com//subscriptions/xx/resourceGroups/xx/providers/Microsoft.Cache/Redis/xx"
	if v, ok := d.GetOk("redis_cache_id"); ok && v.(string) != "" {
		parameters.CacheContractProperties.ResourceID = utils.String(strings.TrimSuffix(meta.(*clients.Client).Account.Environment.ResourceManagerEndpoint, "/") + v.(string))
	}

	// here we use "PUT" for updating, because `description` is not allowed to be empty string, Then we could not update to remove `description` by `PATCH`
	if _, err := client.CreateOrUpdate(ctx, apimId.ResourceGroup, apimId.ServiceName, name, parameters, ""); err != nil {
		return fmt.Errorf("creating/ updating %q: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceApiManagementRedisCacheRead(d, meta)
}

func resourceApiManagementRedisCacheRead(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RedisCacheID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.CacheName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] apimanagement %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %q: %+v", id, err)
	}
	d.Set("name", id.CacheName)
	d.Set("api_management_id", parse.NewApiManagementID(subscriptionId, id.ResourceGroup, id.ServiceName).ID())
	if props := resp.CacheContractProperties; props != nil {
		d.Set("description", props.Description)

		cacheId := ""
		if props.ResourceID != nil {
			// correct the resourceID issue: "https://management.azure.com//subscriptions/xx/resourceGroups/xx/providers/Microsoft.Cache/Redis/xx"
			cacheId = strings.TrimPrefix(*props.ResourceID, strings.TrimSuffix(meta.(*clients.Client).Account.Environment.ResourceManagerEndpoint, "/"))
		}
		d.Set("redis_cache_id", cacheId)
		d.Set("cache_location", props.UseFromLocation)
	}
	return nil
}

func resourceApiManagementRedisCacheDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RedisCacheID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.ServiceName, id.CacheName, "*"); err != nil {
		return fmt.Errorf("deleting %q: %+v", id, err)
	}
	return nil
}
