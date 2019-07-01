package aws

import (
	"context"
	"fmt"
	"strings"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/cycloidio/terracognita/provider"
	"github.com/cycloidio/terracognita/tag"
	"github.com/pkg/errors"
)

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix aws_ -transform snake -linecomment
const (
	// NoID it's a helper to make the code more readable
	NoID = ""

	// List of all the Resources
	Instance ResourceType = iota
	VPC
	// Do not have them for now as it's not needed
	// but works
	//AMI
	SecurityGroup
	Subnet
	EBSVolume
	// Do not have them for now as it's not needed
	// but works
	//EBSSnapshot
	ElasticacheCluster
	ELB
	ALB
	DBInstance
	S3Bucket
	//S3BucketObject
	CloudfrontDistribution
	CloudfrontOriginAccessIdentity
	CloudfrontPublicKey
	//IAMAccessKey
	IAMAccountAlias
	IAMAccountPasswordPolicy
	IAMGroup
	IAMGroupMembership
	IAMGroupPolicy
	IAMGroupPolicyAttachment
	IAMInstanceProfile
	IAMOpenidConnectProvider
	IAMPolicy
	// As it's deprecated we'll not support it
	//IAMPolicyAttachment
	IAMRole
	IAMRolePolicy
	IAMRolePolicyAttachment
	IAMSAMLProvider // iam_saml_provider
	IAMServerCertificate
	// TODO: Don't know how to get it from AWS SKD
	// IAMServiceLinkedRole
	IAMUser
	IAMUserGroupMembership
	IAMUserPolicy
	IAMUserPolicyAttachment
	Route53DelegationSet
	Route53HealthCheck
	Route53QueryLog
	Route53Record
	Route53Zone
	Route53ZoneAssociation
	Route53ResolverEndpoint
	Route53ResolverRuleAssociation
	SESActiveReceiptRuleSet
	SESDomainIdentity
	SESDomainIdentityVerification
	SESDomainDKIM
	SESDomainMailFrom
	SESReceiptFilter
	SESReceiptRule
	SESReceiptRuleSet
	SESConfigurationSet
	// Read on TF is nil so ...
	// SESEventDestination
	SESIdentityNotificationTopic
	SESTemplate
)

type rtFn func(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		Instance: instances,
		VPC:      vpcs,
		//AMI:      ami,
		SecurityGroup: securityGroups,
		Subnet:        subnets,
		EBSVolume:     ebsVolumes,
		//EBSSnapshot:         ebsSnapshots,
		ElasticacheCluster: elasticacheClusters,
		ELB:                elbs,
		ALB:                albs,
		DBInstance:         dbInstances,
		S3Bucket:           s3Buckets,
		//S3BucketObject:      s3_bucket_objects,
		CloudfrontDistribution:         cloudfrontDistributions,
		CloudfrontOriginAccessIdentity: cloudfrontOriginAccessIdentities,
		CloudfrontPublicKey:            cloudfrontPublicKeys,
		//IAMAccessKey:                   iamAccessKeys,
		IAMAccountAlias:                iamAccountAliases,
		IAMAccountPasswordPolicy:       iamAccountPasswordPolicy,
		IAMGroup:                       cacheIAMGroups,
		IAMGroupMembership:             iamGroupMemberships,
		IAMGroupPolicy:                 iamGroupPolicies,
		IAMGroupPolicyAttachment:       iamGroupPolicyAttachments,
		IAMInstanceProfile:             iamInstanceProfiles,
		IAMOpenidConnectProvider:       iamOpenidConnectProviders,
		IAMPolicy:                      iamPolicies,
		IAMRole:                        cacheIAMRoles,
		IAMRolePolicy:                  iamRolePolicies,
		IAMRolePolicyAttachment:        iamRolePolicyAttachments,
		IAMSAMLProvider:                iamSAMLProviders,
		IAMServerCertificate:           iamServerCertificates,
		IAMUser:                        cacheIAMUsers,
		IAMUserGroupMembership:         iamUserGroupMemberships,
		IAMUserPolicy:                  iamUserPolicies,
		IAMUserPolicyAttachment:        iamUserPolicyAttachments,
		Route53DelegationSet:           route53DelegationSets,
		Route53HealthCheck:             route53HealthChecks,
		Route53QueryLog:                route53QueryLogs,
		Route53Record:                  route53Records,
		Route53Zone:                    cacheRoute53Zones,
		Route53ZoneAssociation:         route53ZoneAssociations,
		Route53ResolverEndpoint:        route53ResolverEndpoints,
		Route53ResolverRuleAssociation: route53ResolverRuleAssociation,
		SESActiveReceiptRuleSet:        sesActiveReceiptRuleSets,
		SESDomainIdentity:              cacheSESDomainIdentities,
		SESDomainIdentityVerification:  sesDomainGeneral,
		SESDomainDKIM:                  sesDomainGeneral,
		SESDomainMailFrom:              sesDomainGeneral,
		SESReceiptFilter:               sesReceiptFilters,
		SESReceiptRule:                 sesReceiptRules,
		SESReceiptRuleSet:              sesReceiptRuleSets,
		SESConfigurationSet:            sesConfigurationSets,
		SESIdentityNotificationTopic:   sesIdentityNotificationTopics,
		SESTemplate:                    sesTemplates,
	}
)

func initializeResource(a *aws, ID, t string) (provider.Resource, error) {
	tfr, ok := a.tfProvider.ResourcesMap[t]
	if !ok {
		return nil, errors.Errorf("the resource %q does not exists on Terraform", t)
	}

	data := tfr.Data(nil)
	data.SetId(ID)
	data.SetType(t)

	return provider.NewResource(
		ID, t, tfr,
		data, a,
	), nil
}

func instances(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	var input = &ec2.DescribeInstancesInput{
		Filters: toEC2Filters(tags),
	}

	instances, err := a.awsr.GetInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range instances[a.Region()].Reservations {
		for _, vv := range v.Instances {
			r, err := initializeResource(a, *vv.InstanceId, resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func vpcs(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	var input = &ec2.DescribeVpcsInput{
		Filters: toEC2Filters(tags),
	}

	vpcs, err := a.awsr.GetVpcs(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range vpcs[a.Region()].Vpcs {
		r, err := initializeResource(a, *v.VpcId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

//func amis(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
//var input = &ec2.DescribeImagesInput{
//Filters: toEC2Filters(tags),
//}

//images, err := a.awsr.GetOwnImages(ctx, input)
//if err != nil {
//return nil, err
//}

//resources := make([]provider.Resource, 0)
//for _, v := range images[a.Region()].Images {
//r, err := initializeResource(a, *v.ImageId, resourceType)
//if err != nil {
//return nil, err
//}
//resources = append(resources, r)
//}

//return resources, nil
//}

func securityGroups(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	var input = &ec2.DescribeSecurityGroupsInput{
		Filters: toEC2Filters(tags),
	}

	sgs, err := a.awsr.GetSecurityGroups(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range sgs[a.Region()].SecurityGroups {
		r, err := initializeResource(a, *v.GroupId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func subnets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	var input = &ec2.DescribeSubnetsInput{
		Filters: toEC2Filters(tags),
	}

	subnets, err := a.awsr.GetSubnets(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range subnets[a.Region()].Subnets {
		r, err := initializeResource(a, *v.SubnetId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func ebsVolumes(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	var input = &ec2.DescribeVolumesInput{
		Filters: toEC2Filters(tags),
	}

	volumes, err := a.awsr.GetVolumes(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range volumes[a.Region()].Volumes {
		r, err := initializeResource(a, *v.VolumeId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

//func ebsSnapshots(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
//var input = &ec2.DescribeSnapshotsInput{
//Filters: toEC2Filters(tags),
//}

//snapshots, err := a.awsr.GetOwnSnapshots(ctx, input)
//if err != nil {
//return nil, err
//}

//resources := make([]provider.Resource, 0)
//for _, v := range snapshots[a.Region()].Snapshots {
//r, err := initializeResource(a, *v.SnapshotId, resourceType)
//if err != nil {
//return nil, err
//}
//resources = append(resources, r)
//}

//return resources, nil
//}

func elasticacheClusters(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	cacheClusters, err := a.awsr.GetElastiCacheClusters(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range cacheClusters[a.Region()].CacheClusters {
		r, err := initializeResource(a, *v.CacheClusterId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func elbs(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	lbs, err := a.awsr.GetLoadBalancers(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range lbs[a.Region()].LoadBalancerDescriptions {
		r, err := initializeResource(a, *v.LoadBalancerName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func albs(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	lbs, err := a.awsr.GetLoadBalancersV2(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range lbs[a.Region()].LoadBalancers {
		r, err := initializeResource(a, *v.LoadBalancerArn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func dbInstances(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	dbs, err := a.awsr.GetDBInstances(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range dbs[a.Region()].DBInstances {
		r, err := initializeResource(a, *v.DBInstanceIdentifier, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func s3Buckets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	buckets, err := a.awsr.ListBuckets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range buckets[a.Region()].Buckets {
		r, err := initializeResource(a, *v.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontDistributions(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	distributions, err := a.awsr.GetCloudFrontDistributions(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range distributions[a.Region()].DistributionList.Items {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontOriginAccessIdentities(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	identitys, err := a.awsr.GetCloudFrontOriginAccessIdentities(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range identitys[a.Region()].CloudFrontOriginAccessIdentityList.Items {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontPublicKeys(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	publicKeys, err := a.awsr.GetCloudFrontPublicKeys(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range publicKeys[a.Region()].PublicKeyList.Items {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

//func iamAccessKeys(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
//// I could get all the users and do the same filtering by user but it
//// seems that it does not need it and returns all the AccessKeys
//accessKeys, err := a.awsr.GetAccessKeys(ctx, nil)
//if err != nil {
//return nil, err
//}

//resources := make([]provider.Resource, 0)
//for _, i := range accessKeys[a.Region()].AccessKeyMetadata {
//r, err := initializeResource(a, *i.AccessKeyId, resourceType)
//if err != nil {
//return nil, err
//}
//err = r.Data().Set("user", i.UserName)
//if err != nil {
//return nil, err
//}
//resources = append(resources, r)
//}

//return resources, nil
//}

func iamAccountAliases(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	accountAliases, err := a.awsr.GetAccountAliases(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range accountAliases[a.Region()].AccountAliases {
		r, err := initializeResource(a, *i, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamAccountPasswordPolicy(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	// As it's for the full account we'll tell TF to fetch it directly with a "" id
	r, err := initializeResource(a, NoID, resourceType)
	if err != nil {
		return nil, err
	}
	return []provider.Resource{r}, nil
}

func iamGroups(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	groups, err := a.awsr.GetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range groups[a.Region()].Groups {
		r, err := initializeResource(a, *i.GroupName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamGroupMemberships(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range groupNames {
		r, err := initializeResource(a, NoID, resourceType)
		if err != nil {
			return nil, err
		}
		err = r.Data().Set("group", i)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamGroupPolicies(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, gn := range groupNames {
		input := &iam.ListGroupPoliciesInput{
			GroupName: awsSDK.String(gn),
		}
		groupPolicies, err := a.awsr.GetGroupPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range groupPolicies[a.Region()].PolicyNames {
			// It needs the ID to be "GN:PN"
			// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_group_policy.go#L134:6
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", gn, *i), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamGroupPolicyAttachments(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, gn := range groupNames {
		input := &iam.ListAttachedGroupPoliciesInput{
			GroupName: awsSDK.String(gn),
		}
		groupPolicies, err := a.awsr.GetAttachedGroupPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range groupPolicies[a.Region()].AttachedPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", gn, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamInstanceProfiles(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	instanceProfiles, err := a.awsr.GetInstanceProfiles(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range instanceProfiles[a.Region()].InstanceProfiles {
		r, err := initializeResource(a, *i.InstanceProfileName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamOpenidConnectProviders(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	openIDConnectProviders, err := a.awsr.GetOpenIDConnectProviders(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range openIDConnectProviders[a.Region()].OpenIDConnectProviderList {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamPolicies(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	input := &iam.ListPoliciesInput{
		Scope: awsSDK.String("Local"),
	}
	policies, err := a.awsr.GetPolicies(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range policies[a.Region()].Policies {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamRoles(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	roles, err := a.awsr.GetRoles(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range roles[a.Region()].Roles {
		r, err := initializeResource(a, *i.RoleName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamRolePolicies(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	roleNames, err := getIAMRoleNames(ctx, a, IAMRole.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, rn := range roleNames {
		input := &iam.ListRolePoliciesInput{
			RoleName: awsSDK.String(rn),
		}
		rolePolicies, err := a.awsr.GetRolePolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range rolePolicies[a.Region()].PolicyNames {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", rn, *i), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamRolePolicyAttachments(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	roleNames, err := getIAMRoleNames(ctx, a, IAMRole.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, rn := range roleNames {
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: awsSDK.String(rn),
		}
		rolePolicies, err := a.awsr.GetAttachedRolePolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range rolePolicies[a.Region()].AttachedPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", rn, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamSAMLProviders(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	samalProviders, err := a.awsr.GetSAMLProviders(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range samalProviders[a.Region()].SAMLProviderList {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamServerCertificates(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	serverCertificates, err := a.awsr.GetServerCertificates(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range serverCertificates[a.Region()].ServerCertificateMetadataList {
		r, err := initializeResource(a, *i.ServerCertificateName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUsers(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	users, err := a.awsr.GetUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range users[a.Region()].Users {
		r, err := initializeResource(a, *i.UserName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUserGroupMemberships(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), tags)
	if err != nil {
		return nil, err
	}

	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, un := range userNames {
		r, err := initializeResource(a, strings.Join(append([]string{un}, groupNames...), "/"), resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUserPolicies(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, un := range userNames {
		input := &iam.ListUserPoliciesInput{
			UserName: awsSDK.String(un),
		}
		userPolicies, err := a.awsr.GetUserPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range userPolicies[a.Region()].PolicyNames {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", un, *i), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamUserPolicyAttachments(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, un := range userNames {
		input := &iam.ListAttachedUserPoliciesInput{
			UserName: awsSDK.String(un),
		}
		userPolicies, err := a.awsr.GetAttachedUserPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range userPolicies[a.Region()].AttachedPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", un, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func route53DelegationSets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53DelegationSets, err := a.awsr.GetReusableDelegationSets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53DelegationSets[a.Region()].DelegationSets {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53HealthChecks(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53HealthChecks, err := a.awsr.GetHealthChecks(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53HealthChecks[a.Region()].HealthChecks {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53QueryLogs(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53QueryLogs, err := a.awsr.GetQueryLoggingConfigs(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53QueryLogs[a.Region()].QueryLoggingConfigs {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53Zones(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53Zones, err := a.awsr.GetHostedZones(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53Zones[a.Region()].HostedZones {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53Records(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	zones, err := getRoute53ZoneIDs(ctx, a, Route53Zone.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, z := range zones {
		input := &route53.ListResourceRecordSetsInput{
			HostedZoneId: awsSDK.String(z),
		}
		r53Records, err := a.awsr.GetResourceRecordSets(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range r53Records[a.Region()].ResourceRecordSets {
			id := []string{z, strings.ToLower(*i.Name), *i.Type}
			if i.SetIdentifier != nil {
				id = append(id, *i.SetIdentifier)
			}
			r, err := initializeResource(a, strings.Join(id, "_"), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func route53ZoneAssociations(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	zones, err := getRoute53ZoneIDs(ctx, a, Route53Zone.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, z := range zones {
		input := &route53.ListVPCAssociationAuthorizationsInput{
			HostedZoneId: awsSDK.String(z),
		}
		r53ZoneAssociations, err := a.awsr.GetVPCAssociationAuthorizations(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range r53ZoneAssociations[a.Region()].VPCs {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", z, *i.VPCId), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func route53ResolverEndpoints(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53ResolverEndpoints, err := a.awsr.GetResolverEndpoints(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53ResolverEndpoints[a.Region()].ResolverEndpoints {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53ResolverRuleAssociation(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	r53ResolverRuleAssociations, err := a.awsr.GetResolverRuleAssociations(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53ResolverRuleAssociations[a.Region()].ResolverRuleAssociations {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesActiveReceiptRuleSets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	if sesActiveReceiptRuleSets[a.Region()].Metadata == nil {
		return nil, nil
	}

	r, err := initializeResource(a, *sesActiveReceiptRuleSets[a.Region()].Metadata.Name, resourceType)
	if err != nil {
		return nil, err
	}

	return []provider.Resource{r}, nil
}

func sesDomainIdentities(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesDomainIdentities, err := a.awsr.GetIdentities(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesDomainIdentities[a.Region()].Identities {
		r, err := initializeResource(a, *i, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesDomainGeneral(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	domainNames, err := getSESDomainIdentityDomains(ctx, a, SESDomainIdentity.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range domainNames {
		r, err := initializeResource(a, i, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesReceiptFilters(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesReceiptFilters, err := a.awsr.GetReceiptFilters(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesReceiptFilters[a.Region()].Filters {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesReceiptRules(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesActiveReceiptRuleSets[a.Region()].Rules {
		r, err := initializeResource(a, fmt.Sprintf("%s:%s", *sesActiveReceiptRuleSets[a.Region()].Metadata.Name, *i.Name), resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesReceiptRuleSets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	if sesActiveReceiptRuleSets[a.Region()].Metadata == nil {
		return nil, nil
	}

	r, err := initializeResource(a, *sesActiveReceiptRuleSets[a.Region()].Metadata.Name, resourceType)
	if err != nil {
		return nil, err
	}

	return []provider.Resource{r}, nil
}

func sesConfigurationSets(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesConfigurationSets, err := a.awsr.GetConfigurationSets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesConfigurationSets[a.Region()].ConfigurationSets {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesIdentityNotificationTopics(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	domainNames, err := getSESDomainIdentityDomains(ctx, a, SESDomainIdentity.String(), tags)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, d := range domainNames {
		// We could just pass domainNames as Identities
		// but then we would not not which NotificationAttributes
		// is of each identity so we have to do it one by one
		input := &ses.GetIdentityNotificationAttributesInput{
			Identities: []*string{&d},
		}

		sesIdentityNotificationTopics, err := a.awsr.GetIdentityNotificationAttributes(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range sesIdentityNotificationTopics[a.Region()].NotificationAttributes {
			var notType string
			if i.BounceTopic != nil {
				notType = ses.NotificationTypeBounce
			} else if i.ComplaintTopic != nil {
				notType = ses.NotificationTypeComplaint
			} else if i.DeliveryTopic != nil {
				notType = ses.NotificationTypeDelivery
			} else {
				// We need the topic, if fore some reason we do not have
				// it we have to continue to the next one
				continue
			}
			r, err := initializeResource(a, fmt.Sprintf("%s|%s", d, notType), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func sesTemplates(ctx context.Context, a *aws, resourceType string, tags []tag.Tag) ([]provider.Resource, error) {
	sesTemplates, err := a.awsr.GetTemplates(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesTemplates[a.Region()].TemplatesMetadata {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func toEC2Filters(tags []tag.Tag) []*ec2.Filter {
	if len(tags) == 0 {
		return nil
	}
	filters := make([]*ec2.Filter, 0, len(tags))

	for _, t := range tags {
		filters = append(filters, t.ToEC2Filter())
	}

	return filters
}
