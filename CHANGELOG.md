## [Unreleased]

### Added

- google resources: `google_compute_backend_bucket`, `google_project_iam_custom_role`, `google_storage_bucket_iam_policy`, `google_compute_instance_iam_policy`
  ([PR #97](https://github.com/cycloidio/terracognita/pull/97))

### Changed

- Provide filters to resource functions instead of tags only
  ([PR #92](https://github.com/cycloidio/terracognita/pull/92))

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

## Fixed

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
