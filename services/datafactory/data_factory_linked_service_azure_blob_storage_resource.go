package datafactory

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datafactory/mgmt/2018-06-01/datafactory"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/datafactory/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/datafactory/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDataFactoryLinkedServiceAzureBlobStorage() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDataFactoryLinkedServiceBlobStorageCreateUpdate,
		Read:   resourceDataFactoryLinkedServiceBlobStorageRead,
		Update: resourceDataFactoryLinkedServiceBlobStorageCreateUpdate,
		Delete: resourceDataFactoryLinkedServiceBlobStorageDelete,

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.LinkedServiceID(id)
			return err
		}, importDataFactoryLinkedService(datafactory.TypeBasicLinkedServiceTypeAzureBlobStorage)),

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
				ValidateFunc: validate.LinkedServiceDatasetName,
			},

			"data_factory_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DataFactoryID,
			},

			"connection_string": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				ExactlyOneOf: []string{"connection_string", "sas_uri", "service_endpoint"},
			},

			"storage_kind": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Storage",
					"StorageV2",
					"BlobStorage",
					"BlockBlobStorage",
				}, false),
			},

			"sas_uri": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				ExactlyOneOf: []string{"connection_string", "sas_uri", "service_endpoint"},
			},

			// TODO for @favoretti: rename this to 'sas_token_linked_key_vault_key' for 3.4.0
			"key_vault_sas_token": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"linked_service_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"secret_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"service_principal_linked_key_vault_key": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"linked_service_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"secret_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"integration_runtime_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"use_managed_identity": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
				ConflictsWith: []string{
					"service_principal_id",
				},
			},

			"service_endpoint": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				ExactlyOneOf: []string{"connection_string", "sas_uri", "service_endpoint"},
			},

			"service_principal_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsUUID,
				ConflictsWith: []string{
					"use_managed_identity",
				},
			},

			"service_principal_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"tenant_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"parameters": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"annotations": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"additional_properties": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func resourceDataFactoryLinkedServiceBlobStorageCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataFactory.LinkedServiceClient
	subscriptionId := meta.(*clients.Client).DataFactory.LinkedServiceClient.SubscriptionID
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	dataFactoryId, err := parse.DataFactoryID(d.Get("data_factory_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewLinkedServiceID(subscriptionId, dataFactoryId.ResourceGroup, dataFactoryId.FactoryName, d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.FactoryName, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Data Factory Blob Storage Anonymous %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_data_factory_linked_service_azure_blob_storage", id.ID())
		}
	}

	blobStorageProperties := &datafactory.AzureBlobStorageLinkedServiceTypeProperties{}

	if v, ok := d.GetOk("connection_string"); ok {
		blobStorageProperties.ConnectionString = &datafactory.SecureString{
			Value: utils.String(v.(string)),
			Type:  datafactory.TypeSecureString,
		}
	}

	if v, ok := d.GetOk("sas_uri"); ok {
		if sasToken, ok := d.GetOk("key_vault_sas_token"); ok {
			blobStorageProperties.SasURI = utils.String(v.(string))
			blobStorageProperties.SasToken = expandAzureKeyVaultSecretReference(sasToken.([]interface{}))
		} else {
			blobStorageProperties.SasURI = &datafactory.SecureString{
				Value: utils.String(v.(string)),
				Type:  datafactory.TypeSecureString,
			}
		}
	}

	if d.Get("use_managed_identity").(bool) {
		if v, ok := d.GetOk("service_endpoint"); ok {
			blobStorageProperties.ServiceEndpoint = utils.String(v.(string))
		}
	} else {
		if v, ok := d.GetOk("service_endpoint"); ok {
			blobStorageProperties.ServiceEndpoint = utils.String(v.(string))
		}
		if kvsp, ok := d.GetOk("service_principal_linked_key_vault_key"); ok {
			blobStorageProperties.ServicePrincipalKey = expandAzureKeyVaultSecretReference(kvsp.([]interface{}))
		} else {
			secureString := datafactory.SecureString{
				Value: utils.String(d.Get("service_principal_key").(string)),
				Type:  datafactory.TypeSecureString,
			}
			blobStorageProperties.ServicePrincipalKey = &secureString
		}

		blobStorageProperties.ServicePrincipalID = utils.String(d.Get("service_principal_id").(string))
		blobStorageProperties.Tenant = utils.String(d.Get("tenant_id").(string))

	}

	blobStorageLinkedService := &datafactory.AzureBlobStorageLinkedService{
		Description: utils.String(d.Get("description").(string)),
		AzureBlobStorageLinkedServiceTypeProperties: blobStorageProperties,
		Type: datafactory.TypeBasicLinkedServiceTypeAzureBlobStorage,
	}

	if v, ok := d.GetOk("parameters"); ok {
		blobStorageLinkedService.Parameters = expandDataFactoryParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("integration_runtime_name"); ok {
		blobStorageLinkedService.ConnectVia = expandDataFactoryLinkedServiceIntegrationRuntime(v.(string))
	}

	if v, ok := d.GetOk("storage_kind"); ok {
		blobStorageLinkedService.AccountKind = utils.String(v.(string))
	}

	if v, ok := d.GetOk("additional_properties"); ok {
		blobStorageLinkedService.AdditionalProperties = v.(map[string]interface{})
	}

	if v, ok := d.GetOk("annotations"); ok {
		annotations := v.([]interface{})
		blobStorageLinkedService.Annotations = &annotations
	}

	linkedService := datafactory.LinkedServiceResource{
		Properties: blobStorageLinkedService,
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.FactoryName, id.Name, linkedService, ""); err != nil {
		return fmt.Errorf("creating/updating Data Factory Blob Storage %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceDataFactoryLinkedServiceBlobStorageRead(d, meta)
}

func resourceDataFactoryLinkedServiceBlobStorageRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataFactory.LinkedServiceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LinkedServiceID(d.Id())
	if err != nil {
		return err
	}

	dataFactoryId := parse.NewDataFactoryID(id.SubscriptionId, id.ResourceGroup, id.FactoryName)

	resp, err := client.Get(ctx, id.ResourceGroup, id.FactoryName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Data Factory Blob Storage %s: %+v", *id, err)
	}

	d.Set("name", resp.Name)
	d.Set("data_factory_id", dataFactoryId.ID())

	blobStorage, ok := resp.Properties.AsAzureBlobStorageLinkedService()
	if !ok {
		return fmt.Errorf("classifying Data Factory Blob Storage %s: Expected: %q Received: %q", *id, datafactory.TypeBasicLinkedServiceTypeAzureBlobStorage, *resp.Type)
	}

	if blobStorage != nil {
		if blobStorage.Tenant != nil {
			d.Set("tenant_id", blobStorage.Tenant)
		}

		if blobStorage.ServicePrincipalID != nil {
			d.Set("service_principal_id", blobStorage.ServicePrincipalID)
			d.Set("use_managed_identity", false)
		} else {
			d.Set("service_endpoint", blobStorage.ServiceEndpoint)
			d.Set("use_managed_identity", true)
		}
	}

	if properties := blobStorage.AzureBlobStorageLinkedServiceTypeProperties; properties != nil {
		d.Set("storage_kind", properties.AccountKind)
		if sasToken := properties.SasToken; sasToken != nil {
			if keyVaultPassword, ok := sasToken.AsAzureKeyVaultSecretReference(); ok {
				if err := d.Set("key_vault_sas_token", flattenAzureKeyVaultSecretReference(keyVaultPassword)); err != nil {
					return fmt.Errorf("Error setting `key_vault_sas_token`: %+v", err)
				}
			}
		}

		if spKey := properties.ServicePrincipalKey; spKey != nil {
			if kvSPkey, ok := spKey.AsAzureKeyVaultSecretReference(); ok {
				if err := d.Set("service_principal_linked_key_vault_key", flattenAzureKeyVaultSecretReference(kvSPkey)); err != nil {
					return fmt.Errorf("Error setting `service_principal_linked_key_vault_key`: %+v", err)
				}
			}
		}
	}

	d.Set("additional_properties", blobStorage.AdditionalProperties)
	d.Set("description", blobStorage.Description)

	annotations := flattenDataFactoryAnnotations(blobStorage.Annotations)
	if err := d.Set("annotations", annotations); err != nil {
		return fmt.Errorf("setting `annotations` for Data Factory Azure Blob Storage %s: %+v", *id, err)
	}

	parameters := flattenDataFactoryParameters(blobStorage.Parameters)
	if err := d.Set("parameters", parameters); err != nil {
		return fmt.Errorf("setting `parameters`: %+v", err)
	}

	if connectVia := blobStorage.ConnectVia; connectVia != nil {
		if connectVia.ReferenceName != nil {
			d.Set("integration_runtime_name", connectVia.ReferenceName)
		}
	}

	return nil
}

func resourceDataFactoryLinkedServiceBlobStorageDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataFactory.LinkedServiceClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LinkedServiceID(d.Id())
	if err != nil {
		return err
	}

	response, err := client.Delete(ctx, id.ResourceGroup, id.FactoryName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(response) {
			return fmt.Errorf("deleting Data Factory Blob Storage %s: %+v", *id, err)
		}
	}

	return nil
}
