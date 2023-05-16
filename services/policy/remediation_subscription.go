package policy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"

	"github.com/Azure/azure-sdk-for-go/services/preview/policyinsights/mgmt/2019-10-01-preview/policyinsights"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/policy/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/policy/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceArmSubscriptionPolicyRemediation() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmSubscriptionPolicyRemediationCreateUpdate,
		Read:   resourceArmSubscriptionPolicyRemediationRead,
		Update: resourceArmSubscriptionPolicyRemediationCreateUpdate,
		Delete: resourceArmSubscriptionPolicyRemediationDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SubscriptionPolicyRemediationID(id)
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
				ValidateFunc: validate.RemediationName,
			},

			"subscription_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: commonids.ValidateSubscriptionID,
			},

			"policy_assignment_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				// TODO: remove this suppression when github issue https://github.com/Azure/azure-rest-api-specs/issues/8353 is addressed
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validate.PolicyAssignmentID,
			},

			"location_filters": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: location.EnhancedValidate,
				},
			},

			"policy_definition_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				// TODO: remove this suppression when github issue https://github.com/Azure/azure-rest-api-specs/issues/8353 is addressed
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validate.PolicyDefinitionID,
			},

			"resource_discovery_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(policyinsights.ExistingNonCompliant),
				ValidateFunc: validation.StringInSlice([]string{
					string(policyinsights.ExistingNonCompliant),
					string(policyinsights.ReEvaluateCompliance),
				}, false),
			},
		},
	}
}

func resourceArmSubscriptionPolicyRemediationCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Policy.RemediationsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	subscriptionId, err := commonids.ParseSubscriptionID(d.Get("subscription_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewSubscriptionPolicyRemediationID(subscriptionId.SubscriptionId, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.GetAtSubscription(ctx, id.SubscriptionId, id.RemediationName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id.ID(), err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_subscription_policy_remediation", *existing.ID)
		}
	}

	parameters := policyinsights.Remediation{
		RemediationProperties: &policyinsights.RemediationProperties{
			Filters: &policyinsights.RemediationFilters{
				Locations: utils.ExpandStringSlice(d.Get("location_filters").([]interface{})),
			},
			PolicyAssignmentID:          utils.String(d.Get("policy_assignment_id").(string)),
			PolicyDefinitionReferenceID: utils.String(d.Get("policy_definition_id").(string)),
			ResourceDiscoveryMode:       policyinsights.ResourceDiscoveryMode(d.Get("resource_discovery_mode").(string)),
		},
	}

	if _, err = client.CreateOrUpdateAtSubscription(ctx, id.SubscriptionId, id.RemediationName, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id.ID(), err)
	}

	d.SetId(id.ID())

	return resourceArmSubscriptionPolicyRemediationRead(d, meta)
}

func resourceArmSubscriptionPolicyRemediationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Policy.RemediationsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionPolicyRemediationID(d.Id())
	if err != nil {
		return fmt.Errorf("reading Policy Remediation: %+v", err)
	}

	subscriptionId := commonids.NewSubscriptionID(id.SubscriptionId)

	resp, err := client.GetAtSubscription(ctx, id.SubscriptionId, id.RemediationName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s does not exist - removing from state", id.ID())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading %s: %+v", id.ID(), err)
	}

	d.Set("name", id.RemediationName)
	d.Set("subscription_id", subscriptionId.ID())

	if props := resp.RemediationProperties; props != nil {
		locations := []interface{}{}
		if filters := props.Filters; filters != nil {
			locations = utils.FlattenStringSlice(filters.Locations)
		}
		if err := d.Set("location_filters", locations); err != nil {
			return fmt.Errorf("setting `location_filters`: %+v", err)
		}

		d.Set("policy_assignment_id", props.PolicyAssignmentID)
		d.Set("policy_definition_id", props.PolicyDefinitionReferenceID)
		d.Set("resource_discovery_mode", string(props.ResourceDiscoveryMode))
	}

	return nil
}

func resourceArmSubscriptionPolicyRemediationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Policy.RemediationsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SubscriptionPolicyRemediationID(d.Id())
	if err != nil {
		return err
	}

	// we have to cancel the remediation first before deleting it when the resource_discovery_mode is set to ReEvaluateCompliance
	// therefore we first retrieve the remediation to see if the resource_discovery_mode is switched to ReEvaluateCompliance
	existing, err := client.GetAtSubscription(ctx, id.SubscriptionId, id.RemediationName)
	if err != nil {
		if utils.ResponseWasNotFound(existing.Response) {
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	if existing.RemediationProperties != nil && existing.RemediationProperties.ResourceDiscoveryMode == policyinsights.ReEvaluateCompliance {
		// Remediation can only be canceld when it is in "Evaluating" status, otherwise, API might raise error (e.g. canceling a "Completed" remediation returns 400).
		if existing.RemediationProperties.ProvisioningState != nil && *existing.RemediationProperties.ProvisioningState == "Evaluating" {
			log.Printf("[DEBUG] cancelling the remediation first before deleting it when `resource_discovery_mode` is set to `ReEvaluateCompliance`")
			if _, err := client.CancelAtSubscription(ctx, id.SubscriptionId, id.RemediationName); err != nil {
				return fmt.Errorf("cancelling %s: %+v", id.ID(), err)
			}

			log.Printf("[DEBUG] waiting for the %s to be canceled", id.ID())
			stateConf := &pluginsdk.StateChangeConf{
				Pending: []string{"Cancelling"},
				Target: []string{
					"Succeeded", "Canceled", "Failed",
				},
				Refresh:    subscriptionPolicyRemediationCancellationRefreshFunc(ctx, client, *id),
				MinTimeout: 10 * time.Second,
				Timeout:    d.Timeout(pluginsdk.TimeoutDelete),
			}

			if _, err := stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("waiting for %s to be canceled: %+v", id.ID(), err)
			}
		}
	}

	_, err = client.DeleteAtSubscription(ctx, id.SubscriptionId, id.RemediationName)

	return err
}

func subscriptionPolicyRemediationCancellationRefreshFunc(ctx context.Context, client *policyinsights.RemediationsClient, id parse.SubscriptionPolicyRemediationId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.GetAtSubscription(ctx, id.SubscriptionId, id.RemediationName)
		if err != nil {
			return nil, "", fmt.Errorf("issuing read request for %s: %+v", id.ID(), err)
		}

		if resp.RemediationProperties == nil {
			return nil, "", fmt.Errorf("`properties` was nil")
		}
		if resp.RemediationProperties.ProvisioningState == nil {
			return nil, "", fmt.Errorf("`properties.ProvisioningState` was nil")
		}
		return resp, *resp.RemediationProperties.ProvisioningState, nil
	}
}
