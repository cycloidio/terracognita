package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cycloidio/raws"
	"github.com/cycloidio/terraforming/util"
	"github.com/cycloidio/terraforming/util/writer"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

const (
	// Provider it's the name of the provider
	Provider = "aws"
)

type funcReader func(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error

var (
	awsResources = map[string]funcReader{
		"aws_instance":            awsInstance,
		"aws_vpc":                 awsVpc,
		"aws_ami":                 awsAmi,
		"aws_security_group":      awsSecurityGroup,
		"aws_subnet":              awsSubnet,
		"aws_ebs_volume":          awsEbsVolume,
		"aws_ebs_snapshot":        awsEbsSnapshot,
		"aws_elasticache_cluster": awsElasticacheCluster,
		"aws_elb":                 awsElb,
		"aws_alb":                 awsAlb,
		"aws_db_instance":         awsDbInstance,
		"aws_s3_bucket":           awsS3Bucket,
		//"aws_s3_bucket_object": aws_s3_bucket_object,
		"aws_cloudfront_distribution":           awsCloudFrontDistribution,
		"aws_cloudfront_origin_access_identity": awsCloudFrontOriginAccessIdentity,
		"aws_cloudfront_public_key":             awsCloudFrontPublicKey,
		"aws_iam_access_key":                    awsIAMAccessKey,
		"aws_iam_account_alias":                 awsIAMAccountAlias,
		"aws_iam_account_password_policy":       awsIAMAccountPasswordPolicy,
		"aws_iam_group":                         awsIAMGroup,
		"aws_iam_group_membership":              awsIAMGroupMembership,
		"aws_iam_group_policy":                  awsIAMGroupPolicy,
		"aws_iam_group_policy_attachment":       awsIAMGroupPolicyAttachment,
		"aws_iam_instance_profile":              awsIAMInstanceProfile,
		"aws_iam_openid_connect_provider":       awsIAMOpenIDConnectProvicer,
		"aws_iam_policy":                        awsIAMPolicy,
		// As it's deprecated we'll not support it
		//"aws_iam_policy_attachment"
		"aws_iam_role":                   awsIAMRole,
		"aws_iam_role_policy":            awsIAMRolePolicy,
		"aws_iam_role_policy_attachment": awsIAMRolePolicyAttachment,
		"aws_iam_saml_provider":          awsIAMSAMLProvider,
		"aws_iam_server_certificate":     awsIAMServerCertificate,
		// TODO: Don't know how to get it from AWS SKD
		//"aws_iam_service_linked_role"
		"aws_iam_user":                  awsIAMUser,
		"aws_iam_user_group_membership": awsIAMUserGroupMembership,
		// Can not be Read
		// aws_iam_user_login_profile
		"aws_iam_user_policy":            awsIAMUserPolicy,
		"aws_iam_user_policy_attachment": awsIAMUserPolicyAttachment,
		// TODO: Requires to many mandatory fields
		// let's see if we can do it later
		//"aws_iam_user_ssh_key":           awsIAMUserSSHKey,
	}
)

// Import imports from AWS the resources filtered by tags
// and include|exclude and writes the result HCL to out
func Import(ctx context.Context, accessKey, secretKey, region string, tags []util.Tag, include, exclude []string, state bool, out io.Writer) error {
	log.SetFlags(0)
	awsr, err := raws.NewAWSReader(ctx, accessKey, secretKey, []string{region}, nil)
	if err != nil {
		return fmt.Errorf("could not initialize 'raws' because: %s", err)
	}

	cfg := tfaws.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}

	awsClient, err := cfg.Client()
	if err != nil {
		return fmt.Errorf("could not initialize 'terraform/aws.Config.Client()' because: %s", err)
	}

	mapInclude := make(map[string]struct{})
	for _, r := range include {
		mapInclude[r] = struct{}{}
	}
	mapExclude := make(map[string]struct{})
	for _, r := range exclude {
		mapExclude[r] = struct{}{}
	}

	var wr writer.Writer
	if state {
		wr = writer.NewTFStateWriter()
	} else {
		wr = writer.NewHCLWriter()
	}

	for r, fn := range awsResources {
		if len(include) != 0 {
			if _, ok := mapInclude[r]; !ok {
				continue
			}
		}
		if len(exclude) != 0 {
			if _, ok := mapExclude[r]; ok {
				continue
			}
		}
		err = fn(ctx, awsClient, awsr, r, region, tags, state, wr)
		if err != nil {
			return fmt.Errorf("could not import the resource %q because: %s", r, err)
		}
	}

	if state {
		// Write to root because then the NewState is called
		// it creates by default a 'root' one and then on the
		// AddModuleState we replace that empty module for this one
		ms := &terraform.ModuleState{
			Path: []string{"root"},
		}

		tfw := wr.(*writer.TFStateWriter)
		ms.Resources = tfw.Config

		state := terraform.NewState()
		state.AddModuleState(ms)

		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		err := enc.Encode(state)
		if err != nil {
			return fmt.Errorf("could not encode state due to: %s", err)
		}
	} else {
		hclw := wr.(*writer.HCLWriter)
		b, err := json.Marshal(hclw.Config)
		if err != nil {
			return err
		}

		f, err := hcl.ParseBytes(b)
		if err != nil {
			return fmt.Errorf("error while 'hcl.ParseBytes': %s", err)
		}

		buff := &bytes.Buffer{}
		err = printer.Fprint(buff, f.Node)
		if err != nil {
			return fmt.Errorf("error while pretty printing HCL: %s", err)
		}

		buff = bytes.NewBuffer(util.FormatHCL(buff.Bytes()))

		err = fmtcmd.Run(nil, nil, buff, out, fmtcmd.Options{})
		if err != nil {
			return fmt.Errorf("error while fmt HCL: %s", err)
		}
	}
	return nil
}

func awsInstance(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeInstancesInput{
		Filters: toEC2Filters(tags),
	}

	instances, err := awsr.GetInstances(ctx, input)
	if err != nil {
		return err
	}

	instanceIDs := make([]string, 0)
	for _, v := range instances[region].Reservations {
		for _, vv := range v.Instances {
			instanceIDs = append(instanceIDs, *vv.InstanceId)
		}
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, instanceIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsVpc(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeVpcsInput{
		Filters: toEC2Filters(tags),
	}

	vpcs, err := awsr.GetVpcs(ctx, input)
	if err != nil {
		return err
	}

	vpcsIDs := make([]string, 0)
	for _, v := range vpcs[region].Vpcs {
		vpcsIDs = append(vpcsIDs, *v.VpcId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, vpcsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsAmi(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeImagesInput{
		Filters: toEC2Filters(tags),
	}

	images, err := awsr.GetOwnImages(ctx, input)
	if err != nil {
		return err
	}

	imagesIDs := make([]string, 0)
	for _, v := range images[region].Images {
		imagesIDs = append(imagesIDs, *v.ImageId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, imagesIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsSecurityGroup(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeSecurityGroupsInput{
		Filters: toEC2Filters(tags),
	}

	sgs, err := awsr.GetSecurityGroups(ctx, input)
	if err != nil {
		return err
	}

	sgsIDs := make([]string, 0)
	for _, v := range sgs[region].SecurityGroups {
		sgsIDs = append(sgsIDs, *v.GroupId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, sgsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsSubnet(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeSubnetsInput{
		Filters: toEC2Filters(tags),
	}

	subnets, err := awsr.GetSubnets(ctx, input)
	if err != nil {
		return err
	}

	subnetsIDs := make([]string, 0)
	for _, v := range subnets[region].Subnets {
		subnetsIDs = append(subnetsIDs, *v.SubnetId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, subnetsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsEbsVolume(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeVolumesInput{
		Filters: toEC2Filters(tags),
	}

	volumes, err := awsr.GetVolumes(ctx, input)
	if err != nil {
		return err
	}

	volumesIDs := make([]string, 0)
	for _, v := range volumes[region].Volumes {
		volumesIDs = append(volumesIDs, *v.VolumeId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, volumesIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsEbsSnapshot(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	var input = &ec2.DescribeSnapshotsInput{
		Filters: toEC2Filters(tags),
	}

	snapshots, err := awsr.GetOwnSnapshots(ctx, input)
	if err != nil {
		return err
	}

	snapshotsIDs := make([]string, 0)
	for _, v := range snapshots[region].Snapshots {
		snapshotsIDs = append(snapshotsIDs, *v.SnapshotId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, snapshotsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsElasticacheCluster(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	cacheClusters, err := awsr.GetElastiCacheClusters(ctx, nil)
	if err != nil {
		return err
	}

	cacheClustersIDs := make([]string, 0)
	for _, v := range cacheClusters[region].CacheClusters {
		cacheClustersIDs = append(cacheClustersIDs, *v.CacheClusterId)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, cacheClustersIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsElb(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	lbs, err := awsr.GetLoadBalancers(ctx, nil)
	if err != nil {
		return err
	}

	lbsIDs := make([]string, 0)
	for _, v := range lbs[region].LoadBalancerDescriptions {
		lbsIDs = append(lbsIDs, *v.LoadBalancerName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, lbsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsAlb(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	lbs, err := awsr.GetLoadBalancersV2(ctx, nil)
	if err != nil {
		return err
	}

	lbsIDs := make([]string, 0)
	for _, v := range lbs[region].LoadBalancers {
		lbsIDs = append(lbsIDs, *v.LoadBalancerArn)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, lbsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsDbInstance(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	dbs, err := awsr.GetDBInstances(ctx, nil)
	if err != nil {
		return err
	}

	dbsIDs := make([]string, 0)
	for _, v := range dbs[region].DBInstances {
		dbsIDs = append(dbsIDs, *v.DBInstanceIdentifier)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, dbsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsS3Bucket(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	buckets, err := awsr.ListBuckets(ctx, nil)
	if err != nil {
		return err
	}

	bucketsIDs := make([]string, 0)
	for _, v := range buckets[region].Buckets {
		bucketsIDs = append(bucketsIDs, *v.Name)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, bucketsIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

//func aws_s3_bucket_object(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, state bool, w writer.Writer) error {
//res["resource"].(map[string]map[string]interface{})["aws_s3_bucket_object"] = make(map[string]interface{})

//for _, b := range res["resource"].(map[string]map[string]interface{})["aws_s3_bucket"] {
//bucketName := b.(map[string]interface{})["bucket"].(string)

//objects, err := awsr.ListObjects(ctx, &s3.ListObjectsInput{
//Bucket: aws.String(bucketName),
//})
//if err != nil {
//return err
//}

//objectsIDs := make([]string, 0)
//for _, v := range objects[region].Contents {
//objectsIDs = append(objectsIDs, *v.Key)
//}

//for _, id := range objectsIDs {
//p := tfaws.Provider().(*schema.Provider)
//resource := p.ResourcesMap["aws_s3_bucket_object"]
//srd := resource.Data(nil)
////srd.SetId(id)
//err = srd.Set("bucket", bucketName)
//if err != nil {
//return err
//}
//err = srd.Set("key", id)
//if err != nil {
//return err
//}

//err = resource.Read(srd, tfAWSClient)
//if err != nil {
//return err
//}

//if srd.Get("tags.client").(string) != tagClientValue {
//continue
//}

//res["resource"].(map[string]map[string]interface{})["aws_s3_bucket_object"][pwgen.Alpha(5)] = mergeFullConfig(srd, state bool, resource.Schema, "")
//}
//}

//return nil
//}

func awsCloudFrontDistribution(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	distributions, err := awsr.GetCloudFrontDistributions(ctx, nil)
	if err != nil {
		return err
	}

	distributionIDs := make([]string, 0)
	for _, i := range distributions[region].DistributionList.Items {
		distributionIDs = append(distributionIDs, *i.Id)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, distributionIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsCloudFrontOriginAccessIdentity(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	identitys, err := awsr.GetCloudFrontOriginAccessIdentities(ctx, nil)
	if err != nil {
		return err
	}

	identityIDs := make([]string, 0)
	for _, i := range identitys[region].CloudFrontOriginAccessIdentityList.Items {
		identityIDs = append(identityIDs, *i.Id)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, identityIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsCloudFrontPublicKey(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	publicKeys, err := awsr.GetCloudFrontPublicKeys(ctx, nil)
	if err != nil {
		return err
	}

	publicKeyIDs := make([]string, 0)
	for _, i := range publicKeys[region].PublicKeyList.Items {
		publicKeyIDs = append(publicKeyIDs, *i.Id)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, publicKeyIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMAccessKey(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	// I could get all the users and do the same filtering by user but it
	// seems that it does not need it and returns all the AccessKeys
	accessKeys, err := awsr.GetAccessKeys(ctx, nil)
	if err != nil {
		return err
	}

	accessKeyIDs := make([]string, 0)
	mapAccessKeyIDs := make(map[string]string)
	for _, i := range accessKeys[region].AccessKeyMetadata {
		accessKeyIDs = append(accessKeyIDs, *i.AccessKeyId)
		mapAccessKeyIDs[*i.AccessKeyId] = *i.UserName
	}

	rdfn := func(srd *schema.ResourceData) error {
		if n, ok := mapAccessKeyIDs[srd.Id()]; ok {
			srd.Set("user", n)
		} else {
			return errors.Errorf("invalid id: %s", n)
		}

		return nil
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, accessKeyIDs, rdfn, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMAccountAlias(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	accountAliases, err := awsr.GetAccountAliases(ctx, nil)
	if err != nil {
		return err
	}

	accountAliasIDs := make([]string, 0)
	for _, i := range accountAliases[region].AccountAliases {
		accountAliasIDs = append(accountAliasIDs, *i)
	}

	rdfn := func(srd *schema.ResourceData) error {
		// This resource has no IDs so
		// the account_alias acts as it
		srd.Set("account_alias", srd.Id())
		return nil
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, accountAliasIDs, rdfn, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMAccountPasswordPolicy(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	// As it's for the full account we'll tell TF to fetch it directly with a "" id
	err := util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, []string{""}, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMGroup(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	groups, err := awsr.GetGroups(ctx, nil)
	if err != nil {
		return err
	}

	groupIDs := make([]string, 0)
	for _, i := range groups[region].Groups {
		// Internally TF uses the GroupName as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_group.go#L70
		groupIDs = append(groupIDs, *i.GroupName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, groupIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMGroupMembership(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	groupNames, err := getGroupNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, gn := range groupNames {
		rdfn := func(srd *schema.ResourceData) error {
			srd.Set("group", srd.Id())
			return nil
		}
		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, []string{gn}, rdfn, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}
	return nil
}

func awsIAMGroupPolicy(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	groupNames, err := getGroupNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, gn := range groupNames {
		input := &iam.ListGroupPoliciesInput{
			GroupName: aws.String(gn),
		}
		groupPolicies, err := awsr.GetGroupPolicies(ctx, input)
		if err != nil {
			return err
		}

		groupPolicyIDs := make([]string, 0)
		for _, i := range groupPolicies[region].PolicyNames {
			// It needs the ID to be "GN:PN"
			// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_group_policy.go#L134:6
			groupPolicyIDs = append(groupPolicyIDs, fmt.Sprintf("%s:%s", gn, *i))
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, groupPolicyIDs, nil, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMGroupPolicyAttachment(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	groupNames, err := getGroupNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, gn := range groupNames {
		input := &iam.ListAttachedGroupPoliciesInput{
			GroupName: aws.String(gn),
		}
		groupPolicies, err := awsr.GetAttachedGroupPolicies(ctx, input)
		if err != nil {
			return err
		}

		groupPolicyIDs := make([]string, 0)
		mapGroupPolicyID := make(map[string]string)
		for _, i := range groupPolicies[region].AttachedPolicies {
			groupPolicyIDs = append(groupPolicyIDs, fmt.Sprintf("%s:%s", gn, *i))
			mapGroupPolicyID[fmt.Sprintf("%s:%s", gn, *i)] = *i.PolicyArn
		}
		rdfn := func(srd *schema.ResourceData) error {
			srd.Set("group", gn)
			srd.Set("policy_arn", mapGroupPolicyID[srd.Id()])
			return nil
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, groupPolicyIDs, rdfn, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMPolicy(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	input := &iam.ListPoliciesInput{
		Scope: aws.String("Local"),
	}
	policies, err := awsr.GetPolicies(ctx, input)
	if err != nil {
		return err
	}

	policyIDs := make([]string, 0)
	for _, i := range policies[region].Policies {
		// Internally TF uses the ARN as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_policy.go#L121
		policyIDs = append(policyIDs, *i.Arn)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, policyIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMRole(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	roles, err := awsr.GetRoles(ctx, nil)
	if err != nil {
		return err
	}

	roleIDs := make([]string, 0)
	for _, i := range roles[region].Roles {
		// Internally TF uses the RoleName as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_role.go#L162
		roleIDs = append(roleIDs, *i.RoleName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, roleIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMUser(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	users, err := awsr.GetUsers(ctx, nil)
	if err != nil {
		return err
	}

	userIDs := make([]string, 0)
	for _, i := range users[region].Users {
		// Internally TF uses the RoleName as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_user.go#L86
		userIDs = append(userIDs, *i.UserName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, userIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMUserGroupMembership(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	userNames, err := getUserNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	groupNames, err := getGroupNames(ctx, awsr, region)
	if err != nil {
		return err
	}
	gni := make([]interface{}, len(groupNames))
	for i, gn := range groupNames {
		gni[i] = gn
	}
	groupSet := schema.NewSet(schema.HashString, gni)

	for _, un := range userNames {
		rdfn := func(srd *schema.ResourceData) error {
			srd.Set("user", srd.Id())
			// TF will filter the correct ones
			// and not all of them
			srd.Set("groups", groupSet)
			return nil
		}
		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, []string{un}, rdfn, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}
	return nil
}

func awsIAMInstanceProfile(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	instanceProfiles, err := awsr.GetInstanceProfiles(ctx, nil)
	if err != nil {
		return err
	}

	instanceProfileIDs := make([]string, 0)
	for _, i := range instanceProfiles[region].InstanceProfiles {
		// Internally TF uses the RoleName as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_instance_profile.go#L283
		instanceProfileIDs = append(instanceProfileIDs, *i.InstanceProfileName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, instanceProfileIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMOpenIDConnectProvicer(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	openIDConnectProviders, err := awsr.GetOpenIDConnectProviders(ctx, nil)
	if err != nil {
		return err
	}

	openIDConnectProviderIDs := make([]string, 0)
	for _, i := range openIDConnectProviders[region].OpenIDConnectProviderList {
		// Internally TF uses the ARN as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_openid_connect_provider.go#L283
		openIDConnectProviderIDs = append(openIDConnectProviderIDs, *i.Arn)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, openIDConnectProviderIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMSAMLProvider(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	samalProviders, err := awsr.GetSAMLProviders(ctx, nil)
	if err != nil {
		return err
	}

	samalProviderIDs := make([]string, 0)
	for _, i := range samalProviders[region].SAMLProviderList {
		// Internally TF uses the ARN as the ID
		// https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/resource_aws_iam_saml_provider.go#L71
		samalProviderIDs = append(samalProviderIDs, *i.Arn)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, samalProviderIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

func awsIAMRolePolicy(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	roleNames, err := getRoleNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, rn := range roleNames {
		input := &iam.ListRolePoliciesInput{
			RoleName: aws.String(rn),
		}
		rolePolicies, err := awsr.GetRolePolicies(ctx, input)
		if err != nil {
			return err
		}

		rolePolicyIDs := make([]string, 0)
		for _, i := range rolePolicies[region].PolicyNames {
			// It needs the ID to be "RN:PN"
			rolePolicyIDs = append(rolePolicyIDs, fmt.Sprintf("%s:%s", rn, *i))
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, rolePolicyIDs, nil, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMRolePolicyAttachment(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	roleNames, err := getRoleNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, rn := range roleNames {
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(rn),
		}
		rolePolicies, err := awsr.GetAttachedRolePolicies(ctx, input)
		if err != nil {
			return err
		}

		rolePolicyIDs := make([]string, 0)
		mapRolePolicyID := make(map[string]string)
		for _, i := range rolePolicies[region].AttachedPolicies {
			rolePolicyIDs = append(rolePolicyIDs, fmt.Sprintf("%s:%s", rn, *i))
			mapRolePolicyID[fmt.Sprintf("%s:%s", rn, *i)] = *i.PolicyArn
		}
		rdfn := func(srd *schema.ResourceData) error {
			srd.Set("role", rn)
			srd.Set("policy_arn", mapRolePolicyID[srd.Id()])
			return nil
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, rolePolicyIDs, rdfn, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMUserPolicy(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	userNames, err := getUserNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, un := range userNames {
		input := &iam.ListUserPoliciesInput{
			UserName: aws.String(un),
		}
		userPolicies, err := awsr.GetUserPolicies(ctx, input)
		if err != nil {
			return err
		}

		userPolicyIDs := make([]string, 0)
		for _, i := range userPolicies[region].PolicyNames {
			// It needs the ID to be "RN:PN"
			userPolicyIDs = append(userPolicyIDs, fmt.Sprintf("%s:%s", un, *i))
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, userPolicyIDs, nil, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMUserPolicyAttachment(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	userNames, err := getUserNames(ctx, awsr, region)
	if err != nil {
		return err
	}

	for _, un := range userNames {
		input := &iam.ListAttachedUserPoliciesInput{
			UserName: aws.String(un),
		}
		userPolicies, err := awsr.GetAttachedUserPolicies(ctx, input)
		if err != nil {
			return err
		}

		userPolicyIDs := make([]string, 0)
		mapUserPolicyID := make(map[string]string)
		for _, i := range userPolicies[region].AttachedPolicies {
			userPolicyIDs = append(userPolicyIDs, fmt.Sprintf("%s:%s", un, *i))
			mapUserPolicyID[fmt.Sprintf("%s:%s", un, *i)] = *i.PolicyArn
		}
		rdfn := func(srd *schema.ResourceData) error {
			srd.Set("user", un)
			srd.Set("policy_arn", mapUserPolicyID[srd.Id()])
			return nil
		}

		err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, userPolicyIDs, rdfn, w)
		if err != nil {
			return errors.Wrap(err, "failed to ReadIDsAndWrite")
		}
	}

	return nil
}

func awsIAMServerCertificate(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
	serverCertificates, err := awsr.GetServerCertificates(ctx, nil)
	if err != nil {
		return err
	}

	serverCertificateIDs := make([]string, 0)
	for _, i := range serverCertificates[region].ServerCertificateMetadataList {
		serverCertificateIDs = append(serverCertificateIDs, *i.ServerCertificateName)
	}

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, serverCertificateIDs, nil, w)
	if err != nil {
		return errors.Wrap(err, "failed to ReadIDsAndWrite")
	}

	return nil
}

//func awsIAMUserSSHKey(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, resourceType, region string, tags []util.Tag, state bool, w writer.Writer) error {
//userNames, err := getUserNames(ctx, awsr, region)
//if err != nil {
//return err
//}

//for _, un := range userNames {
//input := &iam.GetSSHPublicKeyInput{
//UserName: aws.String(un),
//Encoding: aws.String("SSH"),
//}

//userSSHKey, err := awsr.GetSSHPublicKey(ctx, input)
//if err != nil {
//return err
//}

//rdfn := func(srd *schema.ResourceData) error {
//srd.Set("username", userSSHKey[region].SSHPublicKey.UserName)
//srd.Set("encoding", "SSH")
//return nil
//}

//err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, []string{*userSSHKey[region].SSHPublicKey.SSHPublicKeyId}, rdfn, w)
//if err != nil {
//return errors.Wrap(err, "failed to ReadIDsAndWrite")
//}
//}

//return nil
//}

func toEC2Filters(tags []util.Tag) []*ec2.Filter {
	if len(tags) == 0 {
		return nil
	}
	filters := make([]*ec2.Filter, 0, len(tags))

	for _, t := range tags {
		filters = append(filters, t.ToEC2Filter())
	}

	return filters
}

func getUserNames(ctx context.Context, awsr raws.AWSReader, region string) ([]string, error) {
	users, err := awsr.GetUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	userNames := make([]string, 0)
	for _, i := range users[region].Users {
		userNames = append(userNames, *i.UserName)
	}

	return userNames, nil
}

func getGroupNames(ctx context.Context, awsr raws.AWSReader, region string) ([]string, error) {
	groups, err := awsr.GetGroups(ctx, nil)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, 0)
	for _, i := range groups[region].Groups {
		groupIDs = append(groupIDs, *i.GroupName)
	}

	return groupIDs, nil
}

func getRoleNames(ctx context.Context, awsr raws.AWSReader, region string) ([]string, error) {
	roles, err := awsr.GetRoles(ctx, nil)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0)
	for _, i := range roles[region].Roles {
		roleIDs = append(roleIDs, *i.RoleName)
	}

	return roleIDs, nil
}
