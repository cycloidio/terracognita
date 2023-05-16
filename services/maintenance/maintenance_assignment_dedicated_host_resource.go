package maintenance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/maintenance/mgmt/2021-05-01/maintenance"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	parseCompute "github.com/hashicorp/terraform-provider-azurerm/services/compute/parse"
	validateCompute "github.com/hashicorp/terraform-provider-azurerm/services/compute/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/maintenance/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/maintenance/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceArmMaintenanceAssignmentDedicatedHost() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmMaintenanceAssignmentDedicatedHostCreate,
		Read:   resourceArmMaintenanceAssignmentDedicatedHostRead,
		Delete: resourceArmMaintenanceAssignmentDedicatedHostDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.MaintenanceAssignmentDedicatedHostID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"location": azure.SchemaLocation(),

			"maintenance_configuration_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.MaintenanceConfigurationID,
			},

			"dedicated_host_id": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validateCompute.DedicatedHostID,
				DiffSuppressFunc: suppress.CaseDifference,
			},
		},
	}
}

func resourceArmMaintenanceAssignmentDedicatedHostCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maintenance.ConfigurationAssignmentsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	dedicatedHostIdRaw := d.Get("dedicated_host_id").(string)
	dedicatedHostId, _ := parseCompute.DedicatedHostID(dedicatedHostIdRaw)

	existingList, err := getMaintenanceAssignmentDedicatedHost(ctx, client, dedicatedHostId, dedicatedHostIdRaw)
	if err != nil {
		return err
	}
	if existingList != nil && len(*existingList) > 0 {
		existing := (*existingList)[0]
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_maintenance_assignment_dedicated_host", *existing.ID)
		}
	}

	maintenanceConfigurationID := d.Get("maintenance_configuration_id").(string)
	configurationId, _ := parse.MaintenanceConfigurationIDInsensitively(maintenanceConfigurationID)

	// set assignment name to configuration name
	assignmentName := configurationId.Name
	configurationAssignment := maintenance.ConfigurationAssignment{
		Name:     utils.String(assignmentName),
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		ConfigurationAssignmentProperties: &maintenance.ConfigurationAssignmentProperties{
			MaintenanceConfigurationID: utils.String(maintenanceConfigurationID),
			ResourceID:                 utils.String(dedicatedHostIdRaw),
		},
	}

	// It may take a few minutes after starting a VM for it to become available to assign to a configuration
	err = pluginsdk.Retry(d.Timeout(pluginsdk.TimeoutCreate), func() *pluginsdk.RetryError {
		if _, err := client.CreateOrUpdateParent(ctx, dedicatedHostId.ResourceGroup, "Microsoft.Compute", "hostGroups", dedicatedHostId.HostGroupName, "hosts", dedicatedHostId.HostName, assignmentName, configurationAssignment); err != nil {
			if strings.Contains(err.Error(), "It may take a few minutes after starting a VM for it to become available to assign to a configuration") {
				return pluginsdk.RetryableError(fmt.Errorf("expected VM is available to assign to a configuration but was in pending state, retrying"))
			}
			return pluginsdk.NonRetryableError(fmt.Errorf("issuing creating request for Maintenance Assignment (Dedicated Host ID %q): %+v", dedicatedHostIdRaw, err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	resp, err := getMaintenanceAssignmentDedicatedHost(ctx, client, dedicatedHostId, dedicatedHostIdRaw)
	if err != nil {
		return err
	}
	if resp == nil || len(*resp) == 0 {
		return fmt.Errorf("could not find Maintenance assignment (virtual machine scale set ID: %q)", dedicatedHostIdRaw)
	}
	assignment := (*resp)[0]
	if assignment.ID == nil || *assignment.ID == "" {
		return fmt.Errorf("empty or nil ID of Maintenance Assignment (Dedicated Host ID %q)", dedicatedHostIdRaw)
	}

	d.SetId(*assignment.ID)
	return resourceArmMaintenanceAssignmentDedicatedHostRead(d, meta)
}

func resourceArmMaintenanceAssignmentDedicatedHostRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maintenance.ConfigurationAssignmentsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MaintenanceAssignmentDedicatedHostID(d.Id())
	if err != nil {
		return err
	}

	resp, err := getMaintenanceAssignmentDedicatedHost(ctx, client, id.DedicatedHostId, id.DedicatedHostIdRaw)
	if err != nil {
		return err
	}
	if resp == nil || len(*resp) == 0 {
		d.SetId("")
		return nil
	}
	assignment := (*resp)[0]
	if assignment.ID == nil || *assignment.ID == "" {
		return fmt.Errorf("empty or nil ID of Maintenance Assignment (Dedicated Host ID: %q", id.DedicatedHostIdRaw)
	}

	dedicatedHostId := ""
	if id.DedicatedHostId != nil {
		dedicatedHostId = id.DedicatedHostId.ID()
	}
	d.Set("dedicated_host_id", dedicatedHostId)

	if props := assignment.ConfigurationAssignmentProperties; props != nil {
		d.Set("maintenance_configuration_id", props.MaintenanceConfigurationID)
	}
	return nil
}

func resourceArmMaintenanceAssignmentDedicatedHostDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Maintenance.ConfigurationAssignmentsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MaintenanceAssignmentDedicatedHostID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.DeleteParent(ctx, id.DedicatedHostId.ResourceGroup, "Microsoft.Compute", "hostGroups", id.DedicatedHostId.HostGroupName, "hosts", id.DedicatedHostId.HostName, id.Name); err != nil {
		return fmt.Errorf("deleting Maintenance Assignment to resource %q: %+v", id.DedicatedHostIdRaw, err)
	}

	return nil
}

func getMaintenanceAssignmentDedicatedHost(ctx context.Context, client *maintenance.ConfigurationAssignmentsClient, id *parseCompute.DedicatedHostId, dedicatedHostId string) (result *[]maintenance.ConfigurationAssignment, err error) {
	resp, err := client.ListParent(ctx, id.ResourceGroup, "Microsoft.Compute", "hostGroups", id.HostGroupName, "hosts", id.HostName)
	if err != nil {
		if !utils.ResponseWasNotFound(resp.Response) {
			err = fmt.Errorf("checking for presence of existing Maintenance assignment (Dedicated Host ID %q): %+v", dedicatedHostId, err)
			return
		}
	}
	return resp.Value, nil
}
