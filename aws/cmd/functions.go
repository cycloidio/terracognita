package main

var (
	// functions is the list of fuctions that will be added
	// to the AWSReader with the corresponding implementation
	functions = []Function{
		// ec2
		Function{
			Entity:  "Instances",
			Prefix:  "Describe",
			Service: "ec2",
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
			Entity:  "Images",
			Prefix:  "Describe",
			Service: "ec2",
			Documentation: `
			// GetImages returns all EC2 AMI based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:        "Images",
			Prefix:        "Describe",
			Service:       "ec2",
			FilterByOwner: "Owners",
			Documentation: `
			// GetOwnImages returns all EC2 AMI belonging to the Account ID based on the input given.
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
			Entity:  "AutoScalingGroups",
			Prefix:  "Describe",
			Service: "autoscaling",
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

		// elasticache
		Function{
			FnName:  "GetElastiCacheClusters",
			Entity:  "CacheClusters",
			Prefix:  "Describe",
			Service: "elasticache",
			Documentation: `
			// GetElastiCacheClusters returns all Elasticache clusters based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:   "GetElastiCacheTags",
			Entity:   "TagsForResource",
			Prefix:   "List",
			Service:  "elasticache",
			FnOutput: "TagListMessage",
			Documentation: `
			// GetElastiCacheTags returns a list of tags of Elasticache resources based on its ARN.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// elb
		Function{
			Entity:  "LoadBalancers",
			Prefix:  "Describe",
			Service: "elb",
			Documentation: `
			// GetLoadBalancers returns a list of ELB (v1) based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetLoadBalancersTags",
			Entity:  "Tags",
			Prefix:  "Describe",
			Service: "elb",
			Documentation: `
			// GetLoadBalancersTags returns a list of Tags based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// elbv2
		Function{
			FnName:  "GetLoadBalancersV2",
			Entity:  "LoadBalancers",
			Prefix:  "Describe",
			Service: "elbv2",
			Documentation: `
			// GetLoadBalancersV2 returns a list of ELB (v2) - also known as ALB - based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetLoadBalancersV2Tags",
			Entity:  "Tags",
			Prefix:  "Describe",
			Service: "elbv2",
			Documentation: `
			// GetLoadBalancersV2Tags returns a list of Tags based on the input from the different regions.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// rds
		Function{
			Entity:  "DBInstances",
			Prefix:  "Describe",
			Service: "rds",
			Documentation: `
			// GetDBInstances returns all DB instances based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetDBInstancesTags",
			Entity:  "TagsForResource",
			Prefix:  "List",
			Service: "rds",
			Documentation: `
			// GetDBInstancesTags returns a list of tags from an ARN, extra filters for tags can also be provided.
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// s3
		Function{
			// TODO: https://github.com/cycloidio/raws/issues/44
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
			FnName:  "GetBucketTags",
			Entity:  "BucketTagging",
			Prefix:  "Get",
			Service: "s3",
			Documentation: `
			// GetBucketTags returns tags associated with S3 buckets based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			// TODO: https://github.com/cycloidio/raws/issues/44
			FnName:  "ListObjects",
			Entity:  "Objects",
			Prefix:  "List",
			Service: "s3",
			Documentation: `
			// ListObjects returns a list of all S3 objects in a bucket based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetObjectsTags",
			Entity:  "ObjectTagging",
			Prefix:  "Get",
			Service: "s3",
			Documentation: `
			// GetObjectsTags returns tags associated with S3 objects based on the input given.
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetRecordedResourceCounts",
			Entity:  "DiscoveredResourceCounts",
			Prefix:  "Get",
			Service: "configservice",
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
			FnName:  "GetCloudFrontDistributions",
			Entity:  "Distributions",
			Prefix:  "List",
			Service: "cloudfront",
			Documentation: `
			// GetCloudFrontDistributions returns all the CloudFront Distributions on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			FnName:  "GetCloudFrontPublicKeys",
			Entity:  "PublicKeys",
			Prefix:  "List",
			Service: "cloudfront",
			Documentation: `
			// GetCloudFrontPublicKeys returns all the CloudFront Public Keys on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "CloudFrontOriginAccessIdentities",
			Prefix:  "List",
			Service: "cloudfront",
			Documentation: `
			// GetCloudFrontOriginAccessIdentities returns all the CloudFront Origin Access Identities on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// iam
		Function{
			Entity:  "AccessKeys",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetAccessKeys returns all the IAM AccessKeys on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "AccountAliases",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetAccountAliases returns all the IAM AccountAliases on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		// Check
		Function{
			Entity:  "AccountPasswordPolicy",
			Prefix:  "Get",
			Service: "iam",
			Documentation: `
			// GetAccountPasswordPolicy returns the IAM AccountPasswordPolicy on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Groups",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetGroups returns the IAM Groups on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "GroupPolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetGroupPolicies returns the IAM GroupPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "AttachedGroupPolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetAttachedGroupPolicies returns the IAM AttachedGroupPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "InstanceProfiles",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetIstanceProfiles returns the IAM InstanceProfiles on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "OpenIDConnectProviders",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetOpenIDConnectProviders returns the IAM OpenIDConnectProviders on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Policies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetPolicies returns the IAM Policies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Roles",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetRoles returns the IAM Roles on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "RolePolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetRolePolicies returns the IAM RolePolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "AttachedRolePolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetAttachedRolePolicies returns the IAM AttachedRolePolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "SAMLProviders",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetSAMLProviders returns the IAM SAMLProviders on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "ServerCertificates",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetServerCertificates returns the IAM ServerCertificates on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Users",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetUsers returns the IAM Users on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "UserPolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetUserPolicies returns the IAM UserPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "AttachedUserPolicies",
			Prefix:  "List",
			Service: "iam",
			Documentation: `
			// GetAttachedUserPolicies returns the IAM AttachedUserPolicies on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "SSHPublicKey",
			Prefix:  "Get",
			Service: "iam",
			Documentation: `
			// GetSSHPublicKey returns the IAM SSHPublicKey on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		// ses
		Function{
			Entity:  "ActiveReceiptRuleSet",
			Prefix:  "Describe",
			Service: "ses",
			Documentation: `
			// GetActiveReceiptRuleSet returns the SES ActiveReceiptRuleSet on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Identities",
			Prefix:  "List",
			Service: "ses",
			Documentation: `
			// GetIdentities returns the SES Identities on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "ReceiptFilters",
			Prefix:  "List",
			Service: "ses",
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
			Entity:  "IdentityNotificationAttributes",
			Prefix:  "Get",
			Service: "ses",
			Documentation: `
			// GetIdentityNotificationAttributes returns the SES IdentityNotificationAttributes on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "Templates",
			Prefix:  "List",
			Service: "ses",
			Documentation: `
			// GetTemplates returns the SES Templates on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},

		// route53
		Function{
			Entity:  "ReusableDelegationSets",
			Prefix:  "List",
			Service: "route53",
			Documentation: `
			// GetReusableDelegationSets returns the Route53 ReusableDelegationSets on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "HealthChecks",
			Prefix:  "List",
			Service: "route53",
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
			Entity:  "ResourceRecordSets",
			Prefix:  "List",
			Service: "route53",
			Documentation: `
			// GetResourceRecordSets returns the Route53 ResourceRecordSets on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "HostedZones",
			Prefix:  "List",
			Service: "route53",
			Documentation: `
			// GetHostedZones returns the Route53 HostedZones on the given input
			// Returned values are commented in the interface doc comment block.
			`,
		},
		Function{
			Entity:  "VPCAssociationAuthorizations",
			Prefix:  "List",
			Service: "route53",
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
