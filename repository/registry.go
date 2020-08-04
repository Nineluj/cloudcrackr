/*
Steps for handling images:
* Create repository with New() function here
*
*/
package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

const RepositoryName = "cloudcrackr"

func CreateRepository(sess *session.Session) error {
	client := ecr.New(sess)

	result, err := client.CreateRepository(&ecr.CreateRepositoryInput{
		EncryptionConfiguration: nil,
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(false),
		},
		ImageTagMutability: nil,
		RepositoryName:     aws.String(RepositoryName),
		Tags: []*ecr.Tag{
			{
				Key:   aws.String("service"),
				Value: aws.String("cloudcrackr"),
			},
		},
	})

	return err
}

func createCrackingImage(sess *session.Session) {
	client := ecr.New(sess)

}
