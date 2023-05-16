package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-11-01/compute"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func dataSourceImages() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceImagesRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"tags_filter": tags.Schema(),

			"images": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"location": commonschema.LocationComputed(),

						"zone_resilient": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},

						"os_disk": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"blob_uri": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"caching": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"managed_disk_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"os_state": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"os_type": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"size_gb": {
										Type:     pluginsdk.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"data_disk": {
							Type:     pluginsdk.TypeList,
							Computed: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"blob_uri": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"caching": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"lun": {
										Type:     pluginsdk.TypeInt,
										Computed: true,
									},
									"managed_disk_id": {
										Type:     pluginsdk.TypeString,
										Computed: true,
									},
									"size_gb": {
										Type:     pluginsdk.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"tags": tags.SchemaDataSource(),
					},
				},
			},
		},
	}
}

func dataSourceImagesRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImagesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	filterTags := tags.Expand(d.Get("tags_filter").(map[string]interface{}))

	resp, err := client.ListByResourceGroupComplete(ctx, resourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response().Response) {
			return fmt.Errorf("no images were found in Resource Group %q", resourceGroup)
		}
		return fmt.Errorf("retrieving Images (Resource Group %q): %+v", resourceGroup, err)
	}

	images, err := flattenImagesResult(ctx, resp, filterTags)
	if err != nil {
		return fmt.Errorf("parsing Images (Resource Group %q): %+v", resourceGroup, err)
	}
	if len(images) == 0 {
		return fmt.Errorf("no images were found that match the specified tags")
	}

	d.SetId(time.Now().UTC().String())

	d.Set("resource_group_name", resourceGroup)

	if err := d.Set("images", images); err != nil {
		return fmt.Errorf("setting `images`: %+v", err)
	}

	return nil
}

func flattenImagesResult(ctx context.Context, iterator compute.ImageListResultIterator, filterTags map[string]*string) ([]interface{}, error) {
	results := make([]interface{}, 0)

	for iterator.NotDone() {
		image := iterator.Value()
		found := true
		// Loop through our filter tags and see if they match
		for k, v := range filterTags {
			if v != nil {
				// If the tags do not match return false
				if image.Tags[k] == nil || *v != *image.Tags[k] {
					found = false
				}
			}
		}

		if found {
			results = append(results, flattenImage(image))
		}
		if err := iterator.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func flattenImage(input compute.Image) map[string]interface{} {
	output := make(map[string]interface{})

	output["name"] = input.Name
	output["location"] = location.NormalizeNilable(input.Location)

	if input.ImageProperties != nil {
		if storageProfile := input.ImageProperties.StorageProfile; storageProfile != nil {
			output["zone_resilient"] = storageProfile.ZoneResilient

			output["os_disk"] = flattenAzureRmImageOSDisk(storageProfile.OsDisk)

			output["data_disk"] = flattenAzureRmImageDataDisks(storageProfile.DataDisks)
		}
	}

	output["tags"] = tags.Flatten(input.Tags)

	return output
}
