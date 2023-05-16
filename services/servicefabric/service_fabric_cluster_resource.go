package servicefabric

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/servicefabric/mgmt/2021-06-01/servicefabric"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/servicefabric/parse"
	serviceFabricValidate "github.com/hashicorp/terraform-provider-azurerm/services/servicefabric/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tags"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceFabricCluster() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceFabricClusterCreateUpdate,
		Read:   resourceServiceFabricClusterRead,
		Update: resourceServiceFabricClusterCreateUpdate,
		Delete: resourceServiceFabricClusterDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ClusterID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"reliability_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicefabric.ReliabilityLevelNone),
					string(servicefabric.ReliabilityLevelBronze),
					string(servicefabric.ReliabilityLevelSilver),
					string(servicefabric.ReliabilityLevelGold),
					string(servicefabric.ReliabilityLevelPlatinum),
				}, false),
			},

			"upgrade_mode": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicefabric.UpgradeModeAutomatic),
					string(servicefabric.UpgradeModeManual),
				}, false),
			},

			"service_fabric_zonal_upgrade_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicefabric.SfZonalUpgradeModeHierarchical),
					string(servicefabric.SfZonalUpgradeModeParallel),
				}, false),
			},

			"vmss_zonal_upgrade_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(servicefabric.VmssZonalUpgradeModeHierarchical),
					string(servicefabric.VmssZonalUpgradeModeParallel),
				}, false),
			},

			"cluster_code_version": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"management_endpoint": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vm_image": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"add_on_features": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"azure_active_directory": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"tenant_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"cluster_application_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"client_application_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
					},
				},
			},

			"certificate": {
				Type:          pluginsdk.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"certificate_common_names"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"thumbprint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"thumbprint_secondary": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"x509_store_name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"certificate_common_names": {
				Type:          pluginsdk.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"certificate"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"common_names": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"certificate_common_name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"certificate_issuer_thumbprint": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"x509_store_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"reverse_proxy_certificate": {
				Type:          pluginsdk.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"reverse_proxy_certificate_common_names"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"thumbprint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"thumbprint_secondary": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"x509_store_name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"reverse_proxy_certificate_common_names": {
				Type:          pluginsdk.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"reverse_proxy_certificate"},
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"common_names": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"certificate_common_name": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"certificate_issuer_thumbprint": {
										Type:         pluginsdk.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"x509_store_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"client_certificate_thumbprint": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"thumbprint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"is_admin": {
							Type:     pluginsdk.TypeBool,
							Required: true,
						},
					},
				},
			},

			"client_certificate_common_name": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"common_name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"issuer_thumbprint": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							// todo remove this when https://github.com/Azure/azure-sdk-for-go/issues/17744 is fixed
							DiffSuppressFunc: suppress.CaseDifference,
						},
						"is_admin": {
							Type:     pluginsdk.TypeBool,
							Required: true,
						},
					},
				},
			},

			"diagnostics_config": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"storage_account_name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"protected_account_key_name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"blob_endpoint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"queue_endpoint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"table_endpoint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"upgrade_policy": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"force_restart_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
						"health_check_retry_timeout": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "00:45:00",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"health_check_stable_duration": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "00:01:00",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"health_check_wait_duration": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "00:00:30",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"upgrade_domain_timeout": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "02:00:00",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"upgrade_replica_set_check_timeout": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "10675199.02:48:05.4775807",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"upgrade_timeout": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "12:00:00",
							ValidateFunc: serviceFabricValidate.UpgradeTimeout,
						},
						"health_policy": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"max_unhealthy_applications_percent": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"max_unhealthy_nodes_percent": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 100),
									},
								},
							},
						},
						"delta_health_policy": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"max_delta_unhealthy_applications_percent": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"max_delta_unhealthy_nodes_percent": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"max_upgrade_domain_delta_unhealthy_nodes_percent": {
										Type:         pluginsdk.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 100),
									},
								},
							},
						},
					},
				},
			},

			"fabric_settings": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"parameters": {
							Type:     pluginsdk.TypeMap,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
			},

			"node_type": {
				Type:     pluginsdk.TypeList,
				Required: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"placement_properties": {
							Type:     pluginsdk.TypeMap,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
						"capacities": {
							Type:     pluginsdk.TypeMap,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
						"instance_count": {
							Type:     pluginsdk.TypeInt,
							Required: true,
						},
						"is_primary": {
							Type:     pluginsdk.TypeBool,
							Required: true,
						},
						"is_stateless": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
						"multiple_availability_zones": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
						"client_endpoint_port": {
							Type:     pluginsdk.TypeInt,
							Required: true,
						},
						"http_endpoint_port": {
							Type:     pluginsdk.TypeInt,
							Required: true,
						},
						"reverse_proxy_endpoint_port": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							ValidateFunc: validate.PortNumber,
						},
						"durability_level": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  string(servicefabric.DurabilityLevelBronze),
							ValidateFunc: validation.StringInSlice([]string{
								string(servicefabric.DurabilityLevelBronze),
								string(servicefabric.DurabilityLevelSilver),
								string(servicefabric.DurabilityLevelGold),
							}, false),
						},

						"application_ports": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"start_port": {
										Type:     pluginsdk.TypeInt,
										Required: true,
									},
									"end_port": {
										Type:     pluginsdk.TypeInt,
										Required: true,
									},
								},
							},
						},

						"ephemeral_ports": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"start_port": {
										Type:     pluginsdk.TypeInt,
										Required: true,
									},
									"end_port": {
										Type:     pluginsdk.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"tags": tags.Schema(),

			"cluster_endpoint": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceFabricClusterCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabric.ClustersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewClusterID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_service_fabric_cluster", id.ID())
		}
	}

	addOnFeaturesRaw := d.Get("add_on_features").(*pluginsdk.Set).List()
	addOnFeatures := expandServiceFabricClusterAddOnFeatures(addOnFeaturesRaw)

	azureActiveDirectoryRaw := d.Get("azure_active_directory").([]interface{})
	azureActiveDirectory := expandServiceFabricClusterAzureActiveDirectory(azureActiveDirectoryRaw)

	diagnosticsRaw := d.Get("diagnostics_config").([]interface{})
	diagnostics := expandServiceFabricClusterDiagnosticsConfig(diagnosticsRaw)

	upgradePolicyRaw := d.Get("upgrade_policy").([]interface{})
	upgradePolicy := expandServiceFabricClusterUpgradePolicy(upgradePolicyRaw)

	fabricSettingsRaw := d.Get("fabric_settings").([]interface{})
	fabricSettings := expandServiceFabricClusterFabricSettings(fabricSettingsRaw)

	nodeTypesRaw := d.Get("node_type").([]interface{})
	nodeTypes := expandServiceFabricClusterNodeTypes(nodeTypesRaw)

	location := d.Get("location").(string)
	reliabilityLevel := d.Get("reliability_level").(string)
	managementEndpoint := d.Get("management_endpoint").(string)
	upgradeMode := d.Get("upgrade_mode").(string)
	clusterCodeVersion := d.Get("cluster_code_version").(string)
	vmImage := d.Get("vm_image").(string)
	t := d.Get("tags").(map[string]interface{})

	cluster := servicefabric.Cluster{
		Location: utils.String(location),
		Tags:     tags.Expand(t),
		ClusterProperties: &servicefabric.ClusterProperties{
			AddOnFeatures:                      addOnFeatures,
			AzureActiveDirectory:               azureActiveDirectory,
			CertificateCommonNames:             expandServiceFabricClusterCertificateCommonNames(d),
			ReverseProxyCertificateCommonNames: expandServiceFabricClusterReverseProxyCertificateCommonNames(d),
			DiagnosticsStorageAccountConfig:    diagnostics,
			FabricSettings:                     fabricSettings,
			ManagementEndpoint:                 utils.String(managementEndpoint),
			NodeTypes:                          nodeTypes,
			ReliabilityLevel:                   servicefabric.ReliabilityLevel(reliabilityLevel),
			UpgradeDescription:                 upgradePolicy,
			UpgradeMode:                        servicefabric.UpgradeMode(upgradeMode),
			VMImage:                            utils.String(vmImage),
		},
	}

	if sfZonalUpgradeMode, ok := d.GetOk("service_fabric_zonal_upgrade_mode"); ok {
		cluster.ClusterProperties.SfZonalUpgradeMode = servicefabric.SfZonalUpgradeMode(sfZonalUpgradeMode.(string))
	}

	if vmssZonalUpgradeMode, ok := d.GetOk("vmss_zonal_upgrade_mode"); ok {
		cluster.ClusterProperties.VmssZonalUpgradeMode = servicefabric.VmssZonalUpgradeMode(vmssZonalUpgradeMode.(string))
	}

	if certificateRaw, ok := d.GetOk("certificate"); ok {
		certificate := expandServiceFabricClusterCertificate(certificateRaw.([]interface{}))
		cluster.ClusterProperties.Certificate = certificate
	}

	if reverseProxyCertificateRaw, ok := d.GetOk("reverse_proxy_certificate"); ok {
		reverseProxyCertificate := expandServiceFabricClusterReverseProxyCertificate(reverseProxyCertificateRaw.([]interface{}))
		cluster.ClusterProperties.ReverseProxyCertificate = reverseProxyCertificate
	}

	if clientCertificateThumbprintRaw, ok := d.GetOk("client_certificate_thumbprint"); ok {
		clientCertificateThumbprints := expandServiceFabricClusterClientCertificateThumbprints(clientCertificateThumbprintRaw.([]interface{}))
		cluster.ClusterProperties.ClientCertificateThumbprints = clientCertificateThumbprints
	}

	if clientCertificateCommonNamesRaw, ok := d.GetOk("client_certificate_common_name"); ok {
		clientCertificateCommonNames := expandServiceFabricClusterClientCertificateCommonNames(clientCertificateCommonNamesRaw.([]interface{}))
		cluster.ClusterProperties.ClientCertificateCommonNames = clientCertificateCommonNames
	}

	if clusterCodeVersion != "" {
		cluster.ClusterProperties.ClusterCodeVersion = utils.String(clusterCodeVersion)
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, cluster)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceServiceFabricClusterRead(d, meta)
}

func resourceServiceFabricClusterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabric.ClustersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] Service Fabric Cluster %q (Resource Group %q) was not found - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Service Fabric Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.ClusterProperties; props != nil {
		d.Set("cluster_code_version", props.ClusterCodeVersion)
		d.Set("cluster_endpoint", props.ClusterEndpoint)
		d.Set("management_endpoint", props.ManagementEndpoint)
		d.Set("reliability_level", string(props.ReliabilityLevel))
		d.Set("vm_image", props.VMImage)
		d.Set("upgrade_mode", string(props.UpgradeMode))
		d.Set("service_fabric_zonal_upgrade_mode", string(props.SfZonalUpgradeMode))
		d.Set("vmss_zonal_upgrade_mode", string(props.VmssZonalUpgradeMode))

		addOnFeatures := flattenServiceFabricClusterAddOnFeatures(props.AddOnFeatures)
		if err := d.Set("add_on_features", pluginsdk.NewSet(pluginsdk.HashString, addOnFeatures)); err != nil {
			return fmt.Errorf("setting `add_on_features`: %+v", err)
		}

		azureActiveDirectory := flattenServiceFabricClusterAzureActiveDirectory(props.AzureActiveDirectory)
		if err := d.Set("azure_active_directory", azureActiveDirectory); err != nil {
			return fmt.Errorf("setting `azure_active_directory`: %+v", err)
		}

		certificate := flattenServiceFabricClusterCertificate(props.Certificate)
		if err := d.Set("certificate", certificate); err != nil {
			return fmt.Errorf("setting `certificate`: %+v", err)
		}

		certificateCommonNames := flattenServiceFabricClusterCertificateCommonNames(props.CertificateCommonNames)
		if err := d.Set("certificate_common_names", certificateCommonNames); err != nil {
			return fmt.Errorf("setting `certificate_common_names`: %+v", err)
		}

		reverseProxyCertificate := flattenServiceFabricClusterReverseProxyCertificate(props.ReverseProxyCertificate)
		if err := d.Set("reverse_proxy_certificate", reverseProxyCertificate); err != nil {
			return fmt.Errorf("setting `reverse_proxy_certificate`: %+v", err)
		}

		reverseProxyCertificateCommonNames := flattenServiceFabricClusterCertificateCommonNames(props.ReverseProxyCertificateCommonNames)
		if err := d.Set("reverse_proxy_certificate_common_names", reverseProxyCertificateCommonNames); err != nil {
			return fmt.Errorf("setting `reverse_proxy_certificate_common_names`: %+v", err)
		}

		clientCertificateThumbprints := flattenServiceFabricClusterClientCertificateThumbprints(props.ClientCertificateThumbprints)
		if err := d.Set("client_certificate_thumbprint", clientCertificateThumbprints); err != nil {
			return fmt.Errorf("setting `client_certificate_thumbprint`: %+v", err)
		}

		clientCertificateCommonNames := flattenServiceFabricClusterClientCertificateCommonNames(props.ClientCertificateCommonNames)
		if err := d.Set("client_certificate_common_name", clientCertificateCommonNames); err != nil {
			return fmt.Errorf("setting `client_certificate_common_name`: %+v", err)
		}

		diagnostics := flattenServiceFabricClusterDiagnosticsConfig(props.DiagnosticsStorageAccountConfig)
		if err := d.Set("diagnostics_config", diagnostics); err != nil {
			return fmt.Errorf("setting `diagnostics_config`: %+v", err)
		}

		upgradePolicy := flattenServiceFabricClusterUpgradePolicy(props.UpgradeDescription)
		if err := d.Set("upgrade_policy", upgradePolicy); err != nil {
			return fmt.Errorf("setting `upgrade_policy`: %+v", err)
		}

		fabricSettings := flattenServiceFabricClusterFabricSettings(props.FabricSettings)
		if err := d.Set("fabric_settings", fabricSettings); err != nil {
			return fmt.Errorf("setting `fabric_settings`: %+v", err)
		}

		nodeTypes := flattenServiceFabricClusterNodeTypes(props.NodeTypes)
		if err := d.Set("node_type", nodeTypes); err != nil {
			return fmt.Errorf("setting `node_type`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceServiceFabricClusterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabric.ClustersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Service Fabric Cluster %q (Resource Group %q)", id.Name, id.ResourceGroup)

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Service Fabric Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}

func expandServiceFabricClusterAddOnFeatures(input []interface{}) *[]string {
	output := make([]string, 0)

	for _, v := range input {
		output = append(output, v.(string))
	}

	return &output
}

func expandServiceFabricClusterAzureActiveDirectory(input []interface{}) *servicefabric.AzureActiveDirectory {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	tenantId := v["tenant_id"].(string)
	clusterApplication := v["cluster_application_id"].(string)
	clientApplication := v["client_application_id"].(string)

	config := servicefabric.AzureActiveDirectory{
		TenantID:           utils.String(tenantId),
		ClusterApplication: utils.String(clusterApplication),
		ClientApplication:  utils.String(clientApplication),
	}
	return &config
}

func flattenServiceFabricClusterAzureActiveDirectory(input *servicefabric.AzureActiveDirectory) []interface{} {
	results := make([]interface{}, 0)

	if v := input; v != nil {
		output := make(map[string]interface{})

		if name := v.TenantID; name != nil {
			output["tenant_id"] = *name
		}

		if name := v.ClusterApplication; name != nil {
			output["cluster_application_id"] = *name
		}

		if endpoint := v.ClientApplication; endpoint != nil {
			output["client_application_id"] = *endpoint
		}

		results = append(results, output)
	}

	return results
}

func flattenServiceFabricClusterAddOnFeatures(input *[]string) []interface{} {
	output := make([]interface{}, 0)

	if input != nil {
		for _, v := range *input {
			output = append(output, v)
		}
	}

	return output
}

func expandServiceFabricClusterCertificate(input []interface{}) *servicefabric.CertificateDescription {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	thumbprint := v["thumbprint"].(string)
	x509StoreName := v["x509_store_name"].(string)

	result := servicefabric.CertificateDescription{
		Thumbprint:    utils.String(thumbprint),
		X509StoreName: servicefabric.X509StoreName(x509StoreName),
	}

	if thumb, ok := v["thumbprint_secondary"]; ok {
		result.ThumbprintSecondary = utils.String(thumb.(string))
	}

	return &result
}

func flattenServiceFabricClusterCertificate(input *servicefabric.CertificateDescription) []interface{} {
	results := make([]interface{}, 0)

	if v := input; v != nil {
		output := make(map[string]interface{})

		if thumbprint := input.Thumbprint; thumbprint != nil {
			output["thumbprint"] = *thumbprint
		}

		if thumbprint := input.ThumbprintSecondary; thumbprint != nil {
			output["thumbprint_secondary"] = *thumbprint
		}

		output["x509_store_name"] = string(input.X509StoreName)
		results = append(results, output)
	}

	return results
}

func expandServiceFabricClusterCertificateCommonNames(d *pluginsdk.ResourceData) *servicefabric.ServerCertificateCommonNames {
	i := d.Get("certificate_common_names").([]interface{})
	if len(i) == 0 || i[0] == nil {
		return nil
	}
	input := i[0].(map[string]interface{})

	commonNamesRaw := input["common_names"].(*pluginsdk.Set).List()
	commonNames := make([]servicefabric.ServerCertificateCommonName, 0)

	for _, commonName := range commonNamesRaw {
		commonNameDetails := commonName.(map[string]interface{})
		certificateCommonName := commonNameDetails["certificate_common_name"].(string)
		certificateIssuerThumbprint := commonNameDetails["certificate_issuer_thumbprint"].(string)

		commonName := servicefabric.ServerCertificateCommonName{
			CertificateCommonName:       &certificateCommonName,
			CertificateIssuerThumbprint: &certificateIssuerThumbprint,
		}

		commonNames = append(commonNames, commonName)
	}

	x509StoreName := input["x509_store_name"].(string)

	output := servicefabric.ServerCertificateCommonNames{
		CommonNames:   &commonNames,
		X509StoreName: servicefabric.X509StoreName1(x509StoreName),
	}

	return &output
}

func expandServiceFabricClusterReverseProxyCertificateCommonNames(d *pluginsdk.ResourceData) *servicefabric.ServerCertificateCommonNames {
	i := d.Get("reverse_proxy_certificate_common_names").([]interface{})
	if len(i) == 0 || i[0] == nil {
		return nil
	}
	input := i[0].(map[string]interface{})

	commonNamesRaw := input["common_names"].(*pluginsdk.Set).List()
	commonNames := make([]servicefabric.ServerCertificateCommonName, 0)

	for _, commonName := range commonNamesRaw {
		commonNameDetails := commonName.(map[string]interface{})
		certificateCommonName := commonNameDetails["certificate_common_name"].(string)
		certificateIssuerThumbprint := commonNameDetails["certificate_issuer_thumbprint"].(string)

		commonName := servicefabric.ServerCertificateCommonName{
			CertificateCommonName:       &certificateCommonName,
			CertificateIssuerThumbprint: &certificateIssuerThumbprint,
		}

		commonNames = append(commonNames, commonName)
	}

	x509StoreName := input["x509_store_name"].(string)

	output := servicefabric.ServerCertificateCommonNames{
		CommonNames:   &commonNames,
		X509StoreName: servicefabric.X509StoreName1(x509StoreName),
	}

	return &output
}

func flattenServiceFabricClusterCertificateCommonNames(in *servicefabric.ServerCertificateCommonNames) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	output := make(map[string]interface{})

	if commonNames := in.CommonNames; commonNames != nil {
		common_names := make([]map[string]interface{}, 0)
		for _, i := range *commonNames {
			commonName := make(map[string]interface{})

			if i.CertificateCommonName != nil {
				commonName["certificate_common_name"] = *i.CertificateCommonName
			}

			if i.CertificateIssuerThumbprint != nil {
				commonName["certificate_issuer_thumbprint"] = *i.CertificateIssuerThumbprint
			}

			common_names = append(common_names, commonName)
		}

		output["common_names"] = common_names
	}

	output["x509_store_name"] = string(in.X509StoreName)

	return []interface{}{output}
}

func expandServiceFabricClusterReverseProxyCertificate(input []interface{}) *servicefabric.CertificateDescription {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	thumbprint := v["thumbprint"].(string)
	x509StoreName := v["x509_store_name"].(string)

	result := servicefabric.CertificateDescription{
		Thumbprint:    utils.String(thumbprint),
		X509StoreName: servicefabric.X509StoreName(x509StoreName),
	}

	if thumb, ok := v["thumbprint_secondary"]; ok {
		result.ThumbprintSecondary = utils.String(thumb.(string))
	}

	return &result
}

func flattenServiceFabricClusterReverseProxyCertificate(input *servicefabric.CertificateDescription) []interface{} {
	results := make([]interface{}, 0)

	if v := input; v != nil {
		output := make(map[string]interface{})

		if thumbprint := input.Thumbprint; thumbprint != nil {
			output["thumbprint"] = *thumbprint
		}

		if thumbprint := input.ThumbprintSecondary; thumbprint != nil {
			output["thumbprint_secondary"] = *thumbprint
		}

		output["x509_store_name"] = string(input.X509StoreName)
		results = append(results, output)
	}

	return results
}

func expandServiceFabricClusterClientCertificateThumbprints(input []interface{}) *[]servicefabric.ClientCertificateThumbprint {
	results := make([]servicefabric.ClientCertificateThumbprint, 0)

	for _, v := range input {
		val := v.(map[string]interface{})

		thumbprint := val["thumbprint"].(string)
		isAdmin := val["is_admin"].(bool)

		result := servicefabric.ClientCertificateThumbprint{
			CertificateThumbprint: utils.String(thumbprint),
			IsAdmin:               utils.Bool(isAdmin),
		}
		results = append(results, result)
	}

	return &results
}

func flattenServiceFabricClusterClientCertificateThumbprints(input *[]servicefabric.ClientCertificateThumbprint) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, v := range *input {
		result := make(map[string]interface{})

		if thumbprint := v.CertificateThumbprint; thumbprint != nil {
			result["thumbprint"] = *thumbprint
		}

		if isAdmin := v.IsAdmin; isAdmin != nil {
			result["is_admin"] = *isAdmin
		}

		results = append(results, result)
	}

	return results
}

func expandServiceFabricClusterClientCertificateCommonNames(input []interface{}) *[]servicefabric.ClientCertificateCommonName {
	results := make([]servicefabric.ClientCertificateCommonName, 0)

	for _, v := range input {
		val := v.(map[string]interface{})

		certificate_common_name := val["common_name"].(string)
		certificate_issuer_thumbprint := val["issuer_thumbprint"].(string)
		isAdmin := val["is_admin"].(bool)

		result := servicefabric.ClientCertificateCommonName{
			CertificateCommonName:       utils.String(certificate_common_name),
			CertificateIssuerThumbprint: utils.String(certificate_issuer_thumbprint),
			IsAdmin:                     utils.Bool(isAdmin),
		}
		results = append(results, result)
	}

	return &results
}

func flattenServiceFabricClusterClientCertificateCommonNames(input *[]servicefabric.ClientCertificateCommonName) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, v := range *input {
		result := make(map[string]interface{})

		if certificate_common_name := v.CertificateCommonName; certificate_common_name != nil {
			result["common_name"] = *certificate_common_name
		}

		if certificate_issuer_thumbprint := v.CertificateIssuerThumbprint; certificate_issuer_thumbprint != nil {
			result["issuer_thumbprint"] = *certificate_issuer_thumbprint
		}

		if isAdmin := v.IsAdmin; isAdmin != nil {
			result["is_admin"] = *isAdmin
		}
		results = append(results, result)
	}

	return results
}

func expandServiceFabricClusterDiagnosticsConfig(input []interface{}) *servicefabric.DiagnosticsStorageAccountConfig {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	storageAccountName := v["storage_account_name"].(string)
	protectedAccountKeyName := v["protected_account_key_name"].(string)
	blobEndpoint := v["blob_endpoint"].(string)
	queueEndpoint := v["queue_endpoint"].(string)
	tableEndpoint := v["table_endpoint"].(string)

	config := servicefabric.DiagnosticsStorageAccountConfig{
		StorageAccountName:      utils.String(storageAccountName),
		ProtectedAccountKeyName: utils.String(protectedAccountKeyName),
		BlobEndpoint:            utils.String(blobEndpoint),
		QueueEndpoint:           utils.String(queueEndpoint),
		TableEndpoint:           utils.String(tableEndpoint),
	}
	return &config
}

func flattenServiceFabricClusterDiagnosticsConfig(input *servicefabric.DiagnosticsStorageAccountConfig) []interface{} {
	results := make([]interface{}, 0)

	if v := input; v != nil {
		output := make(map[string]interface{})

		if name := v.StorageAccountName; name != nil {
			output["storage_account_name"] = *name
		}

		if name := v.ProtectedAccountKeyName; name != nil {
			output["protected_account_key_name"] = *name
		}

		if endpoint := v.BlobEndpoint; endpoint != nil {
			output["blob_endpoint"] = *endpoint
		}

		if endpoint := v.QueueEndpoint; endpoint != nil {
			output["queue_endpoint"] = *endpoint
		}

		if endpoint := v.TableEndpoint; endpoint != nil {
			output["table_endpoint"] = *endpoint
		}

		results = append(results, output)
	}

	return results
}

func expandServiceFabricClusterUpgradePolicyDeltaHealthPolicy(input []interface{}) *servicefabric.ClusterUpgradeDeltaHealthPolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	deltaHealthPolicy := &servicefabric.ClusterUpgradeDeltaHealthPolicy{}
	v := input[0].(map[string]interface{})
	deltaHealthPolicy.MaxPercentDeltaUnhealthyNodes = utils.Int32(int32(v["max_delta_unhealthy_nodes_percent"].(int)))
	deltaHealthPolicy.MaxPercentUpgradeDomainDeltaUnhealthyNodes = utils.Int32(int32(v["max_upgrade_domain_delta_unhealthy_nodes_percent"].(int)))
	deltaHealthPolicy.MaxPercentDeltaUnhealthyApplications = utils.Int32(int32(v["max_delta_unhealthy_applications_percent"].(int)))

	return deltaHealthPolicy
}

func expandServiceFabricClusterUpgradePolicyHealthPolicy(input []interface{}) *servicefabric.ClusterHealthPolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	healthPolicy := &servicefabric.ClusterHealthPolicy{}
	v := input[0].(map[string]interface{})
	healthPolicy.MaxPercentUnhealthyApplications = utils.Int32(int32(v["max_unhealthy_applications_percent"].(int)))
	healthPolicy.MaxPercentUnhealthyNodes = utils.Int32(int32(v["max_unhealthy_nodes_percent"].(int)))

	return healthPolicy
}

func expandServiceFabricClusterUpgradePolicy(input []interface{}) *servicefabric.ClusterUpgradePolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	policy := &servicefabric.ClusterUpgradePolicy{}
	v := input[0].(map[string]interface{})

	policy.ForceRestart = utils.Bool(v["force_restart_enabled"].(bool))
	policy.HealthCheckStableDuration = utils.String(v["health_check_stable_duration"].(string))
	policy.UpgradeDomainTimeout = utils.String(v["upgrade_domain_timeout"].(string))
	policy.UpgradeReplicaSetCheckTimeout = utils.String(v["upgrade_replica_set_check_timeout"].(string))
	policy.UpgradeTimeout = utils.String(v["upgrade_timeout"].(string))
	policy.HealthCheckRetryTimeout = utils.String(v["health_check_retry_timeout"].(string))
	policy.HealthCheckWaitDuration = utils.String(v["health_check_wait_duration"].(string))

	if v["health_policy"] != nil {
		policy.HealthPolicy = expandServiceFabricClusterUpgradePolicyHealthPolicy(v["health_policy"].([]interface{}))
	}
	if v["delta_health_policy"] != nil {
		policy.DeltaHealthPolicy = expandServiceFabricClusterUpgradePolicyDeltaHealthPolicy(v["delta_health_policy"].([]interface{}))
	}

	return policy
}

func flattenServiceFabricClusterUpgradePolicy(input *servicefabric.ClusterUpgradePolicy) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make(map[string]interface{})

	if forceRestart := input.ForceRestart; forceRestart != nil {
		output["force_restart_enabled"] = *forceRestart
	}

	if healthCheckRetryTimeout := input.HealthCheckRetryTimeout; healthCheckRetryTimeout != nil {
		output["health_check_retry_timeout"] = *healthCheckRetryTimeout
	}

	if healthCheckStableDuration := input.HealthCheckStableDuration; healthCheckStableDuration != nil {
		output["health_check_stable_duration"] = *healthCheckStableDuration
	}

	if healthCheckWaitDuration := input.HealthCheckWaitDuration; healthCheckWaitDuration != nil {
		output["health_check_wait_duration"] = *healthCheckWaitDuration
	}

	if upgradeDomainTimeout := input.UpgradeDomainTimeout; upgradeDomainTimeout != nil {
		output["upgrade_domain_timeout"] = *upgradeDomainTimeout
	}

	if upgradeReplicaSetCheckTimeout := input.UpgradeReplicaSetCheckTimeout; upgradeReplicaSetCheckTimeout != nil {
		output["upgrade_replica_set_check_timeout"] = *upgradeReplicaSetCheckTimeout
	}

	if upgradeTimeout := input.UpgradeTimeout; upgradeTimeout != nil {
		output["upgrade_timeout"] = *upgradeTimeout
	}

	output["health_policy"] = flattenServiceFabricClusterUpgradePolicyHealthPolicy(input.HealthPolicy)
	output["delta_health_policy"] = flattenServiceFabricClusterUpgradePolicyDeltaHealthPolicy(input.DeltaHealthPolicy)

	return []interface{}{output}
}

func flattenServiceFabricClusterUpgradePolicyHealthPolicy(input *servicefabric.ClusterHealthPolicy) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make(map[string]interface{})

	if input.MaxPercentUnhealthyApplications != nil {
		output["max_unhealthy_applications_percent"] = *input.MaxPercentUnhealthyApplications
	}

	if input.MaxPercentUnhealthyNodes != nil {
		output["max_unhealthy_nodes_percent"] = *input.MaxPercentUnhealthyNodes
	}

	return []interface{}{output}
}

func flattenServiceFabricClusterUpgradePolicyDeltaHealthPolicy(input *servicefabric.ClusterUpgradeDeltaHealthPolicy) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	output := make(map[string]interface{})

	if input.MaxPercentDeltaUnhealthyApplications != nil {
		output["max_delta_unhealthy_applications_percent"] = input.MaxPercentDeltaUnhealthyApplications
	}

	if input.MaxPercentDeltaUnhealthyNodes != nil {
		output["max_delta_unhealthy_nodes_percent"] = input.MaxPercentDeltaUnhealthyNodes
	}

	if input.MaxPercentUpgradeDomainDeltaUnhealthyNodes != nil {
		output["max_upgrade_domain_delta_unhealthy_nodes_percent"] = input.MaxPercentUpgradeDomainDeltaUnhealthyNodes
	}

	return []interface{}{output}
}

func expandServiceFabricClusterFabricSettings(input []interface{}) *[]servicefabric.SettingsSectionDescription {
	results := make([]servicefabric.SettingsSectionDescription, 0)

	for _, v := range input {
		val := v.(map[string]interface{})

		name := val["name"].(string)
		params := make([]servicefabric.SettingsParameterDescription, 0)
		paramsRaw := val["parameters"].(map[string]interface{})
		for k, v := range paramsRaw {
			param := servicefabric.SettingsParameterDescription{
				Name:  utils.String(k),
				Value: utils.String(v.(string)),
			}
			params = append(params, param)
		}

		result := servicefabric.SettingsSectionDescription{
			Name:       utils.String(name),
			Parameters: &params,
		}
		results = append(results, result)
	}

	return &results
}

func flattenServiceFabricClusterFabricSettings(input *[]servicefabric.SettingsSectionDescription) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, v := range *input {
		result := make(map[string]interface{})

		if name := v.Name; name != nil {
			result["name"] = *name
		}

		parameters := make(map[string]interface{})
		if paramsRaw := v.Parameters; paramsRaw != nil {
			for _, p := range *paramsRaw {
				if p.Name == nil || p.Value == nil {
					continue
				}

				parameters[*p.Name] = *p.Value
			}
		}
		result["parameters"] = parameters
		results = append(results, result)
	}

	return results
}

func expandServiceFabricClusterNodeTypes(input []interface{}) *[]servicefabric.NodeTypeDescription {
	results := make([]servicefabric.NodeTypeDescription, 0)

	for _, v := range input {
		node := v.(map[string]interface{})

		name := node["name"].(string)
		instanceCount := node["instance_count"].(int)
		clientEndpointPort := node["client_endpoint_port"].(int)
		httpEndpointPort := node["http_endpoint_port"].(int)
		isPrimary := node["is_primary"].(bool)
		durabilityLevel := node["durability_level"].(string)

		result := servicefabric.NodeTypeDescription{
			Name:                         utils.String(name),
			VMInstanceCount:              utils.Int32(int32(instanceCount)),
			IsPrimary:                    utils.Bool(isPrimary),
			ClientConnectionEndpointPort: utils.Int32(int32(clientEndpointPort)),
			HTTPGatewayEndpointPort:      utils.Int32(int32(httpEndpointPort)),
			DurabilityLevel:              servicefabric.DurabilityLevel(durabilityLevel),
		}

		if isStateless, ok := node["is_stateless"]; ok {
			result.IsStateless = utils.Bool(isStateless.(bool))
		}

		if multipleAvailabilityZones, ok := node["multiple_availability_zones"]; ok {
			result.MultipleAvailabilityZones = utils.Bool(multipleAvailabilityZones.(bool))
		}

		if props, ok := node["placement_properties"]; ok {
			placementProperties := make(map[string]*string)
			for key, value := range props.(map[string]interface{}) {
				placementProperties[key] = utils.String(value.(string))
			}

			result.PlacementProperties = placementProperties
		}

		if caps, ok := node["capacities"]; ok {
			capacities := make(map[string]*string)
			for key, value := range caps.(map[string]interface{}) {
				capacities[key] = utils.String(value.(string))
			}

			result.Capacities = capacities
		}

		if v := int32(node["reverse_proxy_endpoint_port"].(int)); v != 0 {
			result.ReverseProxyEndpointPort = utils.Int32(v)
		}

		applicationPortsRaw := node["application_ports"].([]interface{})
		if len(applicationPortsRaw) > 0 {
			portsRaw := applicationPortsRaw[0].(map[string]interface{})

			startPort := portsRaw["start_port"].(int)
			endPort := portsRaw["end_port"].(int)

			result.ApplicationPorts = &servicefabric.EndpointRangeDescription{
				StartPort: utils.Int32(int32(startPort)),
				EndPort:   utils.Int32(int32(endPort)),
			}
		}

		ephemeralPortsRaw := node["ephemeral_ports"].([]interface{})
		if len(ephemeralPortsRaw) > 0 {
			portsRaw := ephemeralPortsRaw[0].(map[string]interface{})

			startPort := portsRaw["start_port"].(int)
			endPort := portsRaw["end_port"].(int)

			result.EphemeralPorts = &servicefabric.EndpointRangeDescription{
				StartPort: utils.Int32(int32(startPort)),
				EndPort:   utils.Int32(int32(endPort)),
			}
		}

		results = append(results, result)
	}

	return &results
}

func flattenServiceFabricClusterNodeTypes(input *[]servicefabric.NodeTypeDescription) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, v := range *input {
		output := make(map[string]interface{})

		if name := v.Name; name != nil {
			output["name"] = *name
		}

		if placementProperties := v.PlacementProperties; placementProperties != nil {
			output["placement_properties"] = placementProperties
		}

		if capacities := v.Capacities; capacities != nil {
			output["capacities"] = capacities
		}

		if count := v.VMInstanceCount; count != nil {
			output["instance_count"] = int(*count)
		}

		if primary := v.IsPrimary; primary != nil {
			output["is_primary"] = *primary
		}

		if port := v.ClientConnectionEndpointPort; port != nil {
			output["client_endpoint_port"] = *port
		}

		if port := v.HTTPGatewayEndpointPort; port != nil {
			output["http_endpoint_port"] = *port
		}

		if port := v.ReverseProxyEndpointPort; port != nil {
			output["reverse_proxy_endpoint_port"] = *port
		}

		if isStateless := v.IsStateless; isStateless != nil {
			output["is_stateless"] = *isStateless
		}

		if multipleAvailabilityZones := v.MultipleAvailabilityZones; multipleAvailabilityZones != nil {
			output["multiple_availability_zones"] = *multipleAvailabilityZones
		}

		output["durability_level"] = string(v.DurabilityLevel)

		applicationPorts := make([]interface{}, 0)
		if ports := v.ApplicationPorts; ports != nil {
			r := make(map[string]interface{})
			if start := ports.StartPort; start != nil {
				r["start_port"] = int(*start)
			}
			if end := ports.EndPort; end != nil {
				r["end_port"] = int(*end)
			}
			applicationPorts = append(applicationPorts, r)
		}
		output["application_ports"] = applicationPorts

		ephemeralPorts := make([]interface{}, 0)
		if ports := v.EphemeralPorts; ports != nil {
			r := make(map[string]interface{})
			if start := ports.StartPort; start != nil {
				r["start_port"] = int(*start)
			}
			if end := ports.EndPort; end != nil {
				r["end_port"] = int(*end)
			}
			ephemeralPorts = append(ephemeralPorts, r)
		}
		output["ephemeral_ports"] = ephemeralPorts

		results = append(results, output)
	}

	return results
}
