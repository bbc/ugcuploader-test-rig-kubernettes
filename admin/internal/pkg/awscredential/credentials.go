package awscredentials

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	log "github.com/sirupsen/logrus"

	"io/ioutil"
)

type Credentials struct {
}

func (cred *Credentials) GetWebIdentityCredentials() *sts.Credentials {

	webTokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	if webTokenFile != "" {

		content, err := ioutil.ReadFile(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		svc := sts.New(session.New())
		input := &sts.AssumeRoleWithWebIdentityInput{
			RoleArn:          aws.String(os.Getenv("AWS_ROLE_ARN")),
			RoleSessionName:  aws.String("app1"),
			WebIdentityToken: aws.String(string(content)),
		}

		result, err := svc.AssumeRoleWithWebIdentity(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case sts.ErrCodeMalformedPolicyDocumentException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeMalformedPolicyDocumentException)
				case sts.ErrCodePackedPolicyTooLargeException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodePackedPolicyTooLargeException)
				case sts.ErrCodeIDPRejectedClaimException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeIDPRejectedClaimException)
				case sts.ErrCodeIDPCommunicationErrorException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeIDPCommunicationErrorException)
				case sts.ErrCodeInvalidIdentityTokenException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeInvalidIdentityTokenException)
				case sts.ErrCodeExpiredTokenException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeExpiredTokenException)
				case sts.ErrCodeRegionDisabledException:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Errorf(sts.ErrCodeRegionDisabledException)
				default:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Error("WebToken error not sure the reason why")
				}
			} else {
				log.WithFields(log.Fields{
					"err": aerr.Error(),
				}).Errorf("WebToken problem not sure what the reason is")
			}
			return nil
		}
		return result.Credentials
	}
	return nil
}
