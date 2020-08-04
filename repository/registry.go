package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
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

func ListImages(sess *session.Session) ([]string, error) {
	client := ecr.New(sess)

	result, err := client.ListImages(&ecr.ListImagesInput{
		// TODO: use specific registryID in case of >1 present?
		RegistryId:     nil,
		RepositoryName: aws.String(RepoName),
	})

	if err != nil {
		return nil, err
	}

	var imageList []string

	for _, img := range result.ImageIds {
		imageList = append(imageList, img.String())
	}

	return imageList, nil
}

// Don't need this afaik since docker takes base64 string directly
func parseCredentials(creds *string) (string, string, error) {
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(*creds))

	credentials, err := ioutil.ReadAll(decoder)
	if err != nil {
		return "", "", err
	}

	credentialsParts := strings.Split(string(credentials), ":")
	username := credentialsParts[0]
	password := credentialsParts[1]

	return username, password, nil
}

func getECRCredentials(sess *session.Session) (string, error) {
	client := ecr.New(sess)

	// This process is done using docker commands
	// Extract the credentials that we can give to docker to push
	result, err := client.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	authToken := result.AuthorizationData[0].AuthorizationToken

	return *authToken, nil
}

func pushImage(client *dclient.Client, credentials string, imageRef string) error {
	// logic around how to use docker's client.ImagePush()

	// TODO: check tags?
	pushOptions := types.ImagePushOptions{
		RegistryAuth: credentials,
		// Not sure about these two?
		All: false,
		PrivilegeFunc: func() (string, error) {
			fmt.Println("something happened")
			return "", errors.New("my fail")
		},
	}

	readCloser, err := client.ImagePush(context.Background(), imageRef, pushOptions)
	defer func() {
		_ = readCloser.Close()
	}()

	return err
}

func PushImage(sess *session.Session, imgRef string) error {
	credentials, err := getECRCredentials(sess)
	if err != nil {
		return err
	}

	client, err := dclient.NewEnvClient()
	if err != nil {
		return err
	}

	return pushImage(client, credentials, imgRef)
}
