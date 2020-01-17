package aws

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awscredentials "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/awscredential"
	log "github.com/sirupsen/logrus"
)

type S3Operations struct {
}

//GetBucketItems gets items specified by the key from the bucket
func (s3ops *S3Operations) GetBucketItems(bucket string, prefix string, index int) (items []string, probs bool) {

	sess, _ := session.NewSession()
	var config = &aws.Config{}
	if os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
		eksCreds := awscredentials.Credentials{}
		creds := *eksCreds.GetWebIdentityCredentials()
		config = &aws.Config{
			Region:      aws.String("eu-west-2"),
			Credentials: credentials.NewStaticCredentials(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken),
		}
	} else {
		config = &aws.Config{
			Region: aws.String("eu-west-2"),
		}
	}

	// Create a client from just a config.r
	client := s3.New(sess, config)

	log.WithFields(log.Fields{
		"Bucket": bucket,
		"Prefix": prefix,
	}).Info("------ FETCHING BUCKET DETAILS ----")
	params := &s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String(prefix)}
	tenants := make(map[string]struct{})
	// Example iterating over at most 3 pages of a ListObjectsV2 operation.
	err := client.ListObjectsV2Pages(params,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, content := range page.Contents {
				parts := strings.Split(*content.Key, "/")
				// you can use the ,ok idiom to check for existing keys
				if _, ok := tenants[parts[index]]; !ok {
					tenants[parts[index]] = struct{}{}
				}
			}
			return lastPage != true
		})
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err.Error(),
			"Bucket": bucket,
			"Prefix": prefix,
		}).Error("Problems Getting Bucket Details")
		probs = true
	} else {

		keys := make([]string, 0, len(tenants))
		for k := range tenants {
			keys = append(keys, k)
		}
		items = keys
		probs = false
	}
	return
}
