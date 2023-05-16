package streamanalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/streamanalytics/mgmt/2020-03-01/streamanalytics"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type OutputPowerBIResource struct{}

var _ sdk.ResourceWithCustomImporter = OutputPowerBIResource{}

type OutputPowerBIResourceModel struct {
	Name               string `tfschema:"name"`
	StreamAnalyticsJob string `tfschema:"stream_analytics_job_id"`
	DataSet            string `tfschema:"dataset"`
	Table              string `tfschema:"table"`
	GroupID            string `tfschema:"group_id"`
	GroupName          string `tfschema:"group_name"`
}

func (r OutputPowerBIResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"stream_analytics_job_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.StreamingJobID,
		},

		"dataset": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"table": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"group_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},

		"group_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func (r OutputPowerBIResource) Attributes() map[string]*schema.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r OutputPowerBIResource) ModelObject() interface{} {
	return &OutputPowerBIResourceModel{}
}

func (r OutputPowerBIResource) ResourceType() string {
	return "azurerm_stream_analytics_output_powerbi"
}

func (r OutputPowerBIResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model OutputPowerBIResourceModel
			if err := metadata.Decode(&model); err != nil {
				return err
			}

			client := metadata.Client.StreamAnalytics.OutputsClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			streamingJobStruct, err := parse.StreamingJobID(model.StreamAnalyticsJob)
			if err != nil {
				return err
			}
			id := parse.NewOutputID(subscriptionId, streamingJobStruct.ResourceGroup, streamingJobStruct.Name, model.Name)

			existing, err := client.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.Name)
			if err != nil && !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}

			if !utils.ResponseWasNotFound(existing.Response) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			powerbiOutputProps := &streamanalytics.PowerBIOutputDataSourceProperties{
				Dataset:            utils.String(model.DataSet),
				Table:              utils.String(model.Table),
				GroupID:            utils.String(model.GroupID),
				GroupName:          utils.String(model.GroupName),
				AuthenticationMode: streamanalytics.AuthenticationMode("Msi"), // Set authentication mode as "Msi" here since other modes requires params obtainable from portal only.
			}

			props := streamanalytics.Output{
				Name: utils.String(model.Name),
				OutputProperties: &streamanalytics.OutputProperties{
					Datasource: &streamanalytics.PowerBIOutputDataSource{
						Type:                              streamanalytics.TypeBasicOutputDataSourceTypePowerBI,
						PowerBIOutputDataSourceProperties: powerbiOutputProps,
					},
				},
			}

			if _, err = client.CreateOrReplace(ctx, props, id.ResourceGroup, id.StreamingjobName, id.Name, "", ""); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (r OutputPowerBIResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StreamAnalytics.OutputsClient
			id, err := parse.OutputID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var state OutputPowerBIResourceModel
			if err := metadata.Decode(&state); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			needUpdateDataSourceProps := false
			dataSourceProps := streamanalytics.PowerBIOutputDataSourceProperties{}
			d := metadata.ResourceData

			if d.HasChange("dataset") {
				needUpdateDataSourceProps = true
				dataSourceProps.Dataset = &state.DataSet
			}

			if d.HasChange("table") {
				needUpdateDataSourceProps = true
				dataSourceProps.Table = &state.Table
			}

			if d.HasChange("group_name") {
				needUpdateDataSourceProps = true
				dataSourceProps.GroupName = &state.GroupName
			}

			if d.HasChange("group_id") {
				needUpdateDataSourceProps = true
				dataSourceProps.GroupID = &state.GroupID
			}

			if !needUpdateDataSourceProps {
				return nil
			}

			updateDataSource := streamanalytics.PowerBIOutputDataSource{
				Type:                              streamanalytics.TypeBasicOutputDataSourceTypePowerBI,
				PowerBIOutputDataSourceProperties: &dataSourceProps,
			}

			props := streamanalytics.Output{
				OutputProperties: &streamanalytics.OutputProperties{
					Datasource: updateDataSource,
				},
			}

			if _, err = client.Update(ctx, props, id.ResourceGroup, id.StreamingjobName, id.Name, ""); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (r OutputPowerBIResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StreamAnalytics.OutputsClient
			id, err := parse.OutputID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.Name)
			if err != nil {
				if utils.ResponseWasNotFound(resp.Response) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("reading %s: %+v", *id, err)
			}

			if props := resp.OutputProperties; props != nil && props.Datasource != nil {
				v, ok := props.Datasource.AsPowerBIOutputDataSource()
				if !ok {
					return fmt.Errorf("converting output data source to a powerBI output: %+v", err)
				}

				streamingJobId := parse.NewStreamingJobID(id.SubscriptionId, id.ResourceGroup, id.StreamingjobName)

				state := OutputPowerBIResourceModel{
					Name:               id.Name,
					StreamAnalyticsJob: streamingJobId.ID(),
				}

				if v.Dataset != nil {
					state.DataSet = *v.Dataset
				}

				if v.Table != nil {
					state.Table = *v.Table
				}

				if v.GroupID != nil {
					state.GroupID = *v.GroupID
				}

				if v.GroupName != nil {
					state.GroupName = *v.GroupName
				}

				return metadata.Encode(&state)
			}
			return nil
		},
	}
}

func (r OutputPowerBIResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.StreamAnalytics.OutputsClient
			id, err := parse.OutputID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Infof("deleting %s", *id)

			if resp, err := client.Delete(ctx, id.ResourceGroup, id.StreamingjobName, id.Name); err != nil {
				if !response.WasNotFound(resp.Response) {
					return fmt.Errorf("deleting %s: %+v", *id, err)
				}
			}
			return nil
		},
	}
}

func (r OutputPowerBIResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.OutputID
}

func (r OutputPowerBIResource) CustomImporter() sdk.ResourceRunFunc {
	return func(ctx context.Context, metadata sdk.ResourceMetaData) error {
		id, err := parse.OutputID(metadata.ResourceData.Id())
		if err != nil {
			return err
		}

		client := metadata.Client.StreamAnalytics.OutputsClient
		resp, err := client.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.Name)
		if err != nil || resp.OutputProperties == nil {
			return fmt.Errorf("reading %s: %+v", *id, err)
		}

		props := resp.OutputProperties
		if _, ok := props.Datasource.AsPowerBIOutputDataSource(); !ok {
			return fmt.Errorf("specified output is not of type %s", streamanalytics.TypeBasicOutputDataSourceTypePowerBI)
		}
		return nil
	}
}
