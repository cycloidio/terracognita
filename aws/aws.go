package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/hcl/printer"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/chr4/pwgen"
	"github.com/cycloidio/raws"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
	"github.com/hashicorp/terraform/helper/schema"
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

type funcReader func(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error

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
func Import(ctx context.Context, accessKey, secretKey, region string, tags []Tag, include, exclude []string, out io.Writer) error {
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

	mapCfg := map[string]interface{}{
		"resource": make(map[string]map[string]interface{}),
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
		err = fn(ctx, awsClient, awsr, region, tags, mapCfg)
		if err != nil {
			return fmt.Errorf("could not import the resource %q because: %s", r, err)
		}
	}

	b, err := json.Marshal(mapCfg)
	if err != nil {
		return err
	}

	f, err := hcl.ParseBytes(b)
	if err != nil {
		return fmt.Errorf("error while 'hcl.ParseBytes': %s", err)
	}

	var buff bytes.Buffer
	err = printer.Fprint(&buff, f.Node)
	if err != nil {
		return fmt.Errorf("error while pretty printing HCL: %s", err)
	}

	err = fmtcmd.Run(nil, nil, &buff, out, fmtcmd.Options{})
	if err != nil {
		return fmt.Errorf("error while fmt HCL: %s", err)
	}

	return nil
}

func awsInstance(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_instance"] = make(map[string]interface{})

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

	for _, id := range instanceIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_instance"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_instance"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsVpc(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_vpc"] = make(map[string]interface{})

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

	for _, id := range vpcsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_vpc"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_vpc"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsAmi(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_ami"] = make(map[string]interface{})

	var input = &ec2.DescribeImagesInput{
		Filters: toEC2Filters(tags),
	}

	images, err := awsr.GetImages(ctx, input)
	if err != nil {
		return err
	}

	imagesIDs := make([]string, 0)
	for _, v := range images[region].Images {
		imagesIDs = append(imagesIDs, *v.ImageId)
	}

	for _, id := range imagesIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_ami"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_ami"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsSecurityGroup(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_security_group"] = make(map[string]interface{})

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

	for _, id := range sgsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_security_group"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_security_group"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsSubnet(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_subnet"] = make(map[string]interface{})

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

	for _, id := range subnetsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_subnet"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_subnet"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsEbsVolume(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_ebs_volume"] = make(map[string]interface{})

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

	for _, id := range volumesIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_ebs_volume"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_ebs_volume"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsEbsSnapshot(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_ebs_snapshot"] = make(map[string]interface{})

	var input = &ec2.DescribeSnapshotsInput{
		Filters: toEC2Filters(tags),
	}

	snapshots, err := awsr.GetSnapshots(ctx, input)
	if err != nil {
		return err
	}

	snapshotsIDs := make([]string, 0)
	for _, v := range snapshots[region].Snapshots {
		snapshotsIDs = append(snapshotsIDs, *v.SnapshotId)
	}

	for _, id := range snapshotsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_ebs_snapshot"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		res["resource"].(map[string]map[string]interface{})["aws_ebs_snapshot"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsElasticacheCluster(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_elasticache_cluster"] = make(map[string]interface{})

	cacheClusters, err := awsr.GetElastiCacheCluster(ctx, nil)
	if err != nil {
		return err
	}

	cacheClustersIDs := make([]string, 0)
	for _, v := range cacheClusters[region].CacheClusters {
		cacheClustersIDs = append(cacheClustersIDs, *v.CacheClusterId)
	}

	for _, id := range cacheClustersIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_elasticache_cluster"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		res["resource"].(map[string]map[string]interface{})["aws_elasticache_cluster"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsElb(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_elb"] = make(map[string]interface{})

	lbs, err := awsr.GetLoadBalancers(ctx, nil)
	if err != nil {
		return err
	}

	lbsIDs := make([]string, 0)
	for _, v := range lbs[region].LoadBalancerDescriptions {
		lbsIDs = append(lbsIDs, *v.LoadBalancerName)
	}

	for _, id := range lbsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_elb"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		res["resource"].(map[string]map[string]interface{})["aws_elb"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsAlb(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_alb"] = make(map[string]interface{})

	lbs, err := awsr.GetLoadBalancersV2(ctx, nil)
	if err != nil {
		return err
	}

	lbsIDs := make([]string, 0)
	for _, v := range lbs[region].LoadBalancers {
		lbsIDs = append(lbsIDs, *v.LoadBalancerArn)
	}

	for _, id := range lbsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_alb"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		res["resource"].(map[string]map[string]interface{})["aws_alb"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsDbInstance(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_db_instance"] = make(map[string]interface{})

	dbs, err := awsr.GetDBInstances(ctx, nil)
	if err != nil {
		return err
	}

	dbsIDs := make([]string, 0)
	for _, v := range dbs[region].DBInstances {
		dbsIDs = append(dbsIDs, *v.DBInstanceIdentifier)
	}

	for _, id := range dbsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_db_instance"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		res["resource"].(map[string]map[string]interface{})["aws_db_instance"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

func awsS3Bucket(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, region string, tags []Tag, res map[string]interface{}) error {
	res["resource"].(map[string]map[string]interface{})["aws_s3_bucket"] = make(map[string]interface{})

	buckets, err := awsr.ListBuckets(ctx, nil)
	if err != nil {
		return err
	}

	bucketsIDs := make([]string, 0)
	for _, v := range buckets[region].Buckets {
		bucketsIDs = append(bucketsIDs, *v.Name)
	}

	for _, id := range bucketsIDs {
		p := tfaws.Provider().(*schema.Provider)
		resource := p.ResourcesMap["aws_s3_bucket"]
		srd := resource.Data(nil)
		srd.SetId(id)

		err = resource.Read(srd, tfAWSClient)
		if err != nil {
			return err
		}

		for _, t := range tags {
			if srd.Get(fmt.Sprintf("tags.%s", t.Name)).(string) != t.Value {
				continue
			}
		}

		res["resource"].(map[string]map[string]interface{})["aws_s3_bucket"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
	}

	return nil
}

//func aws_s3_bucket_object(ctx context.Context, tfAWSClient interface{}, awsr raws.AWSReader, res map[string]interface{}) error {
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

//res["resource"].(map[string]map[string]interface{})["aws_s3_bucket_object"][pwgen.Alpha(5)] = mergeFullConfig(srd, resource.Schema, "")
//}
//}

//return nil
//}

// mergeFullConfig creates the key to the map and if it had a value before set it, if
func mergeFullConfig(cfgr *schema.ResourceData, sch map[string]*schema.Schema, key string) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range sch {
		kk := key
		if key != "" {
			kk = key + "." + k
		} else {
			kk = k
		}

		// schema.Resource means that it has nested fields
		if sr, ok := v.Elem.(*schema.Resource); ok {
			if v.Type == schema.TypeSet || v.Type == schema.TypeList {
				ar, ok := res[k]
				if !ok {
					ar = make([]interface{}, 0, 0)
				}

				list, ok := cfgr.GetOk(kk)
				if !ok {
					continue
				}
				if list != nil {
					// For the types that are a list, we have to set them in an array, and also
					// add the correct index for the number of setts (entries on the original config)
					// that there are on the provided configuration
					switch val := list.(type) {
					case []map[string]interface{}:
						for i := range val {
							ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, fmt.Sprintf("%s.%d", kk, i)))
						}
					case []interface{}:
						for i := range val {
							ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, fmt.Sprintf("%s.%d", kk, i)))
						}
					}
				} else {
					ar = append(ar.([]interface{}), mergeFullConfig(cfgr, sr.Schema, kk))
				}
				res[k] = ar
			} else {
				res[k] = mergeFullConfig(cfgr, sr.Schema, kk)
			}
			continue
		}

		vv := cfgr.Get(kk)
		if vv == nil {
			continue
		}

		if s, ok := vv.(*schema.Set); ok {
			res[k] = s.List()
		} else {
			res[k] = vv
		}
	}
	return res
}
