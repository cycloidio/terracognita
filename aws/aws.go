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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cycloidio/raws"
	"github.com/cycloidio/terraforming/util"
	"github.com/cycloidio/terraforming/util/writer"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, instanceIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, vpcsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, imagesIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, sgsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, subnetsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, volumesIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, snapshotsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, cacheClustersIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, lbsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, lbsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, dbsIDs, w)
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

	err = util.ReadIDsAndWrite(tfAWSClient, Provider, resourceType, tags, state, bucketsIDs, w)
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
