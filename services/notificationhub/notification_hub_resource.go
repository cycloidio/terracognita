package notificationhub

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/notificationhubs/mgmt/2017-04-01/notificationhubs"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/notificationhub/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/notificationhub/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var notificationHubResourceName = "azurerm_notification_hub"

const (
	apnsProductionName     = "Production"
	apnsProductionEndpoint = "https://api.push.apple.com:443/3/device"
	apnsSandboxName        = "Sandbox"
	apnsSandboxEndpoint    = "https://api.development.push.apple.com:443/3/device"
)

func resourceNotificationHub() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNotificationHubCreateUpdate,
		Read:   resourceNotificationHubRead,
		Update: resourceNotificationHubCreateUpdate,
		Delete: resourceNotificationHubDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NotificationHubID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.NotificationHubResourceV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: pluginsdk.CustomizeDiffShim(func(ctx context.Context, diff *pluginsdk.ResourceDiff, v interface{}) error {
			// NOTE: the ForceNew is to workaround a bug in the Azure SDK where nil-values aren't sent to the API.
			// Bug: https://github.com/Azure/azure-sdk-for-go/issues/2246

			oAPNS, nAPNS := diff.GetChange("apns_credential.#")
			oAPNSi := oAPNS.(int)
			nAPNSi := nAPNS.(int)
			if nAPNSi < oAPNSi {
				diff.ForceNew("apns_credential")
			}

			oGCM, nGCM := diff.GetChange("gcm_credential.#")
			oGCMi := oGCM.(int)
			nGCMi := nGCM.(int)
			if nGCMi < oGCMi {
				diff.ForceNew("gcm_credential")
			}

			return nil
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"namespace_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"apns_credential": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						// NOTE: APNS supports two modes, certificate auth (v1) and token auth (v2)
						// certificate authentication/v1 is marked for deprecation; as such we're not
						// supporting it at this time.
						"application_mode": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								apnsProductionName,
								apnsSandboxName,
							}, false),
						},
						"bundle_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"key_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						// Team ID (within Apple & the Portal) == "AppID" (within the API)
						"team_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"token": {
							Type:      pluginsdk.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},

			"gcm_credential": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"api_key": {
							Type:      pluginsdk.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceNotificationHubCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NotificationHubs.HubsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewNotificationHubID(subscriptionId, d.Get("resource_group_name").(string), d.Get("namespace_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_notification_hub", id.ID())
		}
	}

	parameters := notificationhubs.CreateOrUpdateParameters{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		Properties: &notificationhubs.Properties{
			ApnsCredential: expandNotificationHubsAPNSCredentials(d.Get("apns_credential").([]interface{})),
			GcmCredential:  expandNotificationHubsGCMCredentials(d.Get("gcm_credential").([]interface{})),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.NamespaceName, id.Name, parameters); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	// Notification Hubs are eventually consistent
	log.Printf("[DEBUG] Waiting for %s to become available..", id)
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context had no deadline")
	}
	stateConf := &pluginsdk.StateChangeConf{
		Pending:                   []string{"404"},
		Target:                    []string{"200"},
		Refresh:                   notificationHubStateRefreshFunc(ctx, client, id),
		MinTimeout:                15 * time.Second,
		ContinuousTargetOccurence: 10,
		Timeout:                   time.Until(deadline),
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to become available: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceNotificationHubRead(d, meta)
}

func notificationHubStateRefreshFunc(ctx context.Context, client *notificationhubs.Client, id parse.NotificationHubId) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return nil, "404", nil
			}

			return nil, "", fmt.Errorf("retrieving %s: %+v", id, err)
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}

func resourceNotificationHubRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NotificationHubs.HubsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NotificationHubID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	credentials, err := client.GetPnsCredentials(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving credentials for %s: %+v", *id, err)
	}

	d.Set("name", resp.Name)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := credentials.PnsCredentialsProperties; props != nil {
		apns := flattenNotificationHubsAPNSCredentials(props.ApnsCredential)
		if setErr := d.Set("apns_credential", apns); setErr != nil {
			return fmt.Errorf("setting `apns_credential`: %+v", setErr)
		}

		gcm := flattenNotificationHubsGCMCredentials(props.GcmCredential)
		if setErr := d.Set("gcm_credential", gcm); setErr != nil {
			return fmt.Errorf("setting `gcm_credential`: %+v", setErr)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceNotificationHubDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NotificationHubs.HubsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NotificationHubID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.NamespaceName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting %s: %+v", *id, err)
		}
	}

	return nil
}

func expandNotificationHubsAPNSCredentials(inputs []interface{}) *notificationhubs.ApnsCredential {
	if len(inputs) == 0 {
		return nil
	}

	input := inputs[0].(map[string]interface{})
	applicationMode := input["application_mode"].(string)
	bundleId := input["bundle_id"].(string)
	keyId := input["key_id"].(string)
	teamId := input["team_id"].(string)
	token := input["token"].(string)

	applicationEndpoints := map[string]string{
		apnsProductionName: apnsProductionEndpoint,
		apnsSandboxName:    apnsSandboxEndpoint,
	}
	endpoint := applicationEndpoints[applicationMode]

	credentials := notificationhubs.ApnsCredential{
		ApnsCredentialProperties: &notificationhubs.ApnsCredentialProperties{
			AppID:    utils.String(teamId),
			AppName:  utils.String(bundleId),
			Endpoint: utils.String(endpoint),
			KeyID:    utils.String(keyId),
			Token:    utils.String(token),
		},
	}
	return &credentials
}

func flattenNotificationHubsAPNSCredentials(input *notificationhubs.ApnsCredential) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	output := make(map[string]interface{})

	if bundleId := input.AppName; bundleId != nil {
		output["bundle_id"] = *bundleId
	}

	if endpoint := input.Endpoint; endpoint != nil {
		applicationEndpoints := map[string]string{
			apnsProductionEndpoint: apnsProductionName,
			apnsSandboxEndpoint:    apnsSandboxName,
		}
		applicationMode := applicationEndpoints[*endpoint]
		output["application_mode"] = applicationMode
	}

	if keyId := input.KeyID; keyId != nil {
		output["key_id"] = *keyId
	}

	if teamId := input.AppID; teamId != nil {
		output["team_id"] = *teamId
	}

	if token := input.Token; token != nil {
		output["token"] = *token
	}

	return []interface{}{output}
}

func expandNotificationHubsGCMCredentials(inputs []interface{}) *notificationhubs.GcmCredential {
	if len(inputs) == 0 {
		return nil
	}

	input := inputs[0].(map[string]interface{})
	apiKey := input["api_key"].(string)
	credentials := notificationhubs.GcmCredential{
		GcmCredentialProperties: &notificationhubs.GcmCredentialProperties{
			GoogleAPIKey: utils.String(apiKey),
		},
	}
	return &credentials
}

func flattenNotificationHubsGCMCredentials(input *notificationhubs.GcmCredential) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make(map[string]interface{})
	if props := input.GcmCredentialProperties; props != nil {
		if apiKey := props.GoogleAPIKey; apiKey != nil {
			output["api_key"] = *apiKey
		}
	}

	return []interface{}{output}
}
