package reader

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
	"github.com/aws/aws-sdk-go/service/athena/athenaiface"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/batch/batchiface"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/configservice/configserviceiface"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice/databasemigrationserviceiface"
	"github.com/aws/aws-sdk-go/service/dax/daxiface"
	"github.com/aws/aws-sdk-go/service/directconnect/directconnectiface"
	"github.com/aws/aws-sdk-go/service/directoryservice/directoryserviceiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk/elasticbeanstalkiface"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice/elasticsearchserviceiface"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/emr/emriface"
	"github.com/aws/aws-sdk-go/service/fsx/fsxiface"
	"github.com/aws/aws-sdk-go/service/glue/glueiface"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/lightsail/lightsailiface"
	"github.com/aws/aws-sdk-go/service/mediastore/mediastoreiface"
	"github.com/aws/aws-sdk-go/service/mq/mqiface"
	"github.com/aws/aws-sdk-go/service/neptune/neptuneiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/redshift/redshiftiface"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/route53resolver/route53resolveriface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/aws/aws-sdk-go/service/storagegateway/storagegatewayiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

//go:generate go run ../cmd/ -output reader.go

// New returns an object which also contains the accountID and the region to use.
//
// The accountID is helpful to return only the AMI or snapshots that belong to the account.
//
// While the region has to be a valid AWS region
//
// An error is returned if any of the needed AWS request for creating the reader returns an AWS error, in such case it
// will have any of the common error codes (see below) or EmptyStaticCreds code or a go standard error in case that no
// regions are matched with the ones available, at the time, in AWS.
// See:
//  * https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html#CommonErrors
//  * https://docs.aws.amazon.com/STS/latest/APIReference/CommonErrors.html
func New(ctx context.Context, accessKey, secretKey, region, sessionToken string, config *aws.Config) (Reader, error) {
	var c = connector{}

	creds, ec2s, sts, err := configureAWS(accessKey, secretKey, region, sessionToken)
	if err != nil {
		return nil, err
	}
	c.creds = creds
	if err := c.setAccountID(ctx, sts); err != nil {
		return nil, err
	}

	if err = c.setRegion(ctx, ec2s, region); err != nil {
		return nil, err
	}

	c.setService(config)

	return &c, nil
}

// The connector provides easy access to AWS SDK calls.
//
// By using it, calls can be made directly through multiple regions, and will filter only data that belongs to you.
// For example, when fetching the list of AMI, or snapshots.
//
// In order to start making calls, only calling New is required.
type connector struct {
	region    string
	svc       *serviceConnector
	creds     *credentials.Credentials
	accountID *string
}

func (c *connector) GetAccountID() string {
	return *c.accountID
}

func (c *connector) GetRegion() string {
	return c.region
}

type serviceConnector struct {
	apigateway               apigatewayiface.APIGatewayAPI
	athena                   athenaiface.AthenaAPI
	autoscaling              autoscalingiface.AutoScalingAPI
	batch                    batchiface.BatchAPI
	cloudfront               cloudfrontiface.CloudFrontAPI
	cloudwatch               cloudwatchiface.CloudWatchAPI
	configservice            configserviceiface.ConfigServiceAPI
	databasemigrationservice databasemigrationserviceiface.DatabaseMigrationServiceAPI
	dax                      daxiface.DAXAPI
	directconnect            directconnectiface.DirectConnectAPI
	directoryservice         directoryserviceiface.DirectoryServiceAPI
	dynamodb                 dynamodbiface.DynamoDBAPI
	ec2                      ec2iface.EC2API
	ecs                      ecsiface.ECSAPI
	efs                      efsiface.EFSAPI
	eks                      eksiface.EKSAPI
	elasticache              elasticacheiface.ElastiCacheAPI
	elasticbeanstalk         elasticbeanstalkiface.ElasticBeanstalkAPI
	elasticsearchservice     elasticsearchserviceiface.ElasticsearchServiceAPI
	elb                      elbiface.ELBAPI
	elbv2                    elbv2iface.ELBV2API
	emr                      emriface.EMRAPI
	fsx                      fsxiface.FSxAPI
	glue                     glueiface.GlueAPI
	iam                      iamiface.IAMAPI
	kinesis                  kinesisiface.KinesisAPI
	lambda                   lambdaiface.LambdaAPI
	lightsail                lightsailiface.LightsailAPI
	mediastore               mediastoreiface.MediaStoreAPI
	mq                       mqiface.MQAPI
	neptune                  neptuneiface.NeptuneAPI
	rds                      rdsiface.RDSAPI
	redshift                 redshiftiface.RedshiftAPI
	region                   string
	route53resolver          route53resolveriface.Route53ResolverAPI
	route53                  route53iface.Route53API
	s3downloader             s3manageriface.DownloaderAPI
	s3                       s3iface.S3API
	ses                      sesiface.SESAPI
	session                  *session.Session
	sqs                      sqsiface.SQSAPI
	storagegateway           storagegatewayiface.StorageGatewayAPI
}

/* The default region is only used to (1) get the list of region and
 * (2) get the account ID associated with the credentials.
 *
 * It is not used as a default region for services, therefore if no
 * region is specified when instantiating the connector, then it will
 * not try to establish any connections with AWS services.
 */
const defaultRegion string = "eu-west-1"

// configureAWS creates a new static credential with the passed accessKey and
// secretKey and with it, a sessions which is used to create a EC2 client and
// a Security Token Service client.
// The only AWS error code that this function return is
// * EmptyStaticCreds
func configureAWS(accessKey, secretKey, region, token string) (*credentials.Credentials, ec2iface.EC2API, stsiface.STSAPI, error) {
	if region == "" {
		region = defaultRegion
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, token)
	_, err := creds.Get()
	if err != nil {
		return nil, nil, nil, err
	}
	sess := session.Must(
		session.NewSession(&aws.Config{
			Region:      aws.String(region),
			DisableSSL:  aws.Bool(false),
			MaxRetries:  aws.Int(3),
			Credentials: creds,
		}),
	)
	return creds, ec2.New(sess), sts.New(sess), nil
}

// setAccountID retrieves the caller ID from the Security Token Service and set
// it in the connector.
// An AWS error can be returned with one of the common error codes.
// See https://docs.aws.amazon.com/STS/latest/APIReference/CommonErrors.html
func (c *connector) setAccountID(ctx context.Context, sts stsiface.STSAPI) error {
	resp, err := sts.GetCallerIdentityWithContext(ctx, nil)
	if err != nil {
		return err
	}
	c.accountID = resp.Account
	return nil
}

// setRegion retrieves the AWS available regions and matches with the passed
// region.
// A AWS error can be returned with one of the common error codes or a standard
// go error if enabledRegions is empty or if 0 AWS regions has been matched.
// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html#CommonErrors
func (c *connector) setRegion(ctx context.Context, ec2 ec2iface.EC2API, region string) error {
	if region == "" {
		return errors.New("at least one region name is required")
	}

	regions, err := ec2.DescribeRegionsWithContext(ctx, nil)
	if err != nil {
		return err
	}

	for _, r := range regions.Regions {
		if region == *r.RegionName {
			c.region = region
			return nil
		}
	}

	if c.region == "" {
		return fmt.Errorf("found 0 regions matching: %v", region)
	}

	return nil
}

func (c *connector) setService(config *aws.Config) {
	if config != nil {
		config.Credentials = c.creds
	} else {
		config = &aws.Config{
			DisableSSL:  aws.Bool(false),
			MaxRetries:  aws.Int(3),
			Credentials: c.creds,
		}
	}

	config.Region = aws.String(c.region)
	sess := session.Must(session.NewSession(config))
	svc := &serviceConnector{
		region:  c.region,
		session: sess,
	}
	c.svc = svc
}
