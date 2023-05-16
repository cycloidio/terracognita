package network

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/locks"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceNetworkInterfaceApplicationSecurityGroupAssociation() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetworkInterfaceApplicationSecurityGroupAssociationCreate,
		Read:   resourceNetworkInterfaceApplicationSecurityGroupAssociationRead,
		Delete: resourceNetworkInterfaceApplicationSecurityGroupAssociationDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			splitId := strings.Split(id, "|")
			if _, err := parse.NetworkInterfaceID(splitId[0]); err != nil {
				return err
			}
			if _, err := parse.ApplicationSecurityGroupID(splitId[1]); err != nil {
				return err
			}
			return nil
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.NetworkInterfaceApplicationSecurityGroupAssociationV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"network_interface_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"application_security_group_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},
		},
	}
}

func resourceNetworkInterfaceApplicationSecurityGroupAssociationCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.InterfacesClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Network Interface <-> Application Security Group Association creation.")

	networkInterfaceId := d.Get("network_interface_id").(string)
	applicationSecurityGroupId := d.Get("application_security_group_id").(string)

	id, err := parse.NetworkInterfaceID(networkInterfaceId)
	if err != nil {
		return err
	}

	locks.ByName(id.Name, networkInterfaceResourceName)
	defer locks.UnlockByName(id.Name, networkInterfaceResourceName)

	read, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			log.Printf("[INFO] Network Interface %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	props := read.InterfacePropertiesFormat
	if props == nil {
		return fmt.Errorf("Error: `properties` was nil for %s", *id)
	}
	if props.IPConfigurations == nil {
		return fmt.Errorf("Error: `properties.ipConfigurations` was nil for %s", *id)
	}

	info := parseFieldsFromNetworkInterface(*props)
	resourceId := fmt.Sprintf("%s|%s", networkInterfaceId, applicationSecurityGroupId)
	if utils.SliceContainsValue(info.applicationSecurityGroupIDs, applicationSecurityGroupId) {
		return tf.ImportAsExistsError("azurerm_network_interface_application_security_group_association", resourceId)
	}

	info.applicationSecurityGroupIDs = append(info.applicationSecurityGroupIDs, applicationSecurityGroupId)

	read.InterfacePropertiesFormat.IPConfigurations = mapFieldsToNetworkInterface(props.IPConfigurations, info)

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, read)
	if err != nil {
		return fmt.Errorf("updating Application Security Group Association for %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of Application Security Group Association for %s: %+v", *id, err)
	}

	d.SetId(resourceId)

	return resourceNetworkInterfaceApplicationSecurityGroupAssociationRead(d, meta)
}

func resourceNetworkInterfaceApplicationSecurityGroupAssociationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.InterfacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	splitId := strings.Split(d.Id(), "|")
	if len(splitId) != 2 {
		return fmt.Errorf("Expected ID to be in the format {networkInterfaceId}|{applicationSecurityGroupId} but got %q", d.Id())
	}

	nicID, err := parse.NetworkInterfaceID(splitId[0])
	if err != nil {
		return err
	}

	applicationSecurityGroupId := splitId[1]

	read, err := client.Get(ctx, nicID.ResourceGroup, nicID.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", *nicID)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *nicID, err)
	}

	nicProps := read.InterfacePropertiesFormat
	if nicProps == nil {
		return fmt.Errorf("Error: `properties` was nil for %s", *nicID)
	}

	info := parseFieldsFromNetworkInterface(*nicProps)
	exists := false
	for _, groupId := range info.applicationSecurityGroupIDs {
		if groupId == applicationSecurityGroupId {
			exists = true
		}
	}

	if !exists {
		log.Printf("[DEBUG] Association between %s and Application Security Group %q was not found - removing from state!", *nicID, applicationSecurityGroupId)
		d.SetId("")
		return nil
	}

	d.Set("application_security_group_id", applicationSecurityGroupId)
	d.Set("network_interface_id", read.ID)

	return nil
}

func resourceNetworkInterfaceApplicationSecurityGroupAssociationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.InterfacesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	splitId := strings.Split(d.Id(), "|")
	if len(splitId) != 2 {
		return fmt.Errorf("Expected ID to be in the format {networkInterfaceId}|{applicationSecurityGroupId} but got %q", d.Id())
	}

	nicID, err := parse.NetworkInterfaceID(splitId[0])
	if err != nil {
		return err
	}

	applicationSecurityGroupId := splitId[1]

	locks.ByName(nicID.Name, networkInterfaceResourceName)
	defer locks.UnlockByName(nicID.Name, networkInterfaceResourceName)

	read, err := client.Get(ctx, nicID.ResourceGroup, nicID.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			return fmt.Errorf("%s was not found!", *nicID)
		}

		return fmt.Errorf("retrieving  %s: %+v", *nicID, err)
	}

	props := read.InterfacePropertiesFormat
	if props == nil {
		return fmt.Errorf("Error: `properties` was nil for %s", *nicID)
	}

	if props.IPConfigurations == nil {
		return fmt.Errorf("Error: `properties.ipConfigurations` was nil for %s)", *nicID)
	}

	info := parseFieldsFromNetworkInterface(*props)

	applicationSecurityGroupIds := make([]string, 0)
	for _, v := range info.applicationSecurityGroupIDs {
		if v != applicationSecurityGroupId {
			applicationSecurityGroupIds = append(applicationSecurityGroupIds, v)
		}
	}
	info.applicationSecurityGroupIDs = applicationSecurityGroupIds
	read.InterfacePropertiesFormat.IPConfigurations = mapFieldsToNetworkInterface(props.IPConfigurations, info)

	future, err := client.CreateOrUpdate(ctx, nicID.ResourceGroup, nicID.Name, read)
	if err != nil {
		return fmt.Errorf("removing Application Security Group for %s: %+v", *nicID, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for removal of Application Security Group for %s: %+v", *nicID, err)
	}

	return nil
}
