package azurestackhci

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/azurestackhci/mgmt/2020-10-01/azurestackhci"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/azurestackhci/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/azurestackhci/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceArmStackHCICluster() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmStackHCIClusterCreate,
		Read:   resourceArmStackHCIClusterRead,
		Update: resourceArmStackHCIClusterUpdate,
		Delete: resourceArmStackHCIClusterDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ClusterID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ClusterName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"client_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"tenant_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmStackHCIClusterCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AzureStackHCI.ClusterClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	tenantId := meta.(*clients.Client).Account.TenantId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewClusterID(subscriptionId, resourceGroup, name)

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Azure Stack HCI Cluster %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_stack_hci_cluster", id.ID())
	}

	cluster := azurestackhci.Cluster{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		ClusterProperties: &azurestackhci.ClusterProperties{
			AadClientID: utils.String(d.Get("client_id").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("tenant_id"); ok {
		cluster.ClusterProperties.AadTenantID = utils.String(v.(string))
	} else {
		cluster.ClusterProperties.AadTenantID = utils.String(tenantId)
	}

	if _, err := client.Create(ctx, resourceGroup, name, cluster); err != nil {
		return fmt.Errorf("creating Azure Stack HCI Cluster %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id.ID())

	return resourceArmStackHCIClusterRead(d, meta)
}

func resourceArmStackHCIClusterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AzureStackHCI.ClusterClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Azure Stack HCI Cluster %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Azure Stack HCI Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.ClusterProperties; props != nil {
		d.Set("client_id", props.AadClientID)
		d.Set("tenant_id", props.AadTenantID)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmStackHCIClusterUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AzureStackHCI.ClusterClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	cluster := azurestackhci.ClusterUpdate{}

	if d.HasChange("tags") {
		cluster.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.Name, cluster); err != nil {
		return fmt.Errorf("updating Azure Stack HCI Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return resourceArmStackHCIClusterRead(d, meta)
}

func resourceArmStackHCIClusterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AzureStackHCI.ClusterClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return fmt.Errorf("deleting Azure Stack HCI Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}
