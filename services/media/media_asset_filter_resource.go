package media

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/mediaservices/mgmt/2021-05-01/media"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/media/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

// define constants based on docs https://docs.microsoft.com/en-us/azure/media-services/latest/filters-concept
const incrementsInASecond = 10000000

const (
	nanoSecondsInAIncrement = 100
	milliSecondsInASecond   = 1000
)

func resourceMediaAssetFilter() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceMediaAssetFilterCreateUpdate,
		Read:   resourceMediaAssetFilterRead,
		Update: resourceMediaAssetFilterCreateUpdate,
		Delete: resourceMediaAssetFilterDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.AssetFilterID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-zA-Z0-9(_)]{1,128}$"),
					"Asset Filter name must be 1 - 128 characters long, can contain letters, numbers, underscores, and hyphens (but the first and last character must be a letter or number).",
				),
			},

			"asset_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"first_quality_bitrate": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			"presentation_time_range": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"end_in_units": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},

						"force_end": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},

						"live_backoff_in_units": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},

						"presentation_window_in_units": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},

						"start_in_units": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},

						"unit_timescale_in_miliseconds": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							AtLeastOneOf: []string{
								"presentation_time_range.0.end_in_units", "presentation_time_range.0.force_end", "presentation_time_range.0.live_backoff_in_units",
								"presentation_time_range.0.presentation_window_in_units", "presentation_time_range.0.start_in_units", "presentation_time_range.0.unit_timescale_in_miliseconds",
							},
						},
					},
				},
			},

			"track_selection": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						//lintignore:XS003
						"condition": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"operation": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.FilterTrackPropertyCompareOperationEqual),
											string(media.FilterTrackPropertyCompareOperationNotEqual),
										}, false),
									},

									"property": {
										Type:     pluginsdk.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(media.FilterTrackPropertyTypeBitrate),
											string(media.FilterTrackPropertyTypeFourCC),
											string(media.FilterTrackPropertyTypeLanguage),
											string(media.FilterTrackPropertyTypeName),
											string(media.FilterTrackPropertyTypeType),
										}, false),
									},

									"value": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceMediaAssetFilterCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.AssetFiltersClient
	subscriptionID := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	assetID, err := parse.AssetID(d.Get("asset_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewAssetFilterID(subscriptionID, assetID.ResourceGroup, assetID.MediaserviceName, assetID.Name, d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.MediaserviceName, id.AssetName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of %s: %+v", id, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_media_asset_filter", id.ID())
		}
	}

	parameters := media.AssetFilter{
		FilterProperties: &media.FilterProperties{
			FirstQuality: &media.FirstQuality{},
		},
	}

	if firstQualityBitrate, ok := d.GetOk("first_quality_bitrate"); ok {
		parameters.FilterProperties.FirstQuality.Bitrate = utils.Int32(int32(firstQualityBitrate.(int)))
	}

	if v, ok := d.GetOk("presentation_time_range"); ok {
		parameters.FilterProperties.PresentationTimeRange = expandPresentationTimeRange(v.([]interface{}))
	}

	if v, ok := d.GetOk("track_selection"); ok {
		parameters.FilterProperties.Tracks = expandTracks(v.([]interface{}))
	}

	if _, err = client.CreateOrUpdate(ctx, id.ResourceGroup, id.MediaserviceName, id.AssetName, id.Name, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceMediaAssetFilterRead(d, meta)
}

func resourceMediaAssetFilterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.AssetFiltersClient
	subscriptionID := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AssetFilterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.MediaserviceName, id.AssetName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	assetID := parse.NewAssetID(subscriptionID, id.ResourceGroup, id.MediaserviceName, id.AssetName)
	d.Set("asset_id", assetID.ID())

	if props := resp.FilterProperties; props != nil {
		var firstQualityBitrate int32
		if props.FirstQuality != nil && props.FirstQuality.Bitrate != nil {
			firstQualityBitrate = *props.FirstQuality.Bitrate
		}
		d.Set("first_quality_bitrate", firstQualityBitrate)

		if err := d.Set("presentation_time_range", flattenPresentationTimeRange(resp.PresentationTimeRange)); err != nil {
			return fmt.Errorf("flattening `presentation_time_range`: %s", err)
		}

		if err := d.Set("track_selection", flattenTracks(resp.Tracks)); err != nil {
			return fmt.Errorf("flattening `track`: %s", err)
		}
	}

	return nil
}

func resourceMediaAssetFilterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Media.AssetFiltersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AssetFilterID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.Delete(ctx, id.ResourceGroup, id.MediaserviceName, id.AssetName, id.Name); err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	return nil
}

func expandPresentationTimeRange(input []interface{}) *media.PresentationTimeRange {
	if len(input) == 0 {
		return nil
	}

	timeRange := input[0].(map[string]interface{})
	presentationTimeRange := &media.PresentationTimeRange{}

	var baseUnit int64
	if v := timeRange["unit_timescale_in_miliseconds"]; v != nil {
		timeScaleInMiliSeconds := int64(v.(int))
		presentationTimeRange.Timescale = utils.Int64((incrementsInASecond * nanoSecondsInAIncrement) / milliSecondsInASecond / timeScaleInMiliSeconds)
		baseUnit = milliSecondsInASecond
	}

	if v := timeRange["end_in_units"]; v != nil {
		presentationTimeRange.EndTimestamp = utils.Int64(int64(v.(int)) * baseUnit)
	}

	if v := timeRange["force_end"]; v != nil {
		presentationTimeRange.ForceEndTimestamp = utils.Bool(v.(bool))
	}

	if v := timeRange["live_backoff_in_units"]; v != nil {
		presentationTimeRange.LiveBackoffDuration = utils.Int64(int64(v.(int)) * baseUnit)
	}

	if v := timeRange["presentation_window_in_units"]; v != nil {
		presentationTimeRange.PresentationWindowDuration = utils.Int64(int64(v.(int)) * baseUnit)
	}

	if v := timeRange["start_in_units"]; v != nil {
		presentationTimeRange.StartTimestamp = utils.Int64(int64(v.(int)) * baseUnit)
	}

	return presentationTimeRange
}

func flattenPresentationTimeRange(input *media.PresentationTimeRange) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var timeScale int64
	var baseUnit int64
	if input.Timescale != nil {
		timeScale = (incrementsInASecond * nanoSecondsInAIncrement) / milliSecondsInASecond / *input.Timescale
		baseUnit = milliSecondsInASecond
	}

	var endTimestamp int64
	if input.EndTimestamp != nil {
		endTimestamp = *input.EndTimestamp / baseUnit
	}

	var forceEndTimestamp bool
	if input.ForceEndTimestamp != nil {
		forceEndTimestamp = *input.ForceEndTimestamp
	}

	var liveBackoffDuration int64
	if input.LiveBackoffDuration != nil {
		liveBackoffDuration = *input.LiveBackoffDuration / baseUnit
	}

	var presentationWindowDuration int64
	if input.PresentationWindowDuration != nil {
		presentationWindowDuration = *input.PresentationWindowDuration / baseUnit
	}

	var startTimestamp int64
	if input.StartTimestamp != nil {
		startTimestamp = *input.StartTimestamp / baseUnit
	}

	return []interface{}{
		map[string]interface{}{
			"end_in_units":                  endTimestamp,
			"force_end":                     forceEndTimestamp,
			"live_backoff_in_units":         liveBackoffDuration,
			"presentation_window_in_units":  presentationWindowDuration,
			"start_in_units":                startTimestamp,
			"unit_timescale_in_miliseconds": timeScale,
		},
	}
}

func expandTracks(input []interface{}) *[]media.FilterTrackSelection {
	results := make([]media.FilterTrackSelection, 0)

	for _, trackRaw := range input {
		track := trackRaw.(map[string]interface{})

		if rawSelection := track["condition"]; rawSelection != nil {
			trackSelectionList := rawSelection.([]interface{})
			filterTrackSelections := make([]media.FilterTrackPropertyCondition, 0)
			for _, trackSelection := range trackSelectionList {
				if trackSelection == nil {
					continue
				}
				filterTrackSelection := media.FilterTrackPropertyCondition{}
				track := trackSelection.(map[string]interface{})

				if v := track["operation"]; v != nil {
					filterTrackSelection.Operation = media.FilterTrackPropertyCompareOperation(v.(string))
				}

				if v := track["property"]; v != nil {
					filterTrackSelection.Property = media.FilterTrackPropertyType(v.(string))
				}

				if v := track["value"]; v != nil {
					filterTrackSelection.Value = utils.String(v.(string))
				}

				filterTrackSelections = append(filterTrackSelections, filterTrackSelection)
			}

			results = append(results, media.FilterTrackSelection{
				TrackSelections: &filterTrackSelections,
			})
		}
	}

	return &results
}

func flattenTracks(input *[]media.FilterTrackSelection) []interface{} {
	tracks := make([]interface{}, 0)

	for _, v := range *input {
		selections := make([]interface{}, 0)
		if v.TrackSelections != nil {
			for _, selection := range *v.TrackSelections {
				value := ""
				if selection.Value != nil {
					value = *selection.Value
				}

				selections = append(selections, map[string]interface{}{
					"operation": string(selection.Operation),
					"property":  string(selection.Property),
					"value":     value,
				})
			}
		}
		tracks = append(tracks, map[string]interface{}{
			"condition": selections,
		})
	}

	return tracks
}
