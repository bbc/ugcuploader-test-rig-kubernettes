package cluster

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscredentials "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/awscredential"
)

//Operations performed on the cluster
type Operations struct{}

//DescribeCluster returns a description of the cluster
func (ops Operations) DescribeCluster(clusterName string) (awsRegion string, awsActNmbr string) {

	eksCreds := awscredentials.Credentials{}
	creds := *eksCreds.GetWebIdentityCredentials()
	cfg, err := external.LoadDefaultAWSConfig(
		// Hard coded credentials.
		external.WithCredentialsValue(aws.Credentials{
			AccessKeyID: *creds.AccessKeyId, SecretAccessKey: *creds.SecretAccessKey, SessionToken: *creds.SessionToken,
			Source: "Describe Cluster",
		}),
	)
	cfg.Region = "eu-west-2"
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems Loading Credentials")
	}

	svc := eks.New(cfg)
	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	req := svc.DescribeClusterRequest(input)
	result, err := req.Send(context.Background())
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Error(eks.ErrCodeResourceNotFoundException)
			case eks.ErrCodeException:
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Error(eks.ErrCodeException)
			case eks.ErrCodeServerException:
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Error(eks.ErrCodeServerException)
			case eks.ErrCodeServiceUnavailableException:
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Error(eks.ErrCodeServiceUnavailableException)
			default:
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Error("Not sure what cause DescribeCluster to fail")
			}
		} else {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Describe cluster failed")
		}
		return
	}

	clsArn := *result.Cluster.Arn

	arns := strings.Split(clsArn, ":")
	awsRegion = arns[3]
	awsActNmbr = arns[4]
	return
}
