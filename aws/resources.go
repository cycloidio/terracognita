package aws

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/cycloidio/terracognita/filter"
	"github.com/cycloidio/terracognita/provider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	// Do not have them for now as it's not needed
	// but works
	//AMI

	// Do not have them for now as it's not needed
	// but works
	//EBSSnapshot

	ALB
	ALBListener
	ALBListenerCertificate
	ALBListenerRule
	ALBTargetGroup
	ALBTargetGroupAttachment
	APIGatewayDeployment
	APIGatewayStage
	APIGatewayResource
	APIGatewayRestAPI
	AutoscalingGroup
	AutoscalingPolicy
	CloudfrontDistribution
	CloudfrontOriginAccessIdentity
	CloudfrontPublicKey
	CloudwatchMetricAlarm
	DBInstance
	DBParameterGroup
	DBSubnetGroup
	EBSVolume
	ElasticacheCluster
	ElasticsearchDomain
	ElasticsearchDomainPolicy
	ELB
	IAMAccessKey
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
	IAMUserSSHKey
	KeyPair
	LambdaFunction
	LaunchConfiguration
	LaunchTemplate
	LB
	LBCookieStickinessPolicy
	LBListener
	LBListenerCertificate
	LBListenerRule
	LBTargetGroup
	LBTargetGroupAttachment
	Route53DelegationSet
	Route53HealthCheck
	Route53QueryLog
	Route53Record
	Route53ResolverEndpoint
	Route53ResolverRuleAssociation
	Route53Zone
	Route53ZoneAssociation
	S3Bucket
	//S3BucketObject
	SecurityGroup
	SESActiveReceiptRuleSet
	SESConfigurationSet
	SESDomainDKIM
	SESDomainIdentity
	SESDomainIdentityVerification
	SESDomainMailFrom
	// Read on TF is nil so ...
	// SESEventDestination
	SESIdentityNotificationTopic
	SESReceiptFilter
	SESReceiptRule
	SESReceiptRuleSet
	SESTemplate
	Subnet
	VolumeAttachment
	VPC
	VPCPeeringConnection
)

type rtFn func(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error)

var (
	resources = map[ResourceType]rtFn{
		ALB:                      cacheLoadBalancersV2,
		ALBListener:              cacheLoadBalancersV2Listeners,
		ALBListenerCertificate:   albListenerCertificates,
		ALBListenerRule:          albListenerRules,
		ALBTargetGroup:           albTargetGroups,
		ALBTargetGroupAttachment: albTargetGroupAttachments,
		//AMI:      ami,
		APIGatewayDeployment:           apiGatewayDeployments,
		APIGatewayStage:                apiGatewayStages,
		APIGatewayResource:             apiGatewayResources,
		APIGatewayRestAPI:              apiGatewayRestApis,
		AutoscalingGroup:               autoscalingGroups,
		AutoscalingPolicy:              autoscalingPolicies,
		CloudfrontDistribution:         cloudfrontDistributions,
		CloudfrontOriginAccessIdentity: cloudfrontOriginAccessIdentities,
		CloudfrontPublicKey:            cloudfrontPublicKeys,
		CloudwatchMetricAlarm:          cloudwatchMetricAlarms,
		DBInstance:                     dbInstances,
		DBParameterGroup:               dbParameterGroups,
		DBSubnetGroup:                  dbSubnetGroups,
		//EBSSnapshot:         ebsSnapshots,
		EBSVolume:                      ebsVolumes,
		ElasticacheCluster:             elasticacheClusters,
		ElasticsearchDomain:            elasticsearchDomains,
		ElasticsearchDomainPolicy:      elasticsearchDomains,
		ELB:                            elbs,
		IAMAccessKey:                   iamAccessKeys,
		IAMAccountAlias:                iamAccountAliases,
		IAMAccountPasswordPolicy:       iamAccountPasswordPolicy,
		IAMGroup:                       cacheIAMGroups,
		IAMGroupMembership:             iamGroupMemberships,
		IAMGroupPolicyAttachment:       iamGroupPolicyAttachments,
		IAMGroupPolicy:                 iamGroupPolicies,
		IAMInstanceProfile:             iamInstanceProfiles,
		IAMOpenidConnectProvider:       iamOpenidConnectProviders,
		IAMPolicy:                      iamPolicies,
		IAMRole:                        cacheIAMRoles,
		IAMRolePolicyAttachment:        iamRolePolicyAttachments,
		IAMRolePolicy:                  iamRolePolicies,
		IAMSAMLProvider:                iamSAMLProviders,
		IAMServerCertificate:           iamServerCertificates,
		IAMUser:                        cacheIAMUsers,
		IAMUserGroupMembership:         iamUserGroupMemberships,
		IAMUserPolicyAttachment:        iamUserPolicyAttachments,
		IAMUserPolicy:                  iamUserPolicies,
		IAMUserSSHKey:                  iamUserSSHKeys,
		Instance:                       instances,
		KeyPair:                        keyPairs,
		LambdaFunction:                 lambdaFunctions,
		LaunchConfiguration:            launchConfigurations,
		LaunchTemplate:                 launchTemplates,
		LB:                             cacheLoadBalancersV2,
		LBCookieStickinessPolicy:       lbCookieStickinessPolicies,
		LBListener:                     cacheLoadBalancersV2Listeners,
		LBListenerCertificate:          albListenerCertificates,
		LBListenerRule:                 albListenerRules,
		LBTargetGroup:                  albTargetGroups,
		LBTargetGroupAttachment:        albTargetGroupAttachments,
		Route53DelegationSet:           route53DelegationSets,
		Route53HealthCheck:             route53HealthChecks,
		Route53QueryLog:                route53QueryLogs,
		Route53Record:                  route53Records,
		Route53ResolverEndpoint:        route53ResolverEndpoints,
		Route53ResolverRuleAssociation: route53ResolverRuleAssociation,
		Route53ZoneAssociation:         route53ZoneAssociations,
		Route53Zone:                    cacheRoute53Zones,
		//S3BucketObject:      s3_bucket_objects,
		S3Bucket:                      s3Buckets,
		SecurityGroup:                 securityGroups,
		SESActiveReceiptRuleSet:       sesActiveReceiptRuleSets,
		SESConfigurationSet:           sesConfigurationSets,
		SESDomainDKIM:                 sesDomainGeneral,
		SESDomainIdentity:             cacheSESDomainIdentities,
		SESDomainIdentityVerification: sesDomainGeneral,
		SESDomainMailFrom:             sesDomainGeneral,
		SESIdentityNotificationTopic:  sesIdentityNotificationTopics,
		SESReceiptFilter:              sesReceiptFilters,
		SESReceiptRule:                sesReceiptRules,
		SESReceiptRuleSet:             sesReceiptRuleSets,
		SESTemplate:                   sesTemplates,
		Subnet:                        subnets,
		VolumeAttachment:              volumeAttachments,
		VPCPeeringConnection:          vpcPeeringConnections,
		VPC:                           vpcs,
	}
)

func initializeResource(a *aws, ID, t string) (provider.Resource, error) {
	return provider.NewResource(ID, t, a), nil
}

func instances(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeInstancesInput{
		Filters: toEC2Filters(filters),
	}

	instances, err := a.awsr.GetInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range instances {
		r, err := initializeResource(a, *i.InstanceId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func vpcs(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeVpcsInput{
		Filters: toEC2Filters(filters),
	}

	vpcs, err := a.awsr.GetVpcs(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range vpcs {
		r, err := initializeResource(a, *v.VpcId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

//func amis(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
//var input = &ec2.DescribeImagesInput{
//Filters: toEC2Filters(filters),
//}

//images, err := a.awsr.GetOwnImages(ctx, input)
//if err != nil {
//return nil, err
//}

//resources := make([]provider.Resource, 0)
//for _, v := range images.Images {
//r, err := initializeResource(a, *v.ImageId, resourceType)
//if err != nil {
//return nil, err
//}
//resources = append(resources, r)
//}

//return resources, nil
//}

func vpcPeeringConnections(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeVpcPeeringConnectionsInput{
		Filters: toEC2Filters(filters),
	}

	vpcPeeringConnections, err := a.awsr.GetVpcPeeringConnections(ctx, input)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range vpcPeeringConnections {

		r, err := initializeResource(a, *i.VpcPeeringConnectionId, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func keyPairs(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeKeyPairsInput{
		Filters: toEC2Filters(filters),
	}

	keyPairs, err := a.awsr.GetKeyPairs(ctx, input)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range keyPairs {

		r, err := initializeResource(a, *i.KeyName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func securityGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeSecurityGroupsInput{
		Filters: toEC2Filters(filters),
	}

	sgs, err := a.awsr.GetSecurityGroups(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range sgs {
		r, err := initializeResource(a, *v.GroupId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func subnets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeSubnetsInput{
		Filters: toEC2Filters(filters),
	}

	subnets, err := a.awsr.GetSubnets(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range subnets {
		r, err := initializeResource(a, *v.SubnetId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func ebsVolumes(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeVolumesInput{
		Filters: toEC2Filters(filters),
	}

	volumes, err := a.awsr.GetVolumes(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range volumes {

		// if aws_instance defined, attached volume are done by ebs_block_device block.
		if (len(v.Attachments) != 0) && (filters.IsIncluded("aws_instance") && !filters.IsExcluded("aws_instance")) {
			continue
		}

		r, err := initializeResource(a, *v.VolumeId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func volumeAttachments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if aws_instance defined, attachment are done by ebs_block_device block.
	if filters.IsIncluded("aws_instance") && !filters.IsExcluded("aws_instance") {
		return nil, nil
	}

	var input = &ec2.DescribeVolumesInput{
		Filters: toEC2Filters(filters),
	}

	volumes, err := a.awsr.GetVolumes(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)

	for _, v := range volumes {
		for _, attach := range v.Attachments {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s:%s", *attach.Device, *v.VolumeId, *attach.InstanceId), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

//func ebsSnapshots(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
//var input = &ec2.DescribeSnapshotsInput{
//Filters: toEC2Filters(filters),
//}

//snapshots, err := a.awsr.GetOwnSnapshots(ctx, input)
//if err != nil {
//return nil, err
//}

//resources := make([]provider.Resource, 0)
//for _, v := range snapshots.Snapshots {
//r, err := initializeResource(a, *v.SnapshotId, resourceType)
//if err != nil {
//return nil, err
//}
//resources = append(resources, r)
//}

//return resources, nil
//}

func elasticacheClusters(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	cacheClusters, err := a.awsr.GetElastiCacheClusters(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range cacheClusters {
		r, err := initializeResource(a, *v.CacheClusterId, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func elasticsearchDomains(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// this function is use for both aws_elasticsearch_domain and aws_elasticsearch_domain_policy
	// if both defined, execute only aws_elasticsearch_domain
	if filters.IsIncluded("aws_elasticsearch_domain", "aws_elasticsearch_domain_policy") && (!filters.IsExcluded("aws_elasticsearch_domain") && resourceType == "aws_elasticsearch_domain_policy") {
		return nil, nil
	}

	dnames, err := a.awsr.GetElasticsearchDomainNames(ctx, nil)
	if err != nil {
		return nil, err
	}

	var names []*string
	for _, dn := range dnames {
		names = append(names, dn.DomainName)
	}

	input := &elasticsearchservice.DescribeElasticsearchDomainsInput{
		DomainNames: names,
	}

	domains, err := a.awsr.GetElasticsearchDomains(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, d := range domains {

		if resourceType == "aws_elasticsearch_domain" {
			// Generate aws_elasticsearch_domain
			r, err := initializeResource(a, *d.DomainName, resourceType)
			if err != nil {
				return nil, err
			}

			resources = append(resources, r)
		}

		// if aws_elasticsearch_domain_policy, create resource
		if resourceType == "aws_elasticsearch_domain_policy" {
			// Generate aws_elasticsearch_domain_policy
			r2, err := initializeResource(a, *d.DomainName, resourceType)
			if err != nil {
				return nil, err
			}
			// TODO this resource is not importable. Define our own ResourceImporter
			// Should be removed when terraform will support it
			// more detail: https://github.com/cycloidio/terracognita/issues/120
			importer := &schema.ResourceImporter{
				State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
					d.Set("domain_name", d.Id())
					d.SetId("esd-policy-" + d.Id())

					return []*schema.ResourceData{d}, nil
				},
			}

			r2.SetImporter(importer)
			resources = append(resources, r2)
		}

	}

	return resources, nil
}

func elbs(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbs, err := a.awsr.GetLoadBalancers(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range lbs {
		r, err := initializeResource(a, *v.LoadBalancerName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func lbCookieStickinessPolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbs, err := a.awsr.GetLoadBalancers(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)

	for _, l := range lbs {
		for _, listener := range l.ListenerDescriptions {
			input := &elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: l.LoadBalancerName,
				PolicyNames:      listener.PolicyNames,
			}

			policies, err := a.awsr.GetLoadBalancerPolicies(ctx, input)
			if err != nil {
				return nil, err
			}
			for _, i := range policies {
				if *i.PolicyTypeName == "LBCookieStickinessPolicyType" {
					//lbName, lbPort, policyName
					r, err := initializeResource(a, fmt.Sprintf("%s:%d:%s", *l.LoadBalancerName, *listener.Listener.LoadBalancerPort, *i.PolicyName), resourceType)
					if err != nil {
						return nil, err
					}

					// TODO this resource is not importable. Define our own ResourceImporter
					// Should be removed when terraform will support it
					// more detail: https://github.com/cycloidio/terracognita/issues/120
					importer := &schema.ResourceImporter{
						State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
							parts := strings.SplitN(d.Id(), ":", 3)

							if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
								return nil, fmt.Errorf("unexpected format of ID (%s), expected lbName:lbPort:policyName", d.Id())
							}

							lbPort, err := strconv.Atoi(parts[1])
							if err != nil {
								return nil, fmt.Errorf("unexpected loadbalancer port (%s)", parts[1])
							}

							d.Set("load_balancer", parts[0])
							d.Set("lb_port", lbPort)
							d.Set("name", fmt.Sprintf("%s-%s-stickiness", parts[0], parts[1]))
							d.SetId(fmt.Sprintf("%s:%s:%s", parts[0], parts[1], parts[2]))

							return []*schema.ResourceData{d}, nil
						},
					}

					r.SetImporter(importer)

					resources = append(resources, r)
				}
			}
		}
	}

	return resources, nil
}

func albs(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lbs, err := a.awsr.GetLoadBalancersV2(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range lbs {
		r, err := initializeResource(a, *v.LoadBalancerArn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func albListeners(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	ALBArns, err := getLoadBalancersV2Arns(ctx, a, ALB.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, alb := range ALBArns {

		input := &elbv2.DescribeListenersInput{
			LoadBalancerArn: awsSDK.String(alb),
		}

		albListeners, err := a.awsr.GetLoadBalancersV2Listeners(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range albListeners {
			r, err := initializeResource(a, *i.ListenerArn, resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func albListenerRules(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both defined, keep only aws_alb_listener_rule
	if filters.IsIncluded("aws_alb_listener_rule", "aws_lb_listener_rule") && (!filters.IsExcluded("aws_alb_listener_rule") && resourceType == "aws_lb_listener_rule") {
		return nil, nil
	}

	ALBListeners, err := getLoadBalancersV2ListenersArns(ctx, a, ALBListener.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, l := range ALBListeners {

		input := &elbv2.DescribeRulesInput{
			ListenerArn: awsSDK.String(l),
		}

		albListenerRules, err := a.awsr.GetLoadBalancersV2Rules(ctx, input)

		if err != nil {
			return nil, err
		}

		for _, i := range albListenerRules {

			r, err := initializeResource(a, *i.RuleArn, resourceType)

			if err != nil {
				return nil, err
			}

			resources = append(resources, r)
		}
	}

	return resources, nil
}

func albListenerCertificates(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both defined, keep only aws_alb_listener_certificate
	if filters.IsIncluded("aws_alb_listener_certificate", "aws_lb_listener_certificate") && (!filters.IsExcluded("aws_alb_listener_certificate") && resourceType == "aws_lb_listener_certificate") {
		return nil, nil
	}

	albListernerIncluded := false
	if (filters.IsIncluded("aws_alb_listener") && !filters.IsExcluded("aws_alb_listener")) || (filters.IsIncluded("aws_lb_listener") && !filters.IsExcluded("aws_lb_listener")) {
		albListernerIncluded = true
	}

	ALBListeners, err := getLoadBalancersV2ListenersArns(ctx, a, ALBListener.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, l := range ALBListeners {

		input := &elbv2.DescribeListenerCertificatesInput{
			ListenerArn: awsSDK.String(l),
		}

		albListenerCertificates, err := a.awsr.GetListenerCertificates(ctx, input)

		if err != nil {
			return nil, err
		}

		for _, i := range albListenerCertificates {
			// if filter include aws_alb_listener, check if *i.IsDefault not defined (since default it is already written by aws_alb_listener)
			if albListernerIncluded && *i.IsDefault {
				continue
			}

			r, err := initializeResource(a, fmt.Sprintf("%s_%s", l, *i.CertificateArn), resourceType)
			if err != nil {
				return nil, err
			}

			// TODO this resource is not importable. Define our own ResourceImporter
			// Should be removed when terraform will support it
			// more detail: https://github.com/cycloidio/terracognita/issues/120
			importer := &schema.ResourceImporter{
				State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
					parts := strings.SplitN(d.Id(), "_", 2)

					if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
						return nil, fmt.Errorf("unexpected format of ID (%s), expected listenerArn_certificateArn", d.Id())
					}
					d.Set("listener_arn", parts[0])
					d.Set("certificate_arn", parts[1])
					d.SetId(fmt.Sprintf("%s_%s", parts[0], parts[1]))

					return []*schema.ResourceData{d}, nil
				},
			}

			r.SetImporter(importer)

			resources = append(resources, r)
		}
	}

	return resources, nil
}

func albTargetGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both defined, keep only aws_alb_target_group
	if filters.IsIncluded("aws_alb_target_group", "aws_lb_target_group") && (!filters.IsExcluded("aws_alb_target_group") && resourceType == "aws_lb_target_group") {
		return nil, nil
	}

	albTargetGroups, err := a.awsr.GetLoadBalancersV2TargetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range albTargetGroups {

		r, err := initializeResource(a, *i.TargetGroupArn, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func albTargetGroupAttachments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both defined, keep only aws_alb_target_group_attachment
	if filters.IsIncluded("aws_alb_target_group_attachment", "aws_lb_target_group_attachment") && (!filters.IsExcluded("aws_alb_target_group_attachment") && resourceType == "aws_lb_target_group_attachment") {
		return nil, nil
	}

	albTargetGroups, err := a.awsr.GetLoadBalancersV2TargetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range albTargetGroups {

		input := &elbv2.DescribeTargetHealthInput{
			TargetGroupArn: i.TargetGroupArn,
		}

		targetHealths, err := a.awsr.GetLoadBalancersV2TargetHealth(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, t := range targetHealths {
			r, err := initializeResource(a, fmt.Sprintf("%s_%d_%s", *t.Target.Id, *t.Target.Port, *i.TargetGroupArn), resourceType)
			if err != nil {
				return nil, err
			}

			// TODO this resource is not importable. Define our own ResourceImporter
			// Should be removed when terraform will support it
			// more detail: https://github.com/cycloidio/terracognita/issues/120
			importer := &schema.ResourceImporter{
				State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
					parts := strings.SplitN(d.Id(), "_", 3)

					if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
						return nil, fmt.Errorf("unexpected format of ID (%s), expected targetId_port_TargetGroupArn", d.Id())
					}

					tPort, err := strconv.Atoi(parts[1])
					if err != nil {
						return nil, fmt.Errorf("unexpected target port (%s)", parts[1])
					}

					d.Set("target_id", parts[0])
					d.Set("port", tPort)
					d.Set("target_group_arn", parts[2])

					d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", parts[2])))

					return []*schema.ResourceData{d}, nil
				},
			}

			r.SetImporter(importer)

			resources = append(resources, r)
		}
	}

	return resources, nil
}

func dbInstances(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	dbs, err := a.awsr.GetDBInstances(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range dbs {
		r, err := initializeResource(a, *v.DBInstanceIdentifier, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func dbParameterGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	dbParameterGroups, err := a.awsr.GetDBParameterGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range dbParameterGroups {

		r, err := initializeResource(a, *i.DBParameterGroupName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func dbSubnetGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	dbSubnetGroups, err := a.awsr.GetDBSubnetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range dbSubnetGroups {

		r, err := initializeResource(a, *i.DBSubnetGroupName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func s3Buckets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	buckets, err := a.awsr.ListBuckets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, v := range buckets {
		r, err := initializeResource(a, *v.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontDistributions(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	distributions, err := a.awsr.GetCloudFrontDistributions(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range distributions {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontOriginAccessIdentities(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	identitys, err := a.awsr.GetCloudFrontOriginAccessIdentities(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range identitys {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudfrontPublicKeys(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	publicKeys, err := a.awsr.GetCloudFrontPublicKeys(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range publicKeys {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func cloudwatchMetricAlarms(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	alarms, err := a.awsr.GetMetricAlarms(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range alarms {
		r, err := initializeResource(a, *i.AlarmName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func iamAccessKeys(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// Get the users list
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)

	for _, un := range userNames {
		// get access keys from a user
		iamAccessKeys, err := a.awsr.GetAccessKeys(ctx, &iam.ListAccessKeysInput{UserName: awsSDK.String(un)})
		if err != nil {
			return nil, err
		}

		for _, i := range iamAccessKeys {
			r, err := initializeResource(a, *i.AccessKeyId, resourceType)
			if err != nil {
				return nil, err
			}
			err = r.Data().Set("user", i.UserName)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamAccountAliases(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	accountAliases, err := a.awsr.GetAccountAliases(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range accountAliases {
		r, err := initializeResource(a, *i, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamAccountPasswordPolicy(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// As it's for the full account we'll tell TF to fetch it directly with a "" id
	r, err := initializeResource(a, "iam-account-password-policy", resourceType)
	if err != nil {
		return nil, err
	}
	return []provider.Resource{r}, nil
}

func iamGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	groups, err := a.awsr.GetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range groups {
		r, err := initializeResource(a, *i.GroupName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamGroupMemberships(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)

	for _, i := range groupNames {
		input := &iam.GetGroupInput{
			GroupName: awsSDK.String(i),
		}

		// Check if group have users. If not do not keep it
		users, err := a.awsr.GetGroupUsers(ctx, input)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			continue
		}

		r, err := initializeResource(a, i, resourceType)
		if err != nil {
			return nil, err
		}

		// TODO this resource is not importable. Define our own ResourceImporter
		// Should be removed when terraform will support it https://github.com/terraform-providers/terraform-provider-aws/pull/13795
		// more detail: https://github.com/cycloidio/terracognita/issues/120
		importer := &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				groupName := d.Id()

				d.Set("group", groupName)
				d.SetId(resource.UniqueId())

				return []*schema.ResourceData{d}, nil
			},
		}

		r.SetImporter(importer)

		resources = append(resources, r)
	}

	return resources, nil
}

func iamGroupPolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), filters)
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

		for _, i := range groupPolicies {
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

func iamGroupPolicyAttachments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	groupNames, err := getIAMGroupNames(ctx, a, IAMGroup.String(), filters)
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

		for _, i := range groupPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", gn, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamInstanceProfiles(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	instanceProfiles, err := a.awsr.GetInstanceProfiles(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range instanceProfiles {
		r, err := initializeResource(a, *i.InstanceProfileName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamOpenidConnectProviders(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	openIDConnectProviders, err := a.awsr.GetOpenIDConnectProviders(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range openIDConnectProviders {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamPolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	input := &iam.ListPoliciesInput{
		Scope: awsSDK.String("Local"),
	}
	policies, err := a.awsr.GetPolicies(ctx, input)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range policies {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamRoles(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	roles, err := a.awsr.GetRoles(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range roles {
		r, err := initializeResource(a, *i.RoleName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamRolePolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	roleNames, err := getIAMRoleNames(ctx, a, IAMRole.String(), filters)
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

		for _, i := range rolePolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", rn, *i), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamRolePolicyAttachments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	roleNames, err := getIAMRoleNames(ctx, a, IAMRole.String(), filters)
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

		for _, i := range rolePolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", rn, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamSAMLProviders(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	samalProviders, err := a.awsr.GetSAMLProviders(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range samalProviders {
		r, err := initializeResource(a, *i.Arn, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamServerCertificates(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	serverCertificates, err := a.awsr.GetServerCertificates(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range serverCertificates {
		r, err := initializeResource(a, *i.ServerCertificateName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUsers(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	users, err := a.awsr.GetUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range users {
		r, err := initializeResource(a, *i.UserName, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUserGroupMemberships(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// if both aws_iam_group_membership and aws_iam_user_group_membership defined, keep only aws_iam_group_membership
	if filters.IsIncluded("aws_iam_group_membership") && (!filters.IsExcluded("aws_iam_group_membership")) {
		return nil, nil
	}

	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), filters)
	if err != nil {
		return nil, err
	}
	resources := make([]provider.Resource, 0)

	for _, un := range userNames {
		var input = &iam.ListGroupsForUserInput{
			UserName: awsSDK.String(un),
		}

		groups, err := a.awsr.GetGroupsForUser(ctx, input)
		if err != nil {
			return nil, err
		}

		// If the user has no Groups then we do not need to write membership
		if len(groups) == 0 {
			continue
		}

		groupNames := make([]string, 0, len(groups))
		for _, g := range groups {
			groupNames = append(groupNames, *g.GroupName)
		}

		// The format expected by TF is <user-name>/<group-name1>/...
		r, err := initializeResource(a, strings.Join(append([]string{un}, groupNames...), "/"), resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func iamUserPolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), filters)
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

		for _, i := range userPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", un, *i), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamUserPolicyAttachments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), filters)
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

		for _, i := range userPolicies {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", un, *i.PolicyArn), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func iamUserSSHKeys(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	// Get the users list
	userNames, err := getIAMUserNames(ctx, a, IAMUser.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)

	for _, un := range userNames {
		// get ssh pub Keys from a user
		sshPublicKeys, err := a.awsr.GetSSHPublicKeys(ctx, &iam.ListSSHPublicKeysInput{UserName: awsSDK.String(un)})
		if err != nil {
			return nil, err
		}

		for _, i := range sshPublicKeys {

			r, err := initializeResource(a, fmt.Sprintf("%s:%s:%s", *i.UserName, *i.SSHPublicKeyId, "SSH"), resourceType)
			if err != nil {
				return nil, err
			}

			resources = append(resources, r)
		}
	}

	return resources, nil
}

func route53DelegationSets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	r53DelegationSets, err := a.awsr.GetReusableDelegationSets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53DelegationSets {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53HealthChecks(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	r53HealthChecks, err := a.awsr.GetHealthChecks(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53HealthChecks {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53QueryLogs(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	r53QueryLogs, err := a.awsr.GetQueryLoggingConfigs(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53QueryLogs {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53Zones(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	r53Zones, err := a.awsr.GetHostedZones(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53Zones {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53Records(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	zones, err := getRoute53ZoneIDs(ctx, a, Route53Zone.String(), filters)
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

		for _, i := range r53Records {
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

func route53ZoneAssociations(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	zones, err := getRoute53ZoneIDs(ctx, a, Route53Zone.String(), filters)
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

		for _, i := range r53ZoneAssociations {
			r, err := initializeResource(a, fmt.Sprintf("%s:%s", z, *i.VPCId), resourceType)
			if err != nil {
				return nil, err
			}
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func route53ResolverEndpoints(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	r53ResolverEndpoints, err := a.awsr.GetResolverEndpoints(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53ResolverEndpoints {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func route53ResolverRuleAssociation(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	r53ResolverRuleAssociations, err := a.awsr.GetResolverRuleAssociations(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range r53ResolverRuleAssociations {
		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesActiveReceiptRuleSets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0, 1)
	if sesActiveReceiptRuleSets == nil {
		return resources, nil
	}

	r, err := initializeResource(a, *sesActiveReceiptRuleSets, resourceType)
	if err != nil {
		return nil, err
	}
	resources = append(resources, r)

	return resources, nil
}

func sesDomainIdentities(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesDomainIdentities, err := a.awsr.GetIdentities(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesDomainIdentities {
		r, err := initializeResource(a, *i, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesDomainGeneral(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	domainNames, err := getSESDomainIdentityDomains(ctx, a, SESDomainIdentity.String(), filters)
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

func sesReceiptFilters(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesReceiptFilters, err := a.awsr.GetReceiptFilters(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesReceiptFilters {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesReceiptRules(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	arrmetadata, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRulesSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesActiveReceiptRuleSets {
		r, err := initializeResource(a, fmt.Sprintf("%s:%s", *arrmetadata, *i.Name), resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesReceiptRuleSets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesActiveReceiptRuleSets, err := a.awsr.GetActiveReceiptRuleSet(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0, 1)
	if sesActiveReceiptRuleSets == nil {
		return resources, nil
	}

	r, err := initializeResource(a, *sesActiveReceiptRuleSets, resourceType)
	if err != nil {
		return nil, err
	}

	resources = append(resources, r)
	return resources, nil
}

func sesConfigurationSets(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesConfigurationSets, err := a.awsr.GetConfigurationSets(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesConfigurationSets {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func sesIdentityNotificationTopics(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	domainNames, err := getSESDomainIdentityDomains(ctx, a, SESDomainIdentity.String(), filters)
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

		for _, i := range sesIdentityNotificationTopics {
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

func sesTemplates(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	sesTemplates, err := a.awsr.GetTemplates(ctx, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range sesTemplates {
		r, err := initializeResource(a, *i.Name, resourceType)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func lambdaFunctions(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	lambdaFunctions, err := a.awsr.GetLambdaFunctions(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range lambdaFunctions {

		r, err := initializeResource(a, *i.FunctionName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func launchConfigurations(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	launchConfigurations, err := a.awsr.GetLaunchConfigurations(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range launchConfigurations {

		r, err := initializeResource(a, *i.LaunchConfigurationName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func launchTemplates(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	var input = &ec2.DescribeLaunchTemplatesInput{
		Filters: toEC2Filters(filters),
	}

	launchTemplates, err := a.awsr.GetLaunchTemplates(ctx, input)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range launchTemplates {

		r, err := initializeResource(a, *i.LaunchTemplateId, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func apiGatewayDeployments(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	apiGatewayRestApis, err := getAPIGatewayRestApis(ctx, a, APIGatewayRestAPI.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, rapi := range apiGatewayRestApis {

		var input = &apigateway.GetDeploymentsInput{
			RestApiId: awsSDK.String(rapi),
		}

		apiGatewayDeployments, err := a.awsr.GetAPIGatewayDeployments(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range apiGatewayDeployments {

			r, err := initializeResource(a, fmt.Sprintf("%s:%s", *i.Id, rapi), resourceType)
			if err != nil {
				return nil, err
			}

			// TODO this resource is not importable. Define our own ResourceImporter
			// Should be removed when terraform will support it
			// more detail: https://github.com/cycloidio/terracognita/issues/120
			importer := &schema.ResourceImporter{
				State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
					parts := strings.SplitN(d.Id(), ":", 2)

					if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
						return nil, fmt.Errorf("unexpected format of ID (%s), expected targetId_port_TargetGroupArn", d.Id())
					}

					d.Set("rest_api_id", parts[1])
					d.SetId(parts[0])

					return []*schema.ResourceData{d}, nil
				},
			}

			r.SetImporter(importer)

			resources = append(resources, r)
		}
	}
	return resources, nil
}

func apiGatewayStages(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	apiGatewayRestApis, err := getAPIGatewayRestApis(ctx, a, APIGatewayRestAPI.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, rapi := range apiGatewayRestApis {

		var input = &apigateway.GetStagesInput{
			RestApiId: awsSDK.String(rapi),
		}

		apiGatewayStages, err := a.awsr.GetAPIGatewayStages(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range apiGatewayStages {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", rapi, *i.StageName), resourceType)
			if err != nil {
				return nil, err
			}

			resources = append(resources, r)
		}
	}
	return resources, nil
}

func apiGatewayResources(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {

	apiGatewayRestApis, err := getAPIGatewayRestApis(ctx, a, APIGatewayRestAPI.String(), filters)
	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, rapi := range apiGatewayRestApis {

		var input = &apigateway.GetResourcesInput{
			RestApiId: awsSDK.String(rapi),
		}

		apiGatewayResources, err := a.awsr.GetAPIGatewayResources(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, i := range apiGatewayResources {
			r, err := initializeResource(a, fmt.Sprintf("%s/%s", rapi, *i.Id), resourceType)
			if err != nil {
				return nil, err
			}

			resources = append(resources, r)
		}
	}
	return resources, nil
}

func apiGatewayRestApis(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	apiGatewayRestApis, err := a.awsr.GetAPIGatewayRestAPIs(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range apiGatewayRestApis {

		r, err := initializeResource(a, *i.Id, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func autoscalingGroups(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	autoscalingGroups, err := a.awsr.GetAutoScalingGroups(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range autoscalingGroups {

		r, err := initializeResource(a, *i.AutoScalingGroupName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func autoscalingPolicies(ctx context.Context, a *aws, resourceType string, filters *filter.Filter) ([]provider.Resource, error) {
	autoscalingPolicies, err := a.awsr.GetAutoScalingPolicies(ctx, nil)

	if err != nil {
		return nil, err
	}

	resources := make([]provider.Resource, 0)
	for _, i := range autoscalingPolicies {

		r, err := initializeResource(a, *i.AutoScalingGroupName+"/"+*i.PolicyName, resourceType)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func toEC2Filters(filters *filter.Filter) []*ec2.Filter {
	tags := filters.Tags
	if len(tags) == 0 {
		return nil
	}
	filtersEc2 := make([]*ec2.Filter, 0, len(tags))

	for _, t := range tags {
		filtersEc2 = append(filtersEc2, t.ToEC2Filter())
	}

	return filtersEc2
}
