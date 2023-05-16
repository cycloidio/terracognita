package streamanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/streamanalytics/mgmt/2020-03-01/streamanalytics"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/streamanalytics/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStreamAnalyticsOutputSql() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStreamAnalyticsOutputSqlCreateUpdate,
		Read:   resourceStreamAnalyticsOutputSqlRead,
		Update: resourceStreamAnalyticsOutputSqlCreateUpdate,
		Delete: resourceStreamAnalyticsOutputSqlDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.OutputID(id)
			return err
		}),

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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"stream_analytics_job_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"server": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"database": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"table": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"user": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"password": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"max_batch_count": {
				Type:         pluginsdk.TypeFloat,
				Optional:     true,
				Default:      10000,
				ValidateFunc: validation.FloatBetween(1, 1073741824),
			},

			"max_writer_count": {
				Type:         pluginsdk.TypeFloat,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.FloatBetween(0, 1),
			},
		},
	}
}

func resourceStreamAnalyticsOutputSqlCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.OutputsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewOutputID(subscriptionId, d.Get("resource_group_name").(string), d.Get("stream_analytics_job_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.Name)
		if err != nil && !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for existing %s: %+v", id, err)
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_stream_analytics_output_mssql", id.ID())
		}
	}

	server := d.Get("server").(string)
	databaseName := d.Get("database").(string)
	tableName := d.Get("table").(string)
	sqlUser := d.Get("user").(string)
	sqlUserPassword := d.Get("password").(string)

	props := streamanalytics.Output{
		Name: utils.String(id.Name),
		OutputProperties: &streamanalytics.OutputProperties{
			Datasource: &streamanalytics.AzureSQLDatabaseOutputDataSource{
				Type: streamanalytics.TypeBasicOutputDataSourceTypeMicrosoftSQLServerDatabase,
				AzureSQLDatabaseOutputDataSourceProperties: &streamanalytics.AzureSQLDatabaseOutputDataSourceProperties{
					Server:         utils.String(server),
					Database:       utils.String(databaseName),
					User:           utils.String(sqlUser),
					Password:       utils.String(sqlUserPassword),
					Table:          utils.String(tableName),
					MaxBatchCount:  utils.Float(d.Get("max_batch_count").(float64)),
					MaxWriterCount: utils.Float(d.Get("max_writer_count").(float64)),
				},
			},
		},
	}

	if d.IsNewResource() {
		if _, err := client.CreateOrReplace(ctx, props, id.ResourceGroup, id.StreamingjobName, id.Name, "", ""); err != nil {
			return fmt.Errorf("creating %s: %+v", id, err)
		}

		d.SetId(id.ID())
	} else if _, err := client.Update(ctx, props, id.ResourceGroup, id.StreamingjobName, id.Name, ""); err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	return resourceStreamAnalyticsOutputSqlRead(d, meta)
}

func resourceStreamAnalyticsOutputSqlRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.OutputsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.OutputID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.StreamingjobName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retreving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("stream_analytics_job_name", id.StreamingjobName)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.OutputProperties; props != nil {
		v, ok := props.Datasource.AsAzureSQLDatabaseOutputDataSource()
		if !ok {
			return fmt.Errorf("converting Output Data Source to SQL Output: %+v", err)
		}

		d.Set("server", v.Server)
		d.Set("database", v.Database)
		d.Set("table", v.Table)
		d.Set("user", v.User)

		maxBatchCount := float64(10000)
		if v.MaxBatchCount != nil {
			maxBatchCount = *v.MaxBatchCount
		}
		d.Set("max_batch_count", maxBatchCount)

		maxWriterCount := float64(1)
		if v.MaxWriterCount != nil {
			maxWriterCount = *v.MaxWriterCount
		}
		d.Set("max_writer_count", maxWriterCount)
	}

	return nil
}

func resourceStreamAnalyticsOutputSqlDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.OutputsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.OutputID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, id.ResourceGroup, id.StreamingjobName, id.Name); err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	return nil
}
