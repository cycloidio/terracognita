package main

var (
	// functions is the list of fuctions that will be added
	// to the AWSReader with the corresponding implementation
	functions = []Function{
		// cloudwatch
		Function{
			Entity:          "MetricAlarms",
			FnServiceEntity: "Alarms",
			Prefix:          "Describe",
			Service:         "cloudwatch",
			Documentation: `
			// GetMetricAlarms returns all cloudwatch alarms based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// ec2
		Function{
			FnAttributeList: "Reservations#Instances",
			Entity:          "Instances",
			Prefix:          "Describe",
			Service:         "ec2",
			Documentation: `
			// GetInstances returns all EC2 instances based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Vpcs",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetVpcs returns all EC2 VPCs based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "VpcPeeringConnections",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetVpcPeeringConnections returns all VpcPeeringConnections based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		Function{
			HasNotPagination: true,
			Entity:           "Images",
			Prefix:           "Describe",
			Service:          "ec2",
			Documentation: `
			// GetImages returns all EC2 AMI based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			HasNotPagination: true,
			Entity:           "Images",
			Prefix:           "Describe",
			Service:          "ec2",
			FilterByOwner:    "Owners",
			Documentation: `
			// GetOwnImages returns all EC2 AMI belonging to the Account ID based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			SingularEntity:   "KeyPairInfo",
			HasNotPagination: true,
			Entity:           "KeyPairs",
			Prefix:           "Describe",
			Service:          "ec2",
			Documentation: `
			// GetKeyPairs returns all KeyPairs based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "SecurityGroups",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetSecurityGroups returns all EC2 security groups based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Subnets",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetSubnets returns all EC2 subnets based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Volumes",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetVolumes returns all EC2 volumes based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Snapshots",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetSnapshots returns all snapshots based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:        "Snapshots",
			Prefix:        "Describe",
			Service:       "ec2",
			FilterByOwner: "OwnerIds",
			Documentation: `
			// GetOwnSnapshots returns all snapshots belonging to the Account ID based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "LaunchTemplates",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetLaunchTemplates returns all LaunchTemplate belonging to the Account ID based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// autoscaling
		Function{
			Entity:         "AutoScalingGroups",
			SingularEntity: "Group",
			Prefix:         "Describe",
			Service:        "autoscaling",
			Documentation: `
			// GetAutoScalingGroups returns all AutoScalingGroup belonging to the Account ID based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "LaunchConfigurations",
			Prefix:  "Describe",
			Service: "autoscaling",
			Documentation: `
			// GetLaunchConfigurations returns all LaunchConfiguration belonging to the Account ID based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:          "GetAutoScalingPolicies",
			Entity:          "ScalingPolicies",
			FnServiceEntity: "Policies",
			Prefix:          "Describe",
			Service:         "autoscaling",
			Documentation: `
		  // GetAutoScalingPolicies returns all AutoScalingPolicies belonging to the Account ID based on the input given.
		  // Returned values are commented in the interface doc comment block.
		  `,
		},

		// elasticache
		Function{
			FnName:                "GetElastiCacheClusters",
			Entity:                "CacheClusters",
			Prefix:                "Describe",
			Service:               "elasticache",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetElastiCacheClusters returns all Elasticache clusters based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			HasNotPagination: true,
			FnName:           "GetElastiCacheTags",
			Entity:           "TagsForResource",
			SingularEntity:   "Tag",
			FnAttributeList:  "TagList",
			Prefix:           "List",
			Service:          "elasticache",
			Documentation: `
			// GetElastiCacheTags returns a list of tags of Elasticache resources based on its ARN.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// elb
		Function{
			Entity:                     "LoadBalancers",
			SingularEntity:             "LoadBalancerDescription",
			FnAttributeList:            "LoadBalancerDescriptions",
			Prefix:                     "Describe",
			Service:                    "elb",
			FnPaginationAttribute:      "NextMarker",
			FnInputPaginationAttribute: "Marker",
			Documentation: `
			// GetLoadBalancers returns a list of ELB (v1) based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:           "GetLoadBalancersTags",
			Entity:           "Tags",
			SingularEntity:   "TagDescription",
			FnAttributeList:  "TagDescriptions",
			Prefix:           "Describe",
			HasNotPagination: true,
			Service:          "elb",
			Documentation: `
			// GetLoadBalancersTags returns a list of Tags based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// elbv2
		Function{
			FnName:                     "GetLoadBalancersV2",
			Entity:                     "LoadBalancers",
			Prefix:                     "Describe",
			Service:                    "elbv2",
			FnPaginationAttribute:      "NextMarker",
			FnInputPaginationAttribute: "Marker",
			Documentation: `
			// GetLoadBalancersV2 returns a list of ELB (v2) - also known as ALB - based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:           "GetLoadBalancersV2Tags",
			Entity:           "Tags",
			SingularEntity:   "TagDescription",
			FnAttributeList:  "TagDescriptions",
			Prefix:           "Describe",
			Service:          "elbv2",
			HasNotPagination: true,
			Documentation: `
			// GetLoadBalancersV2Tags returns a list of Tags based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:                     "GetLoadBalancersV2Listeners",
			Entity:                     "Listeners",
			Prefix:                     "Describe",
			Service:                    "elbv2",
			FnPaginationAttribute:      "NextMarker",
			FnInputPaginationAttribute: "Marker",
			Documentation: `
			// GetLoadBalancersV2Listeners returns a list of Listeners based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:                     "GetLoadBalancersV2TargetGroups",
			Entity:                     "TargetGroups",
			Prefix:                     "Describe",
			Service:                    "elbv2",
			FnPaginationAttribute:      "NextMarker",
			FnInputPaginationAttribute: "Marker",
			Documentation: `
			// GetLoadBalancersV2TargetGroups returns a list of TargetGroups based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                     "ListenerCertificates",
			SingularEntity:             "Certificate",
			FnAttributeList:            "Certificates",
			FnPaginationAttribute:      "NextMarker",
			FnInputPaginationAttribute: "Marker",
			Prefix:                     "Describe",
			Service:                    "elbv2",
			Documentation: `
			// GetListenerCertificates returns a list of ListenerCertificates based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:                     "GetLoadBalancersV2Rules",
			Entity:                     "Rules",
			FnInputPaginationAttribute: "Marker",
			FnPaginationAttribute:      "NextMarker",
			Prefix:                     "Describe",
			Service:                    "elbv2",
			Documentation: `
			// GetLoadBalancersV2Rules returns a list of Rules based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// rds
		Function{
			Entity:                "DBInstances",
			FnPaginationAttribute: "Marker",
			Prefix:                "Describe",
			Service:               "rds",
			Documentation: `
			// GetDBInstances returns all DB instances based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:           "GetDBInstancesTags",
			Entity:           "TagsForResource",
			SingularEntity:   "Tag",
			FnAttributeList:  "TagList",
			HasNotPagination: true,
			Prefix:           "List",
			Service:          "rds",
			Documentation: `
			// GetDBInstancesTags returns a list of tags from an ARN, extra filters for tags can also be provided.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "DBParameterGroups",
			FnPaginationAttribute: "Marker",
			Prefix:                "Describe",
			Service:               "rds",
			Documentation: `
			// GetDBParameterGroups returns all DB parameterGroups based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "DBSubnetGroups",
			FnPaginationAttribute: "Marker",
			Prefix:                "Describe",
			Service:               "rds",
			Documentation: `
			// GetDBSubnetGroups returns all DB DBSubnetGroups based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// s3
		Function{
			// TODO: https://github.com/cycloidio/terracognita/issues/76
			FnName:       "ListBuckets",
			Entity:       "Buckets",
			Prefix:       "List",
			Service:      "s3",
			NoGenerateFn: true,
			Documentation: `
			// ListBuckets returns all S3 buckets based on the input given and specifically
			// filtering by Location as ListBuckets does not do it by itself
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:           "GetBucketTags",
			Entity:           "BucketTagging",
			SingularEntity:   "Tag",
			FnAttributeList:  "TagSet",
			HasNotPagination: true,
			Prefix:           "Get",
			Service:          "s3",
			Documentation: `
			// GetBucketTags returns tags associated with S3 buckets based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			// TODO: https://github.com/cycloidio/terracognita/issues/76
			FnName:                "ListObjects",
			Entity:                "Objects",
			FnAttributeList:       "Contents",
			FnPaginationAttribute: "Marker",
			Prefix:                "List",
			Service:               "s3",
			Documentation: `
			// ListObjects returns a list of all S3 objects in a bucket based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:           "GetObjectsTags",
			Entity:           "ObjectTagging",
			SingularEntity:   "Tag",
			FnAttributeList:  "TagSet",
			HasNotPagination: true,
			Prefix:           "Get",
			Service:          "s3",
			Documentation: `
			// GetObjectsTags returns tags associated with S3 objects based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:          "GetRecordedResourceCounts",
			Entity:          "DiscoveredResourceCounts",
			SingularEntity:  "ResourceCount",
			FnAttributeList: "ResourceCounts",
			Prefix:          "Get",
			Service:         "configservice",
			Documentation: `
			// GetRecordedResourceCounts returns counts of the AWS resources which have
			// been recorded by AWS Config.
			// See https://docs.aws.amazon.com/config/latest/APIReference/API_GetDiscoveredResourceCounts.html
			// for more information about what to enable in your AWS account, the list of
			// supported resources, etc.
			`,
		},

		// cloudfront
		Function{
			FnName:                     "GetCloudFrontDistributions",
			Entity:                     "Distributions",
			Prefix:                     "List",
			Service:                    "cloudfront",
			SingularEntity:             "DistributionSummary",
			FnPaginationAttribute:      "DistributionList.NextMarker",
			FnInputPaginationAttribute: "Marker",
			FnAttributeList:            "DistributionList.Items",
			Documentation: `
			// GetCloudFrontDistributions returns all the CloudFront Distributions on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:                     "GetCloudFrontPublicKeys",
			Entity:                     "PublicKeys",
			SingularEntity:             "PublicKeySummary",
			FnAttributeList:            "PublicKeyList.Items",
			FnPaginationAttribute:      "PublicKeyList.NextMarker",
			FnInputPaginationAttribute: "Marker",
			Prefix:                     "List",
			Service:                    "cloudfront",
			Documentation: `
			// GetCloudFrontPublicKeys returns all the CloudFront Public Keys on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                     "CloudFrontOriginAccessIdentities",
			Prefix:                     "List",
			Service:                    "cloudfront",
			SingularEntity:             "OriginAccessIdentitySummary",
			FnAttributeList:            "CloudFrontOriginAccessIdentityList.Items",
			FnPaginationAttribute:      "CloudFrontOriginAccessIdentityList.Marker",
			FnInputPaginationAttribute: "Marker",
			Documentation: `
			// GetCloudFrontOriginAccessIdentities returns all the CloudFront Origin Access Identities on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// iam
		Function{
			Entity:                "AccessKeys",
			Prefix:                "List",
			Service:               "iam",
			FnAttributeList:       "AccessKeyMetadata",
			SingularEntity:        "AccessKeyMetadata",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetAccessKeys returns all the IAM AccessKeys on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "AccountAliases",
			Prefix:                "List",
			Service:               "iam",
			FnPaginationAttribute: "Marker",
			FnOutput:              "string",
			Documentation: `
			// GetAccountAliases returns all the IAM AccountAliases on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		// Check
		Function{
			Entity:           "AccountPasswordPolicy",
			Prefix:           "Get",
			FnAttributeList:  "PasswordPolicy",
			SingularEntity:   "PasswordPolicy",
			HasNotPagination: true,
			HasNoSlice:       true,
			Service:          "iam",
			Documentation: `
			// GetAccountPasswordPolicy returns the IAM AccountPasswordPolicy on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "Groups",
			Prefix:                "List",
			FnPaginationAttribute: "Marker",
			Service:               "iam",
			Documentation: `
			// GetGroups returns the IAM Groups on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "GroupPolicies",
			Prefix:                "List",
			Service:               "iam",
			FnOutput:              "string",
			FnAttributeList:       "PolicyNames",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetGroupPolicies returns the IAM GroupPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "AttachedGroupPolicies",
			FnAttributeList:       "AttachedPolicies",
			SingularEntity:        "AttachedPolicy",
			FnPaginationAttribute: "Marker",
			Prefix:                "List",
			Service:               "iam",
			Documentation: `
			// GetAttachedGroupPolicies returns the IAM AttachedGroupPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "InstanceProfiles",
			Prefix:                "List",
			Service:               "iam",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetIstanceProfiles returns the IAM InstanceProfiles on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:           "OpenIDConnectProviders",
			Prefix:           "List",
			Service:          "iam",
			HasNotPagination: true,
			FnAttributeList:  "OpenIDConnectProviderList",
			SingularEntity:   "OpenIDConnectProviderListEntry",
			Documentation: `
			// GetOpenIDConnectProviders returns the IAM OpenIDConnectProviders on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "Policies",
			Prefix:                "List",
			Service:               "iam",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetPolicies returns the IAM Policies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "Roles",
			Prefix:                "List",
			FnPaginationAttribute: "Marker",
			Service:               "iam",
			Documentation: `
			// GetRoles returns the IAM Roles on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnOutput:              "string",
			FnAttributeList:       "PolicyNames",
			Entity:                "RolePolicies",
			FnPaginationAttribute: "Marker",
			Prefix:                "List",
			Service:               "iam",
			Documentation: `
			// GetRolePolicies returns the IAM RolePolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "AttachedRolePolicies",
			Prefix:                "List",
			FnPaginationAttribute: "Marker",
			FnAttributeList:       "AttachedPolicies",
			SingularEntity:        "AttachedPolicy",
			Service:               "iam",
			Documentation: `
			// GetAttachedRolePolicies returns the IAM AttachedRolePolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:           "SAMLProviders",
			HasNotPagination: true,
			FnAttributeList:  "SAMLProviderList",
			SingularEntity:   "SAMLProviderListEntry",
			Prefix:           "List",
			Service:          "iam",
			Documentation: `
			// GetSAMLProviders returns the IAM SAMLProviders on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "ServerCertificates",
			FnAttributeList:       "ServerCertificateMetadataList",
			SingularEntity:        "ServerCertificateMetadata",
			Prefix:                "List",
			FnPaginationAttribute: "Marker",
			Service:               "iam",
			Documentation: `
			// GetServerCertificates returns the IAM ServerCertificates on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "Users",
			Prefix:                "List",
			FnPaginationAttribute: "Marker",
			Service:               "iam",
			Documentation: `
			// GetUsers returns the IAM Users on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "UserPolicies",
			Prefix:                "List",
			Service:               "iam",
			FnPaginationAttribute: "Marker",
			FnOutput:              "string",
			FnAttributeList:       "PolicyNames",
			Documentation: `
			// GetUserPolicies returns the IAM UserPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "AttachedUserPolicies",
			FnPaginationAttribute: "Marker",
			FnAttributeList:       "AttachedPolicies",
			SingularEntity:        "AttachedPolicy",
			Prefix:                "List",
			Service:               "iam",
			Documentation: `
			// GetAttachedUserPolicies returns the IAM AttachedUserPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "SSHPublicKeys",
			SingularEntity:        "SSHPublicKeyMetadata",
			FnPaginationAttribute: "Marker",
			Prefix:                "List",
			Service:               "iam",
			Documentation: `
			// GetSSHPublicKeys returns the IAM SSHPublicKeys on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// ses
		Function{
			Entity:           "ActiveReceiptRuleSet",
			Prefix:           "Describe",
			HasNoSlice:       true,
			HasNotPagination: true,
			FnAttributeList:  "Metadata.Name",
			FnOutput:         "string",
			Service:          "ses",
			Documentation: `
			// GetActiveReceiptRuleSet returns the SES ActiveReceiptRuleSet on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:           "ActiveReceiptRuleSet",
			FnName:           "GetActiveReceiptRulesSet",
			Prefix:           "Describe",
			HasNotPagination: true,
			FnAttributeList:  "Rules",
			SingularEntity:   "ReceiptRule",
			Service:          "ses",
			Documentation: `
			// GetActiveReceiptRulesSet returns the SES ActiveReceiptRuleSet on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:   "Identities",
			FnOutput: "string",
			Prefix:   "List",
			Service:  "ses",
			Documentation: `
			// GetIdentities returns the SES Identities on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnAttributeList:  "Filters",
			Entity:           "ReceiptFilters",
			HasNotPagination: true,
			Prefix:           "List",
			Service:          "ses",
			Documentation: `
			// GetReceiptFilters returns the SES ReceiptFilters on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "ConfigurationSets",
			Prefix:  "List",
			Service: "ses",
			Documentation: `
			// GetConfigurationSets returns the SES ConfigurationSets on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:           "IdentityNotificationAttributes",
			Prefix:           "Get",
			Service:          "ses",
			FnAttributeList:  "NotificationAttributes",
			SingularEntity:   "IdentityNotificationAttributes",
			IsMap:            true,
			HasNotPagination: true,
			Documentation: `
			// GetIdentityNotificationAttributes returns the SES IdentityNotificationAttributes on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:          "Templates",
			Prefix:          "List",
			FnAttributeList: "TemplatesMetadata",
			SingularEntity:  "TemplateMetadata",
			Service:         "ses",
			Documentation: `
			// GetTemplates returns the SES Templates on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// route53
		Function{
			Entity:                "ReusableDelegationSets",
			FnAttributeList:       "DelegationSets",
			SingularEntity:        "DelegationSet",
			Prefix:                "List",
			Service:               "route53",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetReusableDelegationSets returns the Route53 ReusableDelegationSets on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "HealthChecks",
			Prefix:                "List",
			Service:               "route53",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetHealthChecks returns the Route53 HealthChecks on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "QueryLoggingConfigs",
			Prefix:  "List",
			Service: "route53",
			Documentation: `
			// GetQueryLoggingConfigs returns the Route53 QueryLoggingConfigs on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                     "ResourceRecordSets",
			Prefix:                     "List",
			Service:                    "route53",
			FnInputPaginationAttribute: "StartRecordName",
			FnPaginationAttribute:      "NextRecordName",
			Documentation: `
			// GetResourceRecordSets returns the Route53 ResourceRecordSets on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:                "HostedZones",
			Prefix:                "List",
			Service:               "route53",
			FnPaginationAttribute: "Marker",
			Documentation: `
			// GetHostedZones returns the Route53 HostedZones on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:          "VPCAssociationAuthorizations",
			Prefix:          "List",
			Service:         "route53",
			FnAttributeList: "VPCs",
			SingularEntity:  "VPC",
			Documentation: `
			// GetVPCAssociationAuthorizations returns the Route53 VPCAssociationAuthorizations on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// route53resolver
		Function{
			Entity:  "ResolverEndpoints",
			Prefix:  "List",
			Service: "route53resolver",
			Documentation: `
			// GetResolverEndpoints returns the Route53Resolver ResolverEndpoints on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "ResolverRules",
			Prefix:  "List",
			Service: "route53resolver",
			Documentation: `
			// GetResolverRules returns the Route53Resolver ResolverRules on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "ResolverRuleAssociations",
			Prefix:  "List",
			Service: "route53resolver",
			Documentation: `
			// GetResolverRuleAssociations returns the Route53Resolver ResolverRuleAssociations on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
	}
)
