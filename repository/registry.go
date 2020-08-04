/*
Steps for handling images:
* Create repository with New() function here
*
*/
package repository

import (
	"encoding/base64"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"io/ioutil"
	"strings"
)

const RepoName = "cloudcrackr"

func CreateRepository(sess *session.Session) error {
	client := ecr.New(sess)

	_, err := client.CreateRepository(&ecr.CreateRepositoryInput{
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

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// Since we want this op to be idempotent we don't care about this error
			case ecr.ErrCodeRepositoryAlreadyExistsException:
				return nil
			}
		}
	}

	return err
}

func CreateImage(sess *session.Session) error {
	return errors.New("not implemented")
	client := ecr.New(sess)

	// This process is done using docker commands
	// Extract the credentials that we can give to docker to push
	result, err := client.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}

	authToken := result.AuthorizationData[0].AuthorizationToken
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(*authToken))

	credentials, err := ioutil.ReadAll(decoder)
	if err != nil {
		return err
	}

	credentialsParts := strings.Split(string(credentials), ":")
	username := credentialsParts[0]
	password := credentialsParts[1]

	_ = username
	_ = password

	//exec.Command()

	//client.PutImage(&ecr.PutImageInput{
	//	ImageDigest:            nil,
	//	ImageManifest:          nil,
	//	ImageManifestMediaType: nil,
	//	ImageTag:               nil,
	//	RegistryId:             nil,
	//	RepositoryName:         nil,
	//})

	return nil
}
