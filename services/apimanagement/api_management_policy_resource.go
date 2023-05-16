package apimanagement

import (
	"fmt"
	"html"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceApiManagementPolicy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementPolicyCreateUpdate,
		Read:   resourceApiManagementPolicyRead,
		Update: resourceApiManagementPolicyCreateUpdate,
		Delete: resourceApiManagementPolicyDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.PolicyID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ApiManagementApiPolicyV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"api_management_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"xml_content": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"xml_link"},
				ExactlyOneOf:     []string{"xml_link", "xml_content"},
				DiffSuppressFunc: XmlWithDotNetInterpolationsDiffSuppress,
			},

			"xml_link": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"xml_content"},
				ExactlyOneOf:  []string{"xml_link", "xml_content"},
			},
		},
	}
}

func resourceApiManagementPolicyCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.PolicyClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	apiManagementID := d.Get("api_management_id").(string)
	apiMgmtId, err := parse.ApiManagementID(apiManagementID)
	if err != nil {
		return err
	}
	resourceGroup := apiMgmtId.ResourceGroup
	serviceName := apiMgmtId.ServiceName

	/*
		Other resources would have a check for d.IsNewResource() at this location, and would error out using `tf.ImportAsExistsError` if the resource already existed.
		However, this is a sub-resource, and the API always returns a policy when queried, either a default policy or one configured by the user or by this pluginsdk.
		Instead of the usual check, the resource documentation clearly states that any existing policy will be overwritten if the resource is used.
	*/

	parameters := apimanagement.PolicyContract{}

	xmlContent := d.Get("xml_content").(string)
	xmlLink := d.Get("xml_link").(string)

	if xmlLink != "" {
		parameters.PolicyContractProperties = &apimanagement.PolicyContractProperties{
			Format: apimanagement.PolicyContentFormatRawxmlLink,
			Value:  utils.String(xmlLink),
		}
	} else if xmlContent != "" {
		// this is intentionally an else-if since `xml_content` is computed

		// clear out any existing value for xml_link
		if !d.IsNewResource() {
			d.Set("xml_link", "")
		}

		parameters.PolicyContractProperties = &apimanagement.PolicyContractProperties{
			Format: apimanagement.PolicyContentFormatRawxml,
			Value:  utils.String(xmlContent),
		}
	}

	if parameters.PolicyContractProperties == nil {
		return fmt.Errorf("Either `xml_content` or `xml_link` must be set")
	}

	_, err = client.CreateOrUpdate(ctx, resourceGroup, serviceName, parameters, "")
	if err != nil {
		return fmt.Errorf("creating or updating Policy (Resource Group %q / API Management Service %q): %+v", resourceGroup, serviceName, err)
	}

	id := parse.NewPolicyID(apiMgmtId.SubscriptionId, apiMgmtId.ResourceGroup, apiMgmtId.ServiceName, "policy")
	d.SetId(id.ID())

	return resourceApiManagementPolicyRead(d, meta)
}

func resourceApiManagementPolicyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	serviceClient := meta.(*clients.Client).ApiManagement.ServiceClient
	client := meta.(*clients.Client).ApiManagement.PolicyClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PolicyID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serviceName := id.ServiceName

	serviceResp, err := serviceClient.Get(ctx, resourceGroup, serviceName)
	if err != nil {
		if utils.ResponseWasNotFound(serviceResp.Response) {
			log.Printf("API Management Service %q was not found in Resource Group %q - removing Policy from state!", serviceName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request on API Management Service %q (Resource Group %q): %+v", serviceName, resourceGroup, err)
	}

	d.Set("api_management_id", serviceResp.ID)

	resp, err := client.Get(ctx, resourceGroup, serviceName, apimanagement.PolicyExportFormatXML)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Policy (Resource Group %q / API Management Service %q) was not found - removing from state!", resourceGroup, serviceName)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request for Policy (Resource Group %q / API Management Service %q): %+v", resourceGroup, serviceName, err)
	}

	if properties := resp.PolicyContractProperties; properties != nil {
		policyContent := ""
		if pc := properties.Value; pc != nil {
			policyContent = html.UnescapeString(*pc)
		}

		// when you submit an `xml_link` to the API, the API downloads this link and stores it as `xml_content`
		// as such there is no way to set `xml_link` and we'll let Terraform handle it
		d.Set("xml_content", policyContent)
	}

	return nil
}

func resourceApiManagementPolicyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.PolicyClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PolicyID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serviceName := id.ServiceName

	if resp, err := client.Delete(ctx, resourceGroup, serviceName, ""); err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting Policy (Resource Group %q / API Management Service %q): %+v", resourceGroup, serviceName, err)
		}
	}

	return nil
}
