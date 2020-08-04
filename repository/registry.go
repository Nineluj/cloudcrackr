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

const RepoName = "cloudcrackr"

func CreateRepository(sess *session.Session) error {
	client := ecr.New(sess)

	result, err := client.CreateRepository(&ecr.CreateRepositoryInput{
		EncryptionConfiguration: nil,
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(false),
		},
		ImageTagMutability: nil,
		RepositoryName:     aws.String(RepoName),
		Tags: []*ecr.Tag{
			{
				Key:   aws.String("service"),
				Value: aws.String("cloudcrackr"),
			},
		},
	})

	_ = result

	return err
}

func createCrackingImage(sess *session.Session) {
	//client := ecr.New(sess)

}