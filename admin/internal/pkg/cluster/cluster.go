package cluster

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscredentials "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/awscredential"
	log "github.com/sirupsen/logrus"
)

//Operations performed on the cluster
type Operations struct{}

//DescribeCluster returns a description of the cluster
func (ops Operations) DescribeCluster(clusterName string) (awsRegion string, awsActNmbr string, problems string) {

	webTokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	var cfg aws.Config
	var err error
	if webTokenFile != "" {
		eksCreds := awscredentials.Credentials{}
		creds := eksCreds.GetWebIdentityCredentials()
		keyID := *creds.AccessKeyId
		secretKey := *creds.SecretAccessKey
		st := *creds.SessionToken
		cfg, err = external.LoadDefaultAWSConfig(
			// Hard coded credentials.
			external.WithCredentialsValue(aws.Credentials{
				AccessKeyID: keyID, SecretAccessKey: secretKey, SessionToken: st,
				Source: "Describe Cluster",
			}),
		)
	} else {
		cfg, err = external.LoadDefaultAWSConfig()
	}
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems Loading Credentials")
	}
	cfg.Region = "eu-west-1"
	svc := eks.New(cfg)
	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	req := svc.DescribeClusterRequest(input)
	result, err := req.Send(context.Background())
	if err != nil {
		problems = err.Error()
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

	fmt.Println(clsArn)
	arns := strings.Split(clsArn, ":")
	awsRegion = arns[3]
	awsActNmbr = arns[4]
	return
}
