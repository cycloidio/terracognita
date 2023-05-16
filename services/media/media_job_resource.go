package media

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/mediaservices/mgmt/2021-05-01/media"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/media/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceMediaJob() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMediaJobCreate,
		Read:   resourceMediaJobRead,
		Update: resourceMediaJobUpdate,
		Delete: resourceMediaJobDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.JobID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-zA-Z0-9(_)]{1,128}$"),
					"Job name must be 1 - 128 characters long, can contain letters, numbers, underscores, and hyphens (but the first and last character must be a letter or number).",
				),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"media_services_account_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-z0-9]{3,24}$"),
					"Media Services Account name must be 3 - 24 characters long, contain only lowercase letters and numbers.",
				),
			},

			"transform_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-zA-Z0-9(_)]{1,128}$"),
					"Transform name must be 1 - 128 characters long, can contain letters, numbers, underscores, and hyphens (but the first and last character must be a letter or number).",
				),
			},

			"input_asset": {
				Type:     pluginsdk.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringMatch(
								regexp.MustCompile("^[-a-zA-Z0-9]{1,128}$"),
								"Asset name must be 1 - 128 characters long, contain only letters, hyphen and numbers.",
							),
						},
						"label": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"output_asset": {
				Type:     pluginsdk.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringMatch(
								regexp.MustCompile("^[-a-zA-Z0-9]{1,128}$"),
								"Asset name must be 1 - 128 characters long, contain only letters, hyphen and numbers.",
							),
						},
						"label": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"priority": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  string(media.PriorityNormal),
				ValidateFunc: validation.StringInSlice([]string{
					string(media.PriorityHigh),
					string(media.PriorityNormal),
					string(media.PriorityLow),
				}, false),
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceMediaJobCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.JobsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceId := parse.NewJobID(subscriptionId, d.Get("resource_group_name").(string), d.Get("media_services_account_name").(string), d.Get("transform_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.MediaserviceName, resourceId.TransformName, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Media Job %q (Media Service account %q) (ResourceGroup %q): %s", resourceId.ResourceGroup, resourceId.MediaserviceName, resourceId.Name, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_media_job", resourceId.ID())
		}
	}

	parameters := media.Job{
		JobProperties: &media.JobProperties{
			Description: utils.String(d.Get("description").(string)),
		},
	}

	if v, ok := d.GetOk("priority"); ok {
		parameters.Priority = media.Priority(v.(string))
	}

	if v, ok := d.GetOk("input_asset"); ok {
		parameters.JobProperties.Input = expandInputAsset(v.([]interface{}))
	}

	if v, ok := d.GetOk("output_asset"); ok {
		outputAssets, err := expandOutputAssets(v.([]interface{}))
		if err != nil {
			return err
		}
		parameters.JobProperties.Outputs = outputAssets
	}

	if _, err := client.Create(ctx, resourceId.ResourceGroup, resourceId.MediaserviceName, resourceId.TransformName, resourceId.Name, parameters); err != nil {
		return fmt.Errorf("creating %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())

	return resourceMediaJobRead(d, meta)
}

func resourceMediaJobRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.JobsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.JobID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.MediaserviceName, id.TransformName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("media_services_account_name", id.MediaserviceName)
	d.Set("transform_name", id.TransformName)

	if props := resp.JobProperties; props != nil {
		d.Set("description", props.Description)
		d.Set("priority", string(props.Priority))

		inputAsset, err := flattenInputAsset(props.Input)
		if err != nil {
			return err
		}
		if err = d.Set("input_asset", inputAsset); err != nil {
			return fmt.Errorf("flattening `input_asset`: %s", err)
		}

		outputAssets, err := flattenOutputAssets(props.Outputs)
		if err != nil {
			return err
		}
		if err = d.Set("output_asset", outputAssets); err != nil {
			return fmt.Errorf("flattening `output_asset`: %s", err)
		}
	}
	return nil
}

func resourceMediaJobUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.JobsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.JobID(d.Id())
	if err != nil {
		return err
	}
	description := d.Get("description").(string)

	parameters := media.Job{
		JobProperties: &media.JobProperties{
			Description: utils.String(description),
		},
	}

	if v, ok := d.GetOk("priority"); ok {
		parameters.Priority = media.Priority(v.(string))
	}

	if v, ok := d.GetOk("input_asset"); ok {
		inputAsset := expandInputAsset(v.([]interface{}))
		parameters.JobProperties.Input = inputAsset
	}

	if v, ok := d.GetOk("output_asset"); ok {
		outputAssets, err := expandOutputAssets(v.([]interface{}))
		if err != nil {
			return err
		}
		parameters.JobProperties.Outputs = outputAssets
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.MediaserviceName, id.TransformName, id.Name, parameters); err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	return resourceMediaJobRead(d, meta)
}

func resourceMediaJobDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.JobsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.JobID(d.Id())
	if err != nil {
		return err
	}

	// Cancel the job before we attempt to delete it.
	if _, err = client.CancelJob(ctx, id.ResourceGroup, id.MediaserviceName, id.TransformName, id.Name); err != nil {
		return fmt.Errorf("could not cancel Media Job %q (reource group %q) for delete: %+v", id.Name, id.ResourceGroup, err)
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.MediaserviceName, id.TransformName, id.Name)
	if err != nil {
		if response.WasNotFound(resp.Response) {
			return nil
		}
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}

func expandInputAsset(input []interface{}) media.BasicJobInput {
	inputAsset := input[0].(map[string]interface{})
	assetName := inputAsset["name"].(string)
	label := inputAsset["label"].(string)
	return &media.JobInputAsset{
		AssetName: utils.String(assetName),
		Label:     utils.String(label),
	}
}

func flattenInputAsset(input media.BasicJobInput) ([]interface{}, error) {
	if input == nil {
		return make([]interface{}, 0), nil
	}

	asset, ok := input.AsJobInputAsset()
	if !ok {
		return nil, fmt.Errorf("Unexpected type for Input Asset. Currently only JobInputAsset is supported.")
	}
	assetName := ""
	if asset.AssetName != nil {
		assetName = *asset.AssetName
	}

	label := ""
	if asset.Label != nil {
		label = *asset.Label
	}

	return []interface{}{
		map[string]interface{}{
			"name":  assetName,
			"label": label,
		},
	}, nil
}

func expandOutputAssets(input []interface{}) (*[]media.BasicJobOutput, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("Job must contain at least one output_asset.")
	}
	outputAssets := make([]media.BasicJobOutput, len(input))
	for index, output := range input {
		outputAsset := output.(map[string]interface{})
		assetName := outputAsset["name"].(string)
		label := outputAsset["label"].(string)
		jobOutputAsset := media.JobOutputAsset{
			AssetName: utils.String(assetName),
			Label:     utils.String(label),
		}
		outputAssets[index] = jobOutputAsset
	}

	return &outputAssets, nil
}

func flattenOutputAssets(input *[]media.BasicJobOutput) ([]interface{}, error) {
	if input == nil || len(*input) == 0 {
		return []interface{}{}, nil
	}

	outputAssets := make([]interface{}, len(*input))
	for i, output := range *input {
		outputAssetJob, ok := output.AsJobOutputAsset()
		if !ok {
			return nil, fmt.Errorf("unexpected type for output_asset. Currently only JobOutputAsset is supported.")
		}
		assetName := ""
		if outputAssetJob.AssetName != nil {
			assetName = *outputAssetJob.AssetName
		}

		label := ""
		if outputAssetJob.Label != nil {
			label = *outputAssetJob.Label
		}

		outputAssets[i] = map[string]interface{}{
			"name":  assetName,
			"label": label,
		}
	}
	return outputAssets, nil
}
