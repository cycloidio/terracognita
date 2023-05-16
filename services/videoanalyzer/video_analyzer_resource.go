package videoanalyzer

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	msiparse "github.com/hashicorp/terraform-provider-azurerm/services/msi/parse"
	msivalidate "github.com/hashicorp/terraform-provider-azurerm/services/msi/validate"
	storageValidate "github.com/hashicorp/terraform-provider-azurerm/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/videoanalyzer/sdk/2021-05-01-preview/videoanalyzer"
	"github.com/hashicorp/terraform-provider-azurerm/services/videoanalyzer/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceVideoAnalyzer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVideoAnalyzerCreateUpdate,
		Read:   resourceVideoAnalyzerRead,
		Update: resourceVideoAnalyzerCreateUpdate,
		Delete: resourceVideoAnalyzerDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := videoanalyzer.ParseVideoAnalyzerID(id)
			return err
		}),

		DeprecationMessage: `Video Analyzer (Preview) is now Deprecated and will be Retired on 2022-11-30 - as such the 'azurerm_video_analyzer' resource is deprecated and will be removed in v4.0 of the AzureRM Provider`,

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VideoAnalyzerName(),
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"storage_account": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: storageValidate.StorageAccountID,
						},

						"user_assigned_identity_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: msivalidate.UserAssignedIdentityID,
						},
					},
				},
			},

			"identity": commonschema.UserAssignedIdentityRequired(),

			"tags": commonschema.Tags(),
		},
	}
}

func resourceVideoAnalyzerCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).VideoAnalyzer.VideoAnalyzersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := videoanalyzer.NewVideoAnalyzerID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.VideoAnalyzersGet(ctx, id)
		if err != nil {
			if !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for existing %s: %+v", id, err)
			}
		}

		if !response.WasNotFound(existing.HttpResponse) {
			return tf.ImportAsExistsError("azurerm_video_analyzer", id.ID())
		}
	}

	identity, err := expandAzureRmVideoAnalyzerIdentity(d)
	if err != nil {
		return err
	}
	parameters := videoanalyzer.VideoAnalyzer{
		Properties: &videoanalyzer.VideoAnalyzerPropertiesUpdate{
			StorageAccounts: expandVideoAnalyzerStorageAccounts(d),
		},
		Location: azure.NormalizeLocation(d.Get("location").(string)),
		Identity: identity,
		Tags:     expandTags(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.VideoAnalyzersCreateOrUpdate(ctx, id, parameters); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceVideoAnalyzerRead(d, meta)
}

func resourceVideoAnalyzerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).VideoAnalyzer.VideoAnalyzersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := videoanalyzer.ParseVideoAnalyzerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.VideoAnalyzersGet(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			log.Printf("[INFO] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.AccountName)
	d.Set("resource_group_name", id.ResourceGroupName)

	if model := resp.Model; model != nil {
		d.Set("location", azure.NormalizeLocation(model.Location))

		props := resp.Model.Properties
		if props != nil {
			accounts := flattenVideoAnalyzerStorageAccounts(props.StorageAccounts)
			if err := d.Set("storage_account", accounts); err != nil {
				return fmt.Errorf("flattening `storage_account`: %s", err)
			}
		}

		flattenedIdentity, err := flattenAzureRmVideoServiceIdentity(resp.Model.Identity)
		if err != nil {
			return fmt.Errorf("flattening `identity`: %s", err)
		}

		if err := d.Set("identity", flattenedIdentity); err != nil {
			return fmt.Errorf("setting `identity`: %s", err)
		}

		return tags.FlattenAndSet(d, flattenTags(model.Tags))
	}
	return nil
}

func resourceVideoAnalyzerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).VideoAnalyzer.VideoAnalyzersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := videoanalyzer.ParseVideoAnalyzerID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.VideoAnalyzersDelete(ctx, *id); err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	return nil
}

func expandVideoAnalyzerStorageAccounts(d *pluginsdk.ResourceData) *[]videoanalyzer.StorageAccount {
	storageAccountRaw := d.Get("storage_account").([]interface{})[0].(map[string]interface{})

	results := []videoanalyzer.StorageAccount{
		{
			Id: utils.String(storageAccountRaw["id"].(string)),
			Identity: &videoanalyzer.ResourceIdentity{
				UserAssignedIdentity: storageAccountRaw["user_assigned_identity_id"].(string),
			},
		},
	}

	return &results
}

func flattenVideoAnalyzerStorageAccounts(input *[]videoanalyzer.StorageAccount) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)
	for _, storageAccount := range *input {
		storageAccountId := ""
		if storageAccount.Id != nil {
			storageAccountId = *storageAccount.Id
		}

		userAssignedIdentityId := ""
		if storageAccount.Identity != nil {
			userAssignedIdentityId = storageAccount.Identity.UserAssignedIdentity
		}

		results = append(results, map[string]interface{}{
			"id":                        storageAccountId,
			"user_assigned_identity_id": userAssignedIdentityId,
		})
	}

	return results
}

func expandAzureRmVideoAnalyzerIdentity(d *pluginsdk.ResourceData) (*videoanalyzer.VideoAnalyzerIdentity, error) {
	identityRaw := d.Get("identity").([]interface{})
	if identityRaw[0] == nil {
		return nil, fmt.Errorf("an `identity` block is required")
	}
	identity := identityRaw[0].(map[string]interface{})
	result := &videoanalyzer.VideoAnalyzerIdentity{
		Type: identity["type"].(string),
	}
	var identityIdSet []interface{}
	if identityIds, exists := identity["identity_ids"]; exists {
		identityIdSet = identityIds.(*pluginsdk.Set).List()
	}

	userAssignedIdentities := make(map[string]videoanalyzer.UserAssignedManagedIdentity)
	for _, id := range identityIdSet {
		userAssignedIdentities[id.(string)] = videoanalyzer.UserAssignedManagedIdentity{}
	}
	result.UserAssignedIdentities = &userAssignedIdentities

	return result, nil
}

func flattenAzureRmVideoServiceIdentity(identity *videoanalyzer.VideoAnalyzerIdentity) ([]interface{}, error) {
	if identity == nil {
		return make([]interface{}, 0), nil
	}

	identityIds := make([]interface{}, 0)
	if identity.UserAssignedIdentities != nil {
		/*
		   "userAssignedIdentities": {
		     "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/id1": {
		       "principalId": "00000000-0000-0000-0000-000000000000",
		       "clientId": "00000000-0000-0000-0000-000000000000"
		     },
		   }
		*/
		for key := range *identity.UserAssignedIdentities {
			parsedId, err := msiparse.UserAssignedIdentityID(key)
			if err != nil {
				return nil, err
			}
			identityIds = append(identityIds, parsedId.ID())
		}
	}

	return []interface{}{
		map[string]interface{}{
			"type":         identity.Type,
			"identity_ids": identityIds,
		},
	}, nil
}
