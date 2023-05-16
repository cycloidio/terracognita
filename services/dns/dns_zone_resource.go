package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/dns/migration"
	"github.com/hashicorp/terraform-provider-azurerm/services/dns/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/dns/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceDnsZone() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDnsZoneCreateUpdate,
		Read:   resourceDnsZoneRead,
		Update: resourceDnsZoneCreateUpdate,
		Delete: resourceDnsZoneDelete,

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.DnsZoneV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.DnsZoneID(id)
			return err
		}),
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"number_of_record_sets": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"max_number_of_record_sets": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"name_servers": {
				Type:     pluginsdk.TypeSet,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"soa_record": {
				Type:     pluginsdk.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"email": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validate.DnsZoneSOARecordEmail,
						},

						"host_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"expire_time": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      2419200,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"minimum_ttl": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"refresh_time": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      3600,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"retry_time": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"serial_number": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"ttl": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      3600,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},

						"tags": tags.Schema(),

						"fqdn": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceDnsZoneCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Dns.ZonesClient
	recordSetsClient := meta.(*clients.Client).Dns.RecordSetsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)

	resourceId := parse.NewDnsZoneID(subscriptionId, resGroup, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing DNS Zone %q (Resource Group %q): %s", name, resGroup, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_dns_zone", resourceId.ID())
		}
	}

	location := "global"
	t := d.Get("tags").(map[string]interface{})

	parameters := dns.Zone{
		Location: &location,
		Tags:     tags.Expand(t),
	}

	etag := ""
	ifNoneMatch := "" // set to empty to allow updates to records after creation
	if _, err := client.CreateOrUpdate(ctx, resGroup, name, parameters, etag, ifNoneMatch); err != nil {
		return fmt.Errorf("creating/updating DNS Zone %q (Resource Group %q): %s", name, resGroup, err)
	}

	if v, ok := d.GetOk("soa_record"); ok {
		soaRecord := v.([]interface{})[0].(map[string]interface{})
		rsParameters := dns.RecordSet{
			RecordSetProperties: &dns.RecordSetProperties{
				TTL:       utils.Int64(int64(soaRecord["ttl"].(int))),
				Metadata:  tags.Expand(soaRecord["tags"].(map[string]interface{})),
				SoaRecord: expandArmDNSZoneSOARecord(soaRecord),
			},
		}

		if len(name+strings.TrimSuffix(*rsParameters.RecordSetProperties.SoaRecord.Email, ".")) > 253 {
			return fmt.Errorf("`email` which is concatenated with DNS Zone `name` cannot exceed 253 characters excluding a trailing period")
		}

		if _, err := recordSetsClient.CreateOrUpdate(ctx, resGroup, name, "@", dns.SOA, rsParameters, etag, ifNoneMatch); err != nil {
			return fmt.Errorf("creating/updating DNS SOA Record @ (Zone %q / Resource Group %q): %s", name, resGroup, err)
		}
	}

	d.SetId(resourceId.ID())

	return resourceDnsZoneRead(d, meta)
}

func resourceDnsZoneRead(d *pluginsdk.ResourceData, meta interface{}) error {
	zonesClient := meta.(*clients.Client).Dns.ZonesClient
	recordSetsClient := meta.(*clients.Client).Dns.RecordSetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DnsZoneID(d.Id())
	if err != nil {
		return err
	}

	resp, err := zonesClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading DNS Zone %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	d.Set("number_of_record_sets", resp.NumberOfRecordSets)
	d.Set("max_number_of_record_sets", resp.MaxNumberOfRecordSets)

	nameServers := make([]string, 0)
	if s := resp.NameServers; s != nil {
		nameServers = *s
	}
	if err := d.Set("name_servers", nameServers); err != nil {
		return err
	}

	rsResp, err := recordSetsClient.Get(ctx, id.ResourceGroup, id.Name, "@", dns.SOA)
	if err != nil {
		return fmt.Errorf("reading DNS SOA record @: %v", err)
	}

	if err := d.Set("soa_record", flattenArmDNSZoneSOARecord(&rsResp)); err != nil {
		return fmt.Errorf("setting `soa_record`: %+v", err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceDnsZoneDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Dns.ZonesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DnsZoneID(d.Id())
	if err != nil {
		return err
	}

	etag := ""
	future, err := client.Delete(ctx, id.ResourceGroup, id.Name, etag)
	if err != nil {
		return fmt.Errorf("deleting DNS Zone %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of DNS Zone %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func expandArmDNSZoneSOARecord(input map[string]interface{}) *dns.SoaRecord {
	return &dns.SoaRecord{
		Email:        utils.String(input["email"].(string)),
		Host:         utils.String(input["host_name"].(string)),
		ExpireTime:   utils.Int64(int64(input["expire_time"].(int))),
		MinimumTTL:   utils.Int64(int64(input["minimum_ttl"].(int))),
		RefreshTime:  utils.Int64(int64(input["refresh_time"].(int))),
		RetryTime:    utils.Int64(int64(input["retry_time"].(int))),
		SerialNumber: utils.Int64(int64(input["serial_number"].(int))),
	}
}

func flattenArmDNSZoneSOARecord(input *dns.RecordSet) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	ttl := 0
	if input.TTL != nil {
		ttl = int(*input.TTL)
	}

	metaData := make(map[string]interface{})
	if input.Metadata != nil {
		metaData = tags.Flatten(input.Metadata)
	}

	fqdn := ""
	if input.Fqdn != nil {
		fqdn = *input.Fqdn
	}

	email := ""
	hostName := ""
	expireTime := 0
	minimumTTL := 0
	refreshTime := 0
	retryTime := 0
	serialNumber := 0
	if input.SoaRecord != nil {
		if input.SoaRecord.Email != nil {
			email = *input.SoaRecord.Email
		}

		if input.SoaRecord.Host != nil {
			hostName = *input.SoaRecord.Host
		}

		if input.SoaRecord.ExpireTime != nil {
			expireTime = int(*input.SoaRecord.ExpireTime)
		}

		if input.SoaRecord.MinimumTTL != nil {
			minimumTTL = int(*input.SoaRecord.MinimumTTL)
		}

		if input.SoaRecord.RefreshTime != nil {
			refreshTime = int(*input.SoaRecord.RefreshTime)
		}

		if input.SoaRecord.RetryTime != nil {
			retryTime = int(*input.SoaRecord.RetryTime)
		}

		if input.SoaRecord.SerialNumber != nil {
			serialNumber = int(*input.SoaRecord.SerialNumber)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"email":         email,
			"host_name":     hostName,
			"expire_time":   expireTime,
			"minimum_ttl":   minimumTTL,
			"refresh_time":  refreshTime,
			"retry_time":    retryTime,
			"serial_number": serialNumber,
			"ttl":           ttl,
			"tags":          metaData,
			"fqdn":          fqdn,
		},
	}
}
