package eventgrid

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventgrid/mgmt/2021-12-01/eventgrid"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/eventgrid/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceEventGridDomain() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceEventGridDomainCreateUpdate,
		Read:   resourceEventGridDomainRead,
		Update: resourceEventGridDomainCreateUpdate,
		Delete: resourceEventGridDomainDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.DomainID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringMatch(
						regexp.MustCompile("^[-a-zA-Z0-9]{3,50}$"),
						"EventGrid domain name must be 3 - 50 characters long, contain only letters, numbers and hyphens.",
					),
				),
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"identity": commonschema.SystemOrUserAssignedIdentityOptional(),

			"input_schema": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(eventgrid.InputSchemaEventGridSchema),
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(eventgrid.InputSchemaCloudEventSchemaV10),
					string(eventgrid.InputSchemaCustomEventSchema),
					string(eventgrid.InputSchemaEventGridSchema),
				}, false),
			},

			//lintignore:XS003
			"input_mapping_fields": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"topic": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"event_time": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"event_type": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"subject": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"data_version": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},

			//lintignore:XS003
			"input_mapping_default_values": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"event_type": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"subject": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"data_version": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},

			"public_network_access_enabled": eventSubscriptionPublicNetworkAccessEnabled(),

			"local_auth_enabled": localAuthEnabled(),

			"auto_create_topic_with_first_subscription": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"auto_delete_topic_with_last_subscription": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"inbound_ip_rule": eventSubscriptionInboundIPRule(),

			"endpoint": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"primary_access_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_access_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceEventGridDomainCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).EventGrid.DomainsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewDomainID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_eventgrid_domain", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	domainProperties := &eventgrid.DomainProperties{
		InputSchemaMapping:                   expandAzureRmEventgridDomainInputMapping(d),
		InputSchema:                          eventgrid.InputSchema(d.Get("input_schema").(string)),
		PublicNetworkAccess:                  expandPublicNetworkAccess(d),
		InboundIPRules:                       expandInboundIPRules(d),
		DisableLocalAuth:                     utils.Bool(!d.Get("local_auth_enabled").(bool)),
		AutoCreateTopicWithFirstSubscription: utils.Bool(d.Get("auto_create_topic_with_first_subscription").(bool)),
		AutoDeleteTopicWithLastSubscription:  utils.Bool(d.Get("auto_delete_topic_with_last_subscription").(bool)),
	}

	domain := eventgrid.Domain{
		Location:         &location,
		DomainProperties: domainProperties,
		Tags:             tags.Expand(t),
	}

	if v, ok := d.GetOk("identity"); ok {
		identityRaw := v.([]interface{})
		identity, err := expandIdentity(identityRaw)
		if err != nil {
			return fmt.Errorf("expanding `identity`: %+v", err)
		}
		domain.Identity = identity
	}

	log.Printf("[INFO] preparing arguments for AzureRM EventGrid Domain creation with Properties: %+v", domain)

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, domain)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %s", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for %s to become available: %s", id, err)
	}

	d.SetId(id.ID())

	return resourceEventGridDomainRead(d, meta)
}

func resourceEventGridDomainRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).EventGrid.DomainsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DomainID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] EventGrid Domain %q was not found (Resource Group %q)", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request on EventGrid Domain %q: %+v", id.Name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.DomainProperties; props != nil {
		d.Set("endpoint", props.Endpoint)

		d.Set("input_schema", string(props.InputSchema))

		inputMappingFields, err := flattenAzureRmEventgridDomainInputMapping(props.InputSchemaMapping)
		if err != nil {
			return fmt.Errorf("flattening `input_schema_mapping_fields` for EventGrid Domain %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
		}
		if err := d.Set("input_mapping_fields", inputMappingFields); err != nil {
			return fmt.Errorf("setting `input_schema_mapping_fields` for EventGrid Domain %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
		}

		inputMappingDefaultValues, err := flattenAzureRmEventgridDomainInputMappingDefaultValues(props.InputSchemaMapping)
		if err != nil {
			return fmt.Errorf("flattening `input_schema_mapping_default_values` for EventGrid Domain %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
		}
		if err := d.Set("input_mapping_default_values", inputMappingDefaultValues); err != nil {
			return fmt.Errorf("setting `input_schema_mapping_fields` for EventGrid Domain %q (Resource Group %q): %s", id.Name, id.ResourceGroup, err)
		}

		publicNetworkAccessEnabled := flattenPublicNetworkAccess(props.PublicNetworkAccess)
		if err := d.Set("public_network_access_enabled", publicNetworkAccessEnabled); err != nil {
			return fmt.Errorf("setting `public_network_access_enabled` in EventGrid Domain %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}

		inboundIPRules := flattenInboundIPRules(props.InboundIPRules)
		if err := d.Set("inbound_ip_rule", inboundIPRules); err != nil {
			return fmt.Errorf("setting `inbound_ip_rule` in EventGrid Domain %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}

		localAuthEnabled := true
		if props.DisableLocalAuth != nil {
			localAuthEnabled = !*props.DisableLocalAuth
		}

		d.Set("local_auth_enabled", localAuthEnabled)

		autoCreateTopicWithFirstSubscription := true
		if props.AutoCreateTopicWithFirstSubscription != nil {
			autoCreateTopicWithFirstSubscription = *props.AutoCreateTopicWithFirstSubscription
		}

		d.Set("auto_create_topic_with_first_subscription", autoCreateTopicWithFirstSubscription)

		autoDeleteTopicWithLastSubscription := true
		if props.AutoDeleteTopicWithLastSubscription != nil {
			autoDeleteTopicWithLastSubscription = *props.AutoDeleteTopicWithLastSubscription
		}

		d.Set("auto_delete_topic_with_last_subscription", autoDeleteTopicWithLastSubscription)
	}

	keys, err := client.ListSharedAccessKeys(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Shared Access Keys for EventGrid Domain %q: %+v", id.Name, err)
	}

	flattenedIdentity, err := flattenIdentity(resp.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", flattenedIdentity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	d.Set("primary_access_key", keys.Key1)
	d.Set("secondary_access_key", keys.Key2)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceEventGridDomainDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).EventGrid.DomainsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DomainID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Event Grid Domain %q: %+v", id.Name, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("deleting Event Grid Domain %q: %+v", id.Name, err)
	}

	return nil
}

func expandAzureRmEventgridDomainInputMapping(d *pluginsdk.ResourceData) *eventgrid.JSONInputSchemaMapping {
	imf, imfok := d.GetOk("input_mapping_fields")

	imdv, imdvok := d.GetOk("input_mapping_default_values")

	if !imfok && !imdvok {
		return nil
	}

	jismp := eventgrid.JSONInputSchemaMappingProperties{}

	if imfok {
		mappings := imf.([]interface{})
		if len(mappings) > 0 && mappings[0] != nil {
			mapping := mappings[0].(map[string]interface{})

			if id := mapping["id"].(string); id != "" {
				jismp.ID = &eventgrid.JSONField{SourceField: &id}
			}

			if eventTime := mapping["event_time"].(string); eventTime != "" {
				jismp.EventTime = &eventgrid.JSONField{SourceField: &eventTime}
			}

			if topic := mapping["topic"].(string); topic != "" {
				jismp.Topic = &eventgrid.JSONField{SourceField: &topic}
			}

			if dataVersion := mapping["data_version"].(string); dataVersion != "" {
				jismp.DataVersion = &eventgrid.JSONFieldWithDefault{SourceField: &dataVersion}
			}

			if subject := mapping["subject"].(string); subject != "" {
				jismp.Subject = &eventgrid.JSONFieldWithDefault{SourceField: &subject}
			}

			if eventType := mapping["event_type"].(string); eventType != "" {
				jismp.EventType = &eventgrid.JSONFieldWithDefault{SourceField: &eventType}
			}
		}
	}

	if imdvok {
		mappings := imdv.([]interface{})
		if len(mappings) > 0 && mappings[0] != nil {
			mapping := mappings[0].(map[string]interface{})

			if dataVersion := mapping["data_version"].(string); dataVersion != "" {
				jismp.DataVersion = &eventgrid.JSONFieldWithDefault{DefaultValue: &dataVersion}
			}

			if subject := mapping["subject"].(string); subject != "" {
				jismp.Subject = &eventgrid.JSONFieldWithDefault{DefaultValue: &subject}
			}

			if eventType := mapping["event_type"].(string); eventType != "" {
				jismp.EventType = &eventgrid.JSONFieldWithDefault{DefaultValue: &eventType}
			}
		}
	}

	jsonMapping := eventgrid.JSONInputSchemaMapping{
		JSONInputSchemaMappingProperties: &jismp,
		InputSchemaMappingType:           eventgrid.InputSchemaMappingTypeJSON,
	}

	return &jsonMapping
}

func flattenAzureRmEventgridDomainInputMapping(input eventgrid.BasicInputSchemaMapping) ([]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	result := make(map[string]interface{})

	jsonValues, ok := input.(eventgrid.JSONInputSchemaMapping)
	if !ok {
		return nil, fmt.Errorf("Unable to read JSONInputSchemaMapping")
	}
	props := jsonValues.JSONInputSchemaMappingProperties

	if props.EventTime != nil && props.EventTime.SourceField != nil {
		result["event_time"] = *props.EventTime.SourceField
	}

	if props.ID != nil && props.ID.SourceField != nil {
		result["id"] = *props.ID.SourceField
	}

	if props.Topic != nil && props.Topic.SourceField != nil {
		result["topic"] = *props.Topic.SourceField
	}

	if props.DataVersion != nil && props.DataVersion.SourceField != nil {
		result["data_version"] = *props.DataVersion.SourceField
	}

	if props.EventType != nil && props.EventType.SourceField != nil {
		result["event_type"] = *props.EventType.SourceField
	}

	if props.Subject != nil && props.Subject.SourceField != nil {
		result["subject"] = *props.Subject.SourceField
	}

	return []interface{}{result}, nil
}

func flattenAzureRmEventgridDomainInputMappingDefaultValues(input eventgrid.BasicInputSchemaMapping) ([]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	result := make(map[string]interface{})

	jsonValues, ok := input.(eventgrid.JSONInputSchemaMapping)
	if !ok {
		return nil, fmt.Errorf("Unable to read JSONInputSchemaMapping")
	}
	props := jsonValues.JSONInputSchemaMappingProperties

	if props.DataVersion != nil && props.DataVersion.DefaultValue != nil {
		result["data_version"] = *props.DataVersion.DefaultValue
	}

	if props.EventType != nil && props.EventType.DefaultValue != nil {
		result["event_type"] = *props.EventType.DefaultValue
	}

	if props.Subject != nil && props.Subject.DefaultValue != nil {
		result["subject"] = *props.Subject.DefaultValue
	}

	return []interface{}{result}, nil
}
