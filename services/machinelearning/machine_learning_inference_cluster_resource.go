package machinelearning

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/machinelearningservices/mgmt/2021-07-01/machinelearningservices"
	"github.com/Azure/azure-sdk-for-go/services/preview/containerservice/mgmt/2022-01-02-preview/containerservice"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/machinelearning/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/machinelearning/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceAksInferenceCluster() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceAksInferenceClusterCreate,
		Read:   resourceAksInferenceClusterRead,
		Delete: resourceAksInferenceClusterDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.InferenceClusterID(id)
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
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"kubernetes_cluster_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.KubernetesClusterID,
				// remove in 3.0 of the provider
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"location": azure.SchemaLocation(),

			"machine_learning_workspace_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cluster_purpose": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  string(machinelearningservices.ClusterPurposeFastProd),
				ValidateFunc: validation.StringInSlice([]string{
					string(machinelearningservices.ClusterPurposeDevTest),
					string(machinelearningservices.ClusterPurposeFastProd),
					string(machinelearningservices.ClusterPurposeDenseProd),
				}, false),
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"identity": commonschema.SystemAssignedUserAssignedIdentityOptionalForceNew(),

			"ssl": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"cert": {
							Type:          pluginsdk.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       "",
							ConflictsWith: []string{"ssl.0.leaf_domain_label", "ssl.0.overwrite_existing_domain"},
						},
						"key": {
							Type:          pluginsdk.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       "",
							ConflictsWith: []string{"ssl.0.leaf_domain_label", "ssl.0.overwrite_existing_domain"},
						},
						"cname": {
							Type:          pluginsdk.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       "",
							ConflictsWith: []string{"ssl.0.leaf_domain_label", "ssl.0.overwrite_existing_domain"},
						},
						"leaf_domain_label": {
							Type:          pluginsdk.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       "",
							ConflictsWith: []string{"ssl.0.cert", "ssl.0.key", "ssl.0.cname"},
						},
						"overwrite_existing_domain": {
							Type:          pluginsdk.TypeBool,
							Optional:      true,
							ForceNew:      true,
							Default:       "",
							ConflictsWith: []string{"ssl.0.cert", "ssl.0.key", "ssl.0.cname"},
						},
					},
				},
			},

			"tags": tags.ForceNewSchema(),
		},
	}
}

func resourceAksInferenceClusterCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MachineLearning.ComputeClient
	aksClient := meta.(*clients.Client).Containers.KubernetesClustersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	// Define Inference Cluster Name
	name := d.Get("name").(string)

	// Get Machine Learning Workspace Name and Resource Group from ID
	workspaceID, err := parse.WorkspaceID(d.Get("machine_learning_workspace_id").(string))
	if err != nil {
		return err
	}

	// Check if Inference Cluster already exists
	existing, err := client.Get(ctx, workspaceID.ResourceGroup, workspaceID.Name, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for existing Inference Cluster %q in Workspace %q (Resource Group %q): %s", name, workspaceID.Name, workspaceID.ResourceGroup, err)
		}
	}
	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_machine_learning_inference_cluster", *existing.ID)
	}

	// Get AKS Compute Properties
	aksID, err := parse.KubernetesClusterID(d.Get("kubernetes_cluster_id").(string))
	if err != nil {
		return err
	}
	aks, err := aksClient.Get(ctx, aksID.ResourceGroup, aksID.ManagedClusterName)
	if err != nil {
		return err
	}

	identity, err := expandIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf("expanding `identity`: %+v", err)
	}

	inferenceClusterParameters := machinelearningservices.ComputeResource{
		Properties: expandAksComputeProperties(&aks, d),
		Identity:   identity,
		Location:   utils.String(azure.NormalizeLocation(d.Get("location").(string))),
		Tags:       tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.CreateOrUpdate(ctx, workspaceID.ResourceGroup, workspaceID.Name, name, inferenceClusterParameters)
	if err != nil {
		return fmt.Errorf("creating Inference Cluster %q in workspace %q (Resource Group %q): %+v", name, workspaceID.Name, workspaceID.ResourceGroup, err)
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Inference Cluster %q in workspace %q (Resource Group %q): %+v", name, workspaceID.Name, workspaceID.ResourceGroup, err)
	}
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	id := parse.NewInferenceClusterID(subscriptionId, workspaceID.ResourceGroup, workspaceID.Name, name)
	d.SetId(id.ID())

	return resourceAksInferenceClusterRead(d, meta)
}

func resourceAksInferenceClusterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MachineLearning.ComputeClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.InferenceClusterID(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", id.ComputeName)

	// Check that Inference Cluster Response can be read
	computeResource, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.ComputeName)
	if err != nil {
		if utils.ResponseWasNotFound(computeResource.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on Inference Cluster %q in Workspace %q (Resource Group %q): %+v",
			id.ComputeName, id.WorkspaceName, id.ResourceGroup, err)
	}

	// Retrieve Machine Learning Workspace ID
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	workspaceId := parse.NewWorkspaceID(subscriptionId, id.ResourceGroup, id.WorkspaceName)
	d.Set("machine_learning_workspace_id", workspaceId.ID())

	// use ComputeResource to get to AKS Cluster ID and other properties
	aksComputeProperties, isAks := (machinelearningservices.BasicCompute).AsAKS(computeResource.Properties)
	if !isAks {
		return fmt.Errorf("compute resource %s is not an AKS cluster", id.ComputeName)
	}

	// Retrieve AKS Cluster ID
	aksId, err := parse.KubernetesClusterID(*aksComputeProperties.ResourceID)
	if err != nil {
		return err
	}
	d.Set("kubernetes_cluster_id", aksId.ID())
	d.Set("cluster_purpose", string(aksComputeProperties.Properties.ClusterPurpose))
	d.Set("description", aksComputeProperties.Description)

	// Retrieve location
	if location := computeResource.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	identity, err := flattenIdentity(computeResource.Identity)
	if err != nil {
		return fmt.Errorf("flattening `identity`: %+v", err)
	}
	if err := d.Set("identity", identity); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	return tags.FlattenAndSet(d, computeResource.Tags)
}

func resourceAksInferenceClusterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MachineLearning.ComputeClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()
	id, err := parse.InferenceClusterID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.ComputeName, machinelearningservices.UnderlyingResourceActionDetach)
	if err != nil {
		return fmt.Errorf("deleting Inference Cluster %q in workspace %q (Resource Group %q): %+v",
			id.ComputeName, id.WorkspaceName, id.ResourceGroup, err)
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Inference Cluster %q in workspace %q (Resource Group %q): %+v",
			id.ComputeName, id.WorkspaceName, id.ResourceGroup, err)
	}
	return nil
}

func expandAksComputeProperties(aks *containerservice.ManagedCluster, d *pluginsdk.ResourceData) machinelearningservices.AKS {
	fqdn := aks.PrivateFQDN
	if fqdn == nil {
		fqdn = aks.Fqdn
	}

	return machinelearningservices.AKS{
		Properties: &machinelearningservices.AKSProperties{
			ClusterFqdn:      utils.String(*fqdn),
			SslConfiguration: expandSSLConfig(d.Get("ssl").([]interface{})),
			ClusterPurpose:   machinelearningservices.ClusterPurpose(d.Get("cluster_purpose").(string)),
		},
		ComputeLocation: aks.Location,
		Description:     utils.String(d.Get("description").(string)),
		ResourceID:      aks.ID,
	}
}

func expandSSLConfig(input []interface{}) *machinelearningservices.SslConfiguration {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	// SSL Certificate default values
	sslStatus := "Disabled"

	if !(v["cert"].(string) == "" && v["key"].(string) == "" && v["cname"].(string) == "") {
		sslStatus = "Enabled"
	}

	if !(v["leaf_domain_label"].(string) == "") {
		sslStatus = "Auto"
		v["cname"] = ""
	}

	return &machinelearningservices.SslConfiguration{
		Status:                  machinelearningservices.Status1(sslStatus),
		Cert:                    utils.String(v["cert"].(string)),
		Key:                     utils.String(v["key"].(string)),
		Cname:                   utils.String(v["cname"].(string)),
		LeafDomainLabel:         utils.String(v["leaf_domain_label"].(string)),
		OverwriteExistingDomain: utils.Bool(v["overwrite_existing_domain"].(bool)),
	}
}
