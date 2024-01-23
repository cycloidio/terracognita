## [Unreleased]

## Added
- Azurerm added new resource: `azurerm_network_interface_security_group_association`
  ([Issue #389](https://github.com/cycloidio/terracognita/issues/389))

### Fixed
- The generated HCL now has the fixed version for the provider used instead of using the latest one by default
  ([Issue #378](https://github.com/cycloidio/terracognita/issues/378))
- Add resource_group scope to azurerm_storage_account
  ([Issue #393](https://github.com/cycloidio/terracognita/issues/393))
- Make it work with AzureRM government credentials
  ([Issue #405](https://github.com/cycloidio/terracognita/issues/405))

## [0.8.4] _2023-05-18_

## Added
- Azurerm added new resources: `azurerm_recovery_services_vault`, `azurerm_backup_policy_vm`, `azurerm_backup_protected_vm`, `azurerm_data_protection_backup_instance_disk`, ` azurerm_data_protection_backup_policy_disk`
  ([Issue #383](https://github.com/cycloidio/terracognita/issues/383))
- Azurerm now can use the `--tags` filter
  ([Issue #361](https://github.com/cycloidio/terracognita/issues/361))
- Ability declare module variables on map types
  ([Issue #365](https://github.com/cycloidio/terracognita/issues/365))
- Support for `ExactlyOneOf` configuration for the schema so the generated HCL is correct
  ([Issue #340](https://github.com/cycloidio/terracognita/issues/340))

### Changed
- Update azurerm template for irregular cases in List method arguments order
  ([Issue #383](https://github.com/cycloidio/terracognita/issues/383))
- Validation for specific provider tags to ignore is done now on the Provider implementation
  ([Issue #358](https://github.com/cycloidio/terracognita/issues/358))
- Added new Azure resource: `azurerm_api_management`
  ([PR#363](https://github.com/cycloidio/terracognita/pull/363))
- Added new Azure resource: `azurerm_backup_policy_vm_workload`
  ([PR#386](https://github.com/cycloidio/terracognita/pull/386))

### Fixed

- Nested HCL Maps now are written correctly
  ([Issue #337](https://github.com/cycloidio/terracognita/issues/337))
- Cyclic dependencies between resources now it's fixed
  ([Issue #379](https://github.com/cycloidio/terracognita/issues/379))
- `aws_db_subnet_group` that have `name: "default"` are now ignored as they are managed by AWS
  ([Issue #376](https://github.com/cycloidio/terracognita/issues/376))
- `aws_alb_listener_rule` and `aws_lb_listener_rule` that have `priority: 99999` are now ignored as they are managed by AWS
  ([Issue #375](https://github.com/cycloidio/terracognita/issues/375))

## [0.8.3] _2023-03-14_

### Fixed

- Azurerm force `create_option="Attach"` with `azurerm_virtual_machine_data_disk_attachment`
  ([PR #359](https://github.com/cycloidio/terracognita/pull/359))
- Azurerm `azurerm_network_security_group.security_rule.protocol` now has the right format
  ([PR #357](https://github.com/cycloidio/terracognita/pull/357))

### Changed

- Interpolation now will compare value lowcasing them beforehand
  ([PR #356](https://github.com/cycloidio/terracognita/pull/356))

## [0.8.2] _2023-03-07_

### Added

- Added new Azure resource: `azurerm_virtual_machine_data_disk_attachment`
  ([PR#334](https://github.com/cycloidio/terracognita/pull/334))
- Allow the writing of tf/provider into a separated config key
  ([PR#319](https://github.com/cycloidio/terracognita/pull/319))
- Add new aws resource: aws_ecs_task_definition
  ([PR#333](https://github.com/cycloidio/terracognita/pull/333))
- Added new Azure resource: `azurerm_data_protection_backup_vault`
  ([PR#349](https://github.com/cycloidio/terracognita/pull/349))

### Changed

- Azure: Set a valide `admin_password` with `azurerm_windows_virtual_machine`
  ([Issue #352](https://github.com/cycloidio/terracognita/issues/352))
- Azure: azure: do not define external disk with `azurerm_virtual_machine`
  ([PR #336](https://github.com/cycloidio/terracognita/pull/336))
- Improved the way resource references/interpolations work, now it's more deterministic
  ([Issue #346](https://github.com/cycloidio/terracognita/issues/346))


### Fixed

- Tags are being used again for filtering when importing
  ([Issue #322](https://github.com/cycloidio/terracognita/issues/322))
- Remove duplicate names from the reader cache on AzureRM
  ([Issue #341](https://github.com/cycloidio/terracognita/issues/341))
- Added a new Provider function to let the Provider fix resources content before writing it and fixed some AzureRM resources with it
  ([Issue #322](https://github.com/cycloidio/terracognita/issues/322))
- ModueleVariables now works with nested fields of array blocks
  ([Issue #344](https://github.com/cycloidio/terracognita/issues/344))

## [0.8.1] _2022-08-10_

### Fixed

- Bug with tfdocs, updated it to the latest version
  ([PR #317](https://github.com/cycloidio/terracognita/pull/317))

## [0.8.0] _2022-08-09_

### Fixed

- Repetitive blocks now have the proper variables within them
  ([Issue #285](https://github.com/cycloidio/terracognita/issues/285))
- Script to update providers now pushed the tag to the fork
  ([PR #294](https://github.com/cycloidio/terracognita/pull/294))
- `aws_lb_target_group_attachment` was raising a nil pointer exception
  ([Issue #297](https://github.com/cycloidio/terracognita/issues/297))
-  fix resource name in `azurerm_dns_aaaa_record` and `azurerm_mssql_elastic_pool` and also fix caching of resources
  ([Issue #303](https://github.com/cycloidio/terracognita/issues/303))
  ([Issue #305](https://github.com/cycloidio/terracognita/issues/305))


### Added
- Updated Contribute GCP section, added template modularity and added new GCP resources: `google_dns_policy`, `google_billing_subaccount`, `google_sql_database`,  `google_compute_address`, ` google_compute_attached_disk`, `google_compute_autoscaler `,  `google_compute_global_address`, `google_compute_image`, `google_compute_instance_group_manager`, `google_compute_instance_template`, `google_compute_managed_ssl_certificate`, `google_compute_network_endpoint_group`, `google_compute_route`, `google_compute_security_policy`, `google_compute_service_attachment`, `google_compute_snapshot`, `google_compute_ssl_policy`, `google_compute_subnetwork`closes issue #188, `google_compute_target_grpc_proxy`, `google_compute_target_instance`, `google_compute_target_pool`, `google_compute_target_ssl_proxy`, `google_compute_target_tcp_proxy`, `google_filestore_instance`, `google_container_cluster`,`google_container_node_pool`,`google_redis_instance`,`google_logging_metric`,`google_monitoring_alert_policy`,`google_monitoring_group`, `google_monitoring_notification_channel`,`google_monitoring_uptime_check_config`
  ([Issue #188](https://github.com/cycloidio/terracognita/issues/188))
  ([Issue #273](https://github.com/cycloidio/terracognita/issues/273))
- Support for vSphere provider
  ([Issue #296](https://github.com/cycloidio/terracognita/issues/296))


- Added new AWS resources: `aws_ec2_transit_gateway_peering_attachment`, `aws_ec2_transit_gateway_peering_attachment_accepter`, `aws_ec2_transit_gateway_prefix_list_reference`, `aws_ec2_transit_gateway_route`, `aws_ec2_transit_gateway_route_table_association`, `aws_ec2_transit_gateway_route_table_propagation`, `aws_ec2_transit_gateway_vpc_attachment_accepter`
  ([Issue #299](https://github.com/cycloidio/terracognita/issues/299))
- Update tfdocs version to v0.0.0-20220809093344-d999d1c2069e and added app service azurerm resources: `azurerm_linux_web_app`, `azurerm_linux_web_app_slot`, `azurerm_service_plan`, `azurerm_source_control_token`, `azurerm_static_site`, `azurerm_static_site_custom_domain`, `azurerm_web_app_active_slot`, `azurerm_web_app_hybrid_connection`, `azurerm_windows_web_app`, `azurerm_windows_web_app_slot`
  ([PR #314](https://github.com/cycloidio/terracognita/pull/314))

## [0.7.6] _2022-05-11_

### Changed

- Module no longer has variables commented in it
  ([Issue #290](https://github.com/cycloidio/terracognita/issues/290))

## [0.7.5] _2022-05-10_

### Added

- aws resources: `aws_vpc_endpoint`
  ([Issue #254](https://github.com/cycloidio/terracognita/issues/254))
- New flag `--hcl-provider-block` to be able to opt out of the `provider "" {}` on HCL
  ([Issue #250](https://github.com/cycloidio/terracognita/issues/250))
- Azure resources : `azurerm_managed_disk`, `azurerm_virtual_machine_scale_set_extension`, `azurerm_linux_virtual_machine`,
`azurerm_linux_virtual_machine`, `azurerm_linux_virtual_machine_scale_set`, `azurerm_windows_virtual_machine`, `azurerm_windows_virtual_machine_scale_set`, `azurerm_kubernetes_cluster`, `azurerm_kubernetes_cluster_node_pool`, `azurerm_network_interface`, `azurerm_virtual_hub`, `azurerm_virtual_hub_bgp_connection`, `azurerm_virtual_hub_connection`, `azurerm_virtual_hub_ip`, `azurerm_virtual_hub_route_table`, `azurerm_virtual_hub_security_partner_provider`,`azurerm_mssql_database`, `azurerm_mssql_elasticpool`,`azurerm_mssql_firewall_rule`,`azurerm_mssql_server`,`azurerm_mssql_server_security_alert_policy`,`azurerm_mssql_server_vulnerability_assessment`,`azurerm_mssql_virtual_machine`,`azurerm_mssql_virtual_network_rule`,`azurerm_redis_cache`,`azurerm_redis_firewall_rule`,`azurerm_dns_zone`,`azurerm_dns_a_record`,`azurerm_dns_aaaa_record`,`azurerm_dns_caa_record`,`azurerm_dns_cname_record`,`azurerm_dns_mx_record`,`azurerm_dns_ns_record`,`azurerm_dns_ptr_record`,`azurerm_dns_srv_record`,`azurerm_dns_txt_record`,`azurerm_private_dns_zone`,`azurerm_private_dns_zone_virtual_network_link`,`azurerm_private_dns_a_record`,`azurerm_private_dns_aaaa_record`,`azurerm_private_dns_cname_record`,`azurerm_private_dns_mx_record`,`azurerm_private_dns_ptr_record`,`azurerm_private_dns_srv_record`,`azurerm_private_dns_txt_record`,`azurerm_lb`,`azurerm_lb_backend_address_pool`,`azurerm_lb_rule `,`azurerm_lb_outbound_rule`,`azurerm_lb_nat_rule`,`azurerm_lb_nat_pool`,`azurerm_lb_probe`,`azurerm_policy_remediation`,`azurerm_policy_set_definition`,`azurerm_key_vault`,`azurerm_key_vault_access_policy`,`azurerm_application_insights`,`azurerm_application_insights_api_key`,`azurerm_application_insights_analytics_item`,`azurerm_application_insights_web_test`,`azurerm_log_analytics_workspace`,`azurerm_log_analytics_linked_service`,`azurerm_log_analytics_datasource_windows_performance_counter`,`azurerm_log_analytics_datasource_windows_event`,`azurerm_monitor_action_group`,`azurerm_monitor_activity_log_alert`,`azurerm_monitor_autoscale_setting`,`azurerm_monitor_log_profile`,`azurerm_monitor_metric_alert`
  ([Issue #100](https://github.com/cycloidio/terracognita/issues/100))
- Added new AWS resources: `aws_route_table`, `aws_ec2_transit_gateway`, `aws_ec2_transit_gateway_vpc_attachment`,`aws_ec2_transit_gateway_route_table`, `aws_ec2_transit_gateway_multicast_domain`
  ([Issue #286](https://github.com/cycloidio/terracognita/issues/286))

### Changed

- Update terraform from v0.13.5 to v1.1.9
  ([PR #284](https://github.com/cycloidio/terracognita/pull/264))
- Update terraform-provider-google from v3.67.0 to v4.9.0
  ([PR #263](https://github.com/cycloidio/terracognita/pull/264))
- Update terraform-provider-aws from v3.40.0 to v4.9.0
  ([PR #263](https://github.com/cycloidio/terracognita/pull/264))
  ([PR #284](https://github.com/cycloidio/terracognita/pull/264))
- Update terraform-provider-azurerm from v1.44.0 to v3.3.0
  ([PR #263](https://github.com/cycloidio/terracognita/pull/263))
  ([PR #284](https://github.com/cycloidio/terracognita/pull/264))
- AzureRM now supports multiple Resource Group Names
  ([Issue #266](https://github.com/cycloidio/terracognita/issues/266))
- Azure API resources update to latest version, fix caching issue in resources.go, add modularity to template for irregular cases, update Azurerm contribute readme, removed Azurerm 3.0 deprecated resources:`azurerm_virtual_machine_scale_set`,`azurerm_sql_server`, `azurerm_sql_database`, `azurerm_sql_elasticpool`, `azurerm_sql_firewall_rule`, `azurerm_sql_server` and removed temporatily support for `azurerm_web_application_firewall_policy` due to json issue reported on sdk
  ([Issue #100](https://github.com/cycloidio/terracognita/issues/100))
- Update tfdocs version to v0.0.0-20220509071309-2f31fd03120a
  ([Issue #286](https://github.com/cycloidio/terracognita/issues/286))

### Fixed

- Issue with importing `google_storage_bucket_iam_policy`
  ([Issue #258](https://github.com/cycloidio/terracognita/issues/258))
- Removed default region used on AWS initialization now uses the one specified by the user
  ([Issue #253](https://github.com/cycloidio/terracognita/issues/253))
- HCL provider generation now users the Defaults instead of setting empty values
  ([Issue #268](https://github.com/cycloidio/terracognita/issues/268))

## [0.7.4] _2021-09-23_

### Changed

- Updated mxwriter to v1.0.4 that fixes an internal bug
  ([PR #236](https://github.com/cycloidio/terracognita/issues/236))

### Fixed

- Fix usage of `yaml` extension with `--module-variables`
  ([PR #257](https://github.com/cycloidio/terracognita/pull/257))
- Do not stop when service not enable with Azure and Google
  ([Issue #247](https://github.com/cycloidio/terracognita/issues/247))
- Google resources that do not support Labels are no longer filtered by it
  ([PR #236](https://github.com/cycloidio/terracognita/issues/237))
- Fix error to enable support for azurerm_resource_group
  ([Issue #232](https://github.com/cycloidio/terracognita/issues/232))
- TFState when generating modules now has the module on it too
  ([Issue #240](https://github.com/cycloidio/terracognita/issues/240))


## [0.7.3] _2021-08-30_

### Added

- aws resources: `aws_autoscaling_schedule`
  ([PR #194](https://github.com/cycloidio/terracognita/pull/224))

### Changed

- Support filter with `tag` attribute used by AWS on `aws_autoscaling_group` resource
  ([Issue #223](https://github.com/cycloidio/terracognita/issues/223))
- Resource name improved by low casing all the options
  ([Issue #225](https://github.com/cycloidio/terracognita/issues/225))


### Fixed

- Integration between Interpolation and Modules has been changed to not generate invalid HCL references
  ([Issue #219](https://github.com/cycloidio/terracognita/issues/219)
- Variables names will now be normalized to be valid HCL
  ([PR #219](https://github.com/cycloidio/terracognita/pull/227)

## [0.7.2] _2021-08-13_

### Added

  ([PR #194](https://github.com/cycloidio/terracognita/pull/224))
- azure resources (compute): `azurerm_availability_set`,`azurerm_image`, `azurerm_container_registry`, `azurerm_container_registry_webhook`, `azurerm_application_gateway`,	`azurerm_application_security_group`, `azurerm_network_ddos_protection_plan`, `azurerm_firewall`, `azurerm_local_network_gateway` , `azurerm_nat_gateway`, `azurerm_network_profile`, `azurerm_network_security_rule`, `azurerm_public_ip`, `azurerm_public_ip_prefix`, `azurerm_route`, `azurerm_route_table`, `azurerm_virtual_network_gateway`, `azurerm_virtual_network_gateway_connection`, `azurerm_virtual_network_peering`, `azurerm_web_application_firewall_policy`, `azurerm_storage_account`, `azurerm_storage_blob`, `azurerm_storage_queue`, `azurerm_storage_share`, `azurerm_storage_table`, `azurerm_mariadb_configuration`, `azurerm_mariadb_database`, `azurerm_mariadb_firewall_rule`, `azurerm_mariadb_server`, `azurerm_mariadb_virtual_network_rule`,  `azurerm_mysql_configuration`, `azurerm_mysql_database`, `azurerm_mysql_firewall_rule`, `azurerm_mysql_server`, `azurerm_mysql_virtual_network_rule`, `azurerm_postgresql_configuration`, `azurerm_postgresql_database`, `azurerm_postgresql_firewall_rule`, `azurerm_postgresql_server`, `azurerm_postgresql_virtual_network_rule`, `azurerm_sql_database`, `azurerm_sql_elasticpool`, `azurerm_sql_firewall_rule`,`azurerm_sql_server`
  ([Issue #100](https://github.com/cycloidio/terracognita/issues/100)
- `provider` and `terraform` blocks to the HCL generation
  ([Issue #136](https://github.com/cycloidio/terracognita/issues/136))


### Changed

- Update aws resources regarding missing pagination and filter
  ([PR #202](https://github.com/cycloidio/terracognita/pull/202))
- Resource names now are generated removing invalid characters instead just assigning a random alphanumeric value
  ([Issue #208](https://github.com/cycloidio/terracognita/issues/208))

### Fixed

- Import with `aws_alb_target_group_attachment` now validates if the needed values are present
  ([Issue #213](https://github.com/cycloidio/terracognita/issues/213))

## [0.7.1] _2021-07-15_

### Added

- aws resources: `aws_eip`, `aws_dynamodb_global_table`, `aws_dynamodb_table`, `aws_ecs_cluster`, `aws_ecs_service`, `aws_athena_workgroup`, `aws_glue_catalog_database`, `aws_glue_catalog_table`, `aws_fsx_lustre_file_system`, `aws_batch_job_definition`, `aws_dax_cluster`, `aws_directory_service_directory`, `aws_dms_replication_instance`, `aws_dx_gateway`, `aws_efs_file_system`, `aws_eks_cluster`, `aws_elasticache_replication_group`, `aws_elastic_beanstalk_application`, `aws_emr_cluster`, `aws_internet_gateway`, `aws_kinesis_stream`, `aws_lightsail_instance`, `aws_media_store_container`, `aws_mq_broker`, `aws_nat_gateway`, `aws_neptune_cluster`, `aws_rds_cluster`, `aws_rds_global_cluster`, `aws_redshift_cluster`, `aws_sqs_queue`, `aws_storagegateway_gateway`, `aws_vpn_gateway`.
  ([PR #194](https://github.com/cycloidio/terracognita/pull/194))
- Extra validations for the resource names
  ([PR #201](https://github.com/cycloidio/terracognita/pull/201))

### Changed

- Updated `tfdocs` to have missing resources that where causing import errors
  ([Issue #199](https://github.com/cycloidio/terracognita/issues/199))

### Fixed

- Skip aws `RequestError` errors generaly caused by service not available in a region
  ([Issue #171](https://github.com/cycloidio/terracognita/issues/171))
- Module source now is prefixed with `./` as expected
  ([PR #209](https://github.com/cycloidio/terracognita/pull/209))

## [0.7.0] _2021-07-02_

### Fixed

- `tc_category` no longer added to the generated HCL
  ([PR #187](https://github.com/cycloidio/terracognita/pull/187))
- Skip resources that are not Importable from the Provider
  ([PR #191](https://github.com/cycloidio/terracognita/pull/191))

### Changed

- Migrate all the logic to use terraform-plugin-sdk/v2
  ([Issue #151](https://github.com/cycloidio/terracognita/issues/151))

## [0.6.4] _2021-04-29_

### Changed

- AWS error handling from Message to Code and added 'AccessDeniedException'
  ([Issue #171](https://github.com/cycloidio/terracognita/issues/171))

### Fixed

- `--labels` flag is correctly read now on Google CMD
  ([PR #180](https://github.com/cycloidio/terracognita/pull/180))

## [0.6.3] _2021-03-30_

### Fixed

- Empty array values on modules now are generated correctly and not failing
  ([PR #174](https://github.com/cycloidio/terracognita/pull/174))

## [0.6.2] _2021-03-18_

We had an error on the Pipeline of the last release so we made a quick patch release to fix it

## [0.6.1] _2021-03-12_

### Added

- Ability to create Modules directly when importing
  ([Issue #141](https://github.com/cycloidio/terracognita/issues/141))

## [0.6.0] _2020-12-22_

### Added

- state dependencies between resources using `dependencies`
  ([PR #131](https://github.com/cycloidio/terracognita/pull/131))
- aws resource: `aws_alb_listener_certificate`, `aws_lb_cookie_stickiness_policy`, `aws_lb_target_group_attachment`, `aws_volume_attachment`, `aws_elasticsearch_domain`, `aws_elasticsearch_domain_policy`, `aws_lambda_function`, `aws_api_gateway_rest_api`, `aws_api_gateway_deployment`, `aws_api_gateway_stage`, `aws_api_gateway_resource`.
  ([PR #128](https://github.com/cycloidio/terracognita/pull/128))
- cli option to deactivate interpolation
  ([PR #133](https://github.com/cycloidio/terracognita/pull/133))
- AWS support for profile/config
  ([Issue #48](https://github.com/cycloidio/terracognita/issues/48))
- Azure Virtual Desktop resources
  ([PR #145](https://github.com/cycloidio/terracognita/pull/145))
- Log File to always write the last -v logs to
  ([Issue #149](https://github.com/cycloidio/terracognita/issues/149))
- Authentication using AWS session token
  ([Issue #154](https://github.com/cycloidio/terracognita/issues/154))
- Support for Homebrew
  ([Issue #153](https://github.com/cycloidio/terracognita/issues/153))

### Changed

- HCL lib version from V1 to V2 and all the implications
  ([PR #135](https://github.com/cycloidio/terracognita/pull/135))
- All the Provider and Terraform versions
  ([PR #143](https://github.com/cycloidio/terracognita/pull/143))

### Fixed

- Crashing import by adding an error handling on provider errors
  ([Issue #138](https://github.com/cycloidio/terracognita/issues/138))
- No more issues for HCL2 when generated
  ([PR #148](https://github.com/cycloidio/terracognita/pull/148))

## [0.5.1] _2020-07-17_

### Fixed

- Error with the Resource name always being the alphanumeric instead of the Tag Name
  ([PR #124](https://github.com/cycloidio/terracognita/pull/124))
- Pagination and nil pointer errors
  ([PR #123](https://github.com/cycloidio/terracognita/pull/123))
- Error with mutual interpolation between resources
  ([PR #125](https://github.com/cycloidio/terracognita/pull/125))

## [0.5.0] _2020-06-19_

### Added

- provider resource: implement SetImporter to set schema.Resource.Importer when resource is not importable.
  ([PR #116](https://github.com/cycloidio/terracognita/pull/116))
- aws resource: `aws_iam_group_membership`
  ([PR #116](https://github.com/cycloidio/terracognita/pull/116))
- google resources: `google_compute_backend_bucket`, `google_project_iam_custom_role`, `google_storage_bucket_iam_policy`, `google_compute_instance_iam_policy`
  ([PR #97](https://github.com/cycloidio/terracognita/pull/97))
- aws: `aws_lb`, `aws_lb_listener`, `aws_lb_listener_rule`, `aws_lb_target_group`
  ([PR #96](https://github.com/cycloidio/terracognita/pull/96))
- aws: Pagination of all the functions on the reader
  ([Issue #13](https://github.com/cycloidio/terracognita/issues/13))

### Changed

- aws resources: do not write group_membership if the user has no groups.
  ([issue #111](https://github.com/cycloidio/terracognita/issue/111))
- filter: update IsExcluded and add IsIncluded to verify multiple resources.
  ([PR #96](https://github.com/cycloidio/terracognita/pull/96))
- Provide filters to resource functions instead of tags only
  ([PR #92](https://github.com/cycloidio/terracognita/pull/92))
- Upgraded all the Provider and Terraform versions
  ([PR #114](https://github.com/cycloidio/terracognita/pull/114))

### Fixed

- Error when importing `aws_iam_user_group_membership` without groups
  ([Issue #104](https://github.com/cycloidio/terracognita/issues/104))
- util/retry now ignores the internal errors format
  ([Issue #106](https://github.com/cycloidio/terracognita/issues/106))

## [0.4.0] _2020-03-31_

### Added

- aws resources: `aws_db_subnet_group`, `aws_key_pair`, `aws_vpc_peering_connection`, `aws_alb_target_group`, `aws_alb_listener`, `aws_alb_listener_rule`
  ([PR #87](https://github.com/cycloidio/terracognita/pull/87))
- Terraform variable interpolation is available on string values
  ([PR #81](https://github.com/cycloidio/terracognita/pull/81))
- aws resource: `aws_db_parameter_group`, `aws_iam_access_key`, `aws_cloudwatch_metric_alarm`, `aws_autoscaling_policy`, `aws_iam_user_ssh_key`
  ([PR #78](https://github.com/cycloidio/terracognita/pull/78))
- New flag `--target` to allow specific resource+id import
  ([Issue #40](https://github.com/cycloidio/terracognita/issues/40))
- New AzureRM provider
  ([PR #88](https://github.com/cycloidio/terracognita/pull/88))

## [0.3.0] _2020-01-02_

### Added

- google resource: `ComputeDisk`, `StorageBucket` and `SqlDatabaseInstance`
  ([PR #73](https://github.com/cycloidio/terracognita/pull/73))
- google resource: `ComputeSSLCertificate`, `ComputeTargetHTTPProxy`, `ComputeTargetHTTPSProxy` and `ComputeURLMap`
  ([PR #67](https://github.com/cycloidio/terracognita/pull/67))
- google resource: `ComputeHealthCheck`, `ComputeInstanceGroup` and `ComputeBackendService`
  ([PR #64](https://github.com/cycloidio/terracognita/pull/64))
- aws resource: `aws_launch_configuration`, `aws_launch_template` and `aws_autoscaling_group`
  ([PR #68](https://github.com/cycloidio/terracognita/pull/68))
- google resource: compute instance
  ([PR #56](https://github.com/cycloidio/terracognita/pull/56))
- google resource: compute networks and compute firewalls
  ([PR #61](https://github.com/cycloidio/terracognita/pull/61))
- google reader functions are now generated from go:generate
  ([PR #65](https://github.com/cycloidio/terracognita/pull/65))

### Changed

- During import if a resource is invalid we assume it can be skipped
  ([PR #68](https://github.com/cycloidio/terracognita/pull/68))
- 'raws' lib to be an internal library instead of a dependency
  ([Issue #69](https://github.com/cycloidio/terracognita/issues/69))

### Fixed

- '--region' flag working for different subcommands
  ([PR #63](https://github.com/cycloidio/terracognita/pull/63))

## [0.2.0] _2019-10-29_

This version changes the format of the TFState to the Terraform 0.12+ [format](https://www.terraform.io/upgrade-guides/0-12.html)

### Fixed

- HCL formatter to ignore some special keys that fail on the `fmtcmd` of HCL
  ([Issue #36](https://github.com/cycloidio/terracognita/issues/36))

### Changed

- The Terraform version from 0.11 to 0.12 with all the implications (file formats) https://www.terraform.io/upgrade-guides/0-12.html
  ([PR #33](https://github.com/cycloidio/terracognita/pull/33))

## [0.1.6] _2019-07-18_

### Added

- The `version` subcommand to show the actual build version
  ([Issue #24](https://github.com/cycloidio/terracognita/issues/24))
- CI/CD pipeline
  ([Issue #31](https://github.com/cycloidio/terracognita/pull/34))
- The `-verbose` and `-debug` options
  ([Issue #17](https://github.com/cycloidio/terracognita/issues/17))

### Changed

- Update CI/CD pipeline which now also has pre-built binaries, automate github release and docker release image.
  ([Issue #31](https://github.com/cycloidio/terracognita/issues/31))
- Better management of Throttle errors from AWS
  ([PR #49](https://github.com/cycloidio/terracognita/pull/49))

### Fixed

- Error with the Import Filter not validating before Importing/Reading
  ([PR #22](https://github.com/cycloidio/terracognita/pull/22))
- Update to version 1.0.1 of `raws` to fix panic on importing `aws_s3_bucket`
  ([Issue #29](https://github.com/cycloidio/terracognita/issues/29))
- Vendor issue with AWS TF provider and updated it to 2.31.0
  ([PR #54](https://github.com/cycloidio/terracognita/pull/54))
