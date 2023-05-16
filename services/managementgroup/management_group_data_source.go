package managementgroup

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-05-01/managementgroups"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/managementgroup/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceManagementGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceManagementGroupRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "display_name"},
				ValidateFunc: validate.ManagementGroupName,
			},

			"display_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "display_name"},
			},

			"parent_management_group_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"subscription_ids": {
				Type:     pluginsdk.TypeSet,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},
		},
	}
}

func dataSourceManagementGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagementGroups.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	groupName := ""
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	}
	displayName := d.Get("display_name").(string)

	// one of displayName and groupName must be non-empty, this is guaranteed by schema
	// if the user is retrieving the mgmt group by display name, use the list api to get the group name first
	var err error
	if displayName != "" {
		groupName, err = getManagementGroupNameByDisplayName(ctx, client, displayName)
		if err != nil {
			return fmt.Errorf("reading Management Group (Display Name %q): %+v", displayName, err)
		}
	}
	recurse := true
	resp, err := client.Get(ctx, groupName, "children", &recurse, "", managementGroupCacheControl)
	if err != nil {
		if utils.ResponseWasForbidden(resp.Response) {
			return fmt.Errorf("Management Group %q was not found", groupName)
		}

		return fmt.Errorf("reading Management Group %q: %+v", groupName, err)
	}

	id := parse.NewManagementGroupId(groupName)
	d.SetId(id.ID())
	d.Set("name", groupName)

	if props := resp.Properties; props != nil {
		d.Set("display_name", props.DisplayName)

		subscriptionIds, err := flattenManagementGroupDataSourceSubscriptionIds(props.Children)
		if err != nil {
			return fmt.Errorf("flattening `subscription_ids`: %+v", err)
		}
		d.Set("subscription_ids", subscriptionIds)

		parentId := ""
		if details := props.Details; details != nil {
			if parent := details.Parent; parent != nil {
				if pid := parent.ID; pid != nil {
					parentId = *pid
				}
			}
		}
		d.Set("parent_management_group_id", parentId)
	}

	return nil
}

func getManagementGroupNameByDisplayName(ctx context.Context, client *managementgroups.Client, displayName string) (string, error) {
	iterator, err := client.ListComplete(ctx, managementGroupCacheControl, "")
	if err != nil {
		return "", fmt.Errorf("listing Management Groups: %+v", err)
	}

	var results []string
	for iterator.NotDone() {
		group := iterator.Value()
		if group.DisplayName != nil && *group.DisplayName == displayName && group.Name != nil && *group.Name != "" {
			results = append(results, *group.Name)
		}

		if err := iterator.NextWithContext(ctx); err != nil {
			return "", fmt.Errorf("listing Management Groups: %+v", err)
		}
	}

	// we found none
	if len(results) == 0 {
		return "", fmt.Errorf("Management Group (Display Name %q) was not found", displayName)
	}

	// we found more than one
	if len(results) > 1 {
		return "", fmt.Errorf("expected a single Management Group with the Display Name %q but expected one", displayName)
	}

	return results[0], nil
}

func flattenManagementGroupDataSourceSubscriptionIds(input *[]managementgroups.ChildInfo) (*pluginsdk.Set, error) {
	subscriptionIds := &pluginsdk.Set{F: pluginsdk.HashString}
	if input == nil {
		return subscriptionIds, nil
	}

	for _, child := range *input {
		if child.ID == nil {
			continue
		}

		id, err := parseManagementGroupSubscriptionID(*child.ID)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse child Subscription ID %+v", err)
		}

		if id != nil {
			subscriptionIds.Add(id.subscriptionId)
		}
	}

	return subscriptionIds, nil
}
