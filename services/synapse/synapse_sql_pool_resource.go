package synapse

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/synapse/mgmt/2021-03-01/synapse"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	mssqlParse "github.com/hashicorp/terraform-provider-azurerm/services/mssql/parse"
	mssqlValidate "github.com/hashicorp/terraform-provider-azurerm/services/mssql/validate"
	"github.com/hashicorp/terraform-provider-azurerm/services/synapse/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/synapse/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

const (
	DefaultCreateMode            = "Default"
	RecoveryCreateMode           = "Recovery"
	PointInTimeRestoreCreateMode = "PointInTimeRestore"
)

func resourceSynapseSqlPool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSynapseSqlPoolCreate,
		Read:   resourceSynapseSqlPoolRead,
		Update: resourceSynapseSqlPoolUpdate,
		Delete: resourceSynapseSqlPoolDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.SqlPoolID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			d.Set("create_mode", DefaultCreateMode)
			if v, ok := d.GetOk("create_mode"); ok && v.(string) != "" {
				d.Set("create_mode", v)
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SqlPoolName,
			},

			"synapse_workspace_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.WorkspaceID,
			},

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"DW100c",
					"DW200c",
					"DW300c",
					"DW400c",
					"DW500c",
					"DW1000c",
					"DW1500c",
					"DW2000c",
					"DW2500c",
					"DW3000c",
					"DW5000c",
					"DW6000c",
					"DW7500c",
					"DW10000c",
					"DW15000c",
					"DW30000c",
				}, false),
			},

			"create_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  DefaultCreateMode,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					DefaultCreateMode,
					RecoveryCreateMode,
					PointInTimeRestoreCreateMode,
				}, false),
			},

			"collation": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: mssqlValidate.DatabaseCollation(),
			},

			"recovery_database_id": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"restore"},
				ValidateFunc: validation.Any(
					validate.SqlPoolID,
					mssqlValidate.DatabaseID,
				),
			},

			"restore": {
				Type:          pluginsdk.TypeList,
				ForceNew:      true,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"recovery_database_id"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"point_in_time": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},

						"source_database_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.Any(
								validate.SqlPoolID,
								mssqlValidate.DatabaseID,
							),
						},
					},
				},
			},

			"data_encrypted": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceSynapseSqlPoolCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	workspaceClient := meta.(*clients.Client).Synapse.WorkspaceClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	workspaceId, err := parse.WorkspaceID(d.Get("synapse_workspace_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewSqlPoolID(workspaceId.SubscriptionId, workspaceId.ResourceGroup, workspaceId.Name, d.Get("name").(string))
	existing, err := sqlClient.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_synapse_sql_pool", id.ID())
	}

	workspace, err := workspaceClient.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name)
	if err != nil {
		return fmt.Errorf("retrieving Synapse Workspace %q (Resource Group %q): %+v", workspaceId.Name, workspaceId.ResourceGroup, err)
	}

	mode := d.Get("create_mode").(string)
	sqlPoolInfo := synapse.SQLPool{
		Location: workspace.Location,
		SQLPoolResourceProperties: &synapse.SQLPoolResourceProperties{
			CreateMode: utils.String(mode),
		},
		Sku: &synapse.Sku{
			Name: utils.String(d.Get("sku_name").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	switch mode {
	case DefaultCreateMode:
		sqlPoolInfo.SQLPoolResourceProperties.Collation = utils.String(d.Get("collation").(string))
	case RecoveryCreateMode:
		recoveryDatabaseId := constructSourceDatabaseId(d.Get("recovery_database_id").(string))
		if recoveryDatabaseId == "" {
			return fmt.Errorf("`recovery_database_id` must be set when `create_mode` is %q", RecoveryCreateMode)
		}
		sqlPoolInfo.SQLPoolResourceProperties.RecoverableDatabaseID = utils.String(recoveryDatabaseId)
	case PointInTimeRestoreCreateMode:
		restore := d.Get("restore").([]interface{})
		if len(restore) == 0 || restore[0] == nil {
			return fmt.Errorf("`restore` block must be set when `create_mode` is %q", PointInTimeRestoreCreateMode)
		}
		v := restore[0].(map[string]interface{})
		sourceDatabaseId := constructSourceDatabaseId(v["source_database_id"].(string))
		vTime, parseErr := date.ParseTime(time.RFC3339, v["point_in_time"].(string))
		if parseErr != nil {
			return fmt.Errorf("parsing time format: %+v", parseErr)
		}
		sqlPoolInfo.SQLPoolResourceProperties.RestorePointInTime = &date.Time{Time: vTime}
		sqlPoolInfo.SQLPoolResourceProperties.SourceDatabaseID = utils.String(sourceDatabaseId)
	}

	future, err := sqlClient.Create(ctx, id.ResourceGroup, id.WorkspaceName, id.Name, sqlPoolInfo)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}
	if err = future.WaitForCompletionRef(ctx, sqlClient.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	if d.Get("data_encrypted").(bool) {
		parameter := synapse.TransparentDataEncryption{
			TransparentDataEncryptionProperties: &synapse.TransparentDataEncryptionProperties{
				Status: synapse.TransparentDataEncryptionStatusEnabled,
			},
		}
		if _, err := sqlPTDEClient.CreateOrUpdate(ctx, id.ResourceGroup, id.WorkspaceName, id.Name, parameter); err != nil {
			return fmt.Errorf("setting `data_encrypted`: %+v", err)
		}
	}

	d.SetId(id.ID())
	return resourceSynapseSqlPoolRead(d, meta)
}

func resourceSynapseSqlPoolUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SqlPoolID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("data_encrypted") {
		status := synapse.TransparentDataEncryptionStatusDisabled
		if d.Get("data_encrypted").(bool) {
			status = synapse.TransparentDataEncryptionStatusEnabled
		}

		parameter := synapse.TransparentDataEncryption{
			TransparentDataEncryptionProperties: &synapse.TransparentDataEncryptionProperties{
				Status: status,
			},
		}
		if _, err := sqlPTDEClient.CreateOrUpdate(ctx, id.ResourceGroup, id.WorkspaceName, id.Name, parameter); err != nil {
			return fmt.Errorf("updating `data_encrypted`: %+v", err)
		}
	}

	if d.HasChanges("sku_name", "tags") {
		sqlPoolInfo := synapse.SQLPoolPatchInfo{
			Sku: &synapse.Sku{
				Name: utils.String(d.Get("sku_name").(string)),
			},
			Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
		}

		if _, err := sqlClient.Update(ctx, id.ResourceGroup, id.WorkspaceName, id.Name, sqlPoolInfo); err != nil {
			return fmt.Errorf("updating %s: %+v", *id, err)
		}

		// wait for sku scale completion
		if d.HasChange("sku_name") {
			deadline, ok := ctx.Deadline()
			if !ok {
				return fmt.Errorf("context had no deadline")
			}
			stateConf := &pluginsdk.StateChangeConf{
				Pending: []string{
					"Scaling",
				},
				Target: []string{
					"Online",
				},
				Refresh:                   synapseSqlPoolScaleStateRefreshFunc(ctx, sqlClient, id.ResourceGroup, id.WorkspaceName, id.Name),
				MinTimeout:                5 * time.Second,
				ContinuousTargetOccurence: 3,
				Timeout:                   time.Until(deadline),
			}

			if _, err := stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("waiting for scaling of %s: %+v", *id, err)
			}
		}
	}
	return resourceSynapseSqlPoolRead(d, meta)
}

func resourceSynapseSqlPoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SqlPoolID(d.Id())
	if err != nil {
		return err
	}

	resp, err := sqlClient.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	transparentDataEncryption, err := sqlPTDEClient.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Transparent Data Encryption settings of Synapse SqlPool %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}

	workspaceId := parse.NewWorkspaceID(id.SubscriptionId, id.ResourceGroup, id.WorkspaceName).ID()
	d.Set("name", id.Name)
	d.Set("synapse_workspace_id", workspaceId)
	if resp.Sku != nil {
		d.Set("sku_name", resp.Sku.Name)
	}
	if props := resp.SQLPoolResourceProperties; props != nil {
		d.Set("collation", props.Collation)
	}
	if props := transparentDataEncryption.TransparentDataEncryptionProperties; props != nil {
		d.Set("data_encrypted", props.Status == synapse.TransparentDataEncryptionStatusEnabled)
	}

	// whole "restore" block is not returned. to avoid conflict, so set it from the old state
	d.Set("restore", d.Get("restore").([]interface{}))

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceSynapseSqlPoolDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SqlPoolID(d.Id())
	if err != nil {
		return err
	}

	future, err := sqlClient.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Synapse Sql Pool %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, sqlClient.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Synapse Sql Pool %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}
	return nil
}

func synapseSqlPoolScaleStateRefreshFunc(ctx context.Context, client *synapse.SQLPoolsClient, resourceGroup, workspaceName, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.Get(ctx, resourceGroup, workspaceName, name)
		if err != nil {
			return resp, "failed", err
		}
		if resp.SQLPoolResourceProperties == nil || resp.SQLPoolResourceProperties.Status == nil {
			return resp, "failed", nil
		}
		return resp, *resp.SQLPoolResourceProperties.Status, nil
	}
}

// sqlPool backend service is a proxy to sql database
// backend service restore and backup only accept id format of sql database
// so if the id is sqlPool, we need to construct the corresponding sql database id
func constructSourceDatabaseId(id string) string {
	sqlPoolId, err := parse.SqlPoolID(id)
	if err != nil {
		return id
	}
	return mssqlParse.NewDatabaseID(sqlPoolId.SubscriptionId, sqlPoolId.ResourceGroup, sqlPoolId.WorkspaceName, sqlPoolId.Name).ID()
}
