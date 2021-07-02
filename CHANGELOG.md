## [Unreleased]

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
