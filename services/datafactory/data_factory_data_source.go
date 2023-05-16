package datafactory

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datafactory/mgmt/2018-06-01/datafactory"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/datafactory/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceDataFactory() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceDataFactoryRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`),
					`Invalid name for Data Factory, see https://docs.microsoft.com/en-us/azure/data-factory/naming-rules`,
				),
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"identity": commonschema.SystemAssignedUserAssignedIdentityComputed(),

			"github_configuration": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"account_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"branch_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"git_url": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"repository_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"root_folder": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"vsts_configuration": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"account_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"branch_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"project_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"repository_name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"root_folder": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"tenant_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceDataFactoryRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataFactory.FactoriesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewDataFactoryID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.FactoryName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("location", location.NormalizeNilable(resp.Location))

	d.Set("vsts_configuration", []interface{}{})
	d.Set("github_configuration", []interface{}{})
	repoType, repo := flattenDataFactoryRepoConfiguration(&resp)
	if repoType == datafactory.TypeBasicFactoryRepoConfigurationTypeFactoryVSTSConfiguration {
		if err := d.Set("vsts_configuration", repo); err != nil {
			return fmt.Errorf("setting `vsts_configuration`: %+v", err)
		}
	}
	if repoType == datafactory.TypeBasicFactoryRepoConfigurationTypeFactoryGitHubConfiguration {
		if err := d.Set("github_configuration", repo); err != nil {
			return fmt.Errorf("setting `github_configuration`: %+v", err)
		}
	}
	if repoType == datafactory.TypeBasicFactoryRepoConfigurationTypeFactoryRepoConfiguration {
		d.Set("vsts_configuration", repo)
		d.Set("github_configuration", repo)
	}

	identity, err := flattenIdentity(resp.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", identity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
