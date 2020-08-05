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
	log "github.com/visionmedia/go-cli-log"
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

// Gets the authorization token and endpoint for ECR
func getECRDetails(sess *session.Session) (string, string, error) {
	client := ecr.New(sess)

	// This process is done using docker commands
	// Extract the credentials that we can give to docker to push
	result, err := client.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", err
	}

	if len(result.AuthorizationData) > 1 {
		// TODO: look into this
		return "", "", errors.New("received more than one authorization credentials")
	} else if len(result.AuthorizationData) == 0 {
		return "", "", errors.New("couldn't retrieve credentials for ECR")
	}

	endpoint := *result.AuthorizationData[0].ProxyEndpoint
	endpointTrimmed := strings.TrimPrefix(endpoint, "https://")
	if len(endpoint) == len(endpointTrimmed) {
		return "", "", errors.New("expected ECR endpoint to contain https prefix")
	}

	authToken := result.AuthorizationData[0].AuthorizationToken

	return endpointTrimmed, *authToken, nil
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

	if err == nil {
		defer func() {
			_ = readCloser.Close()
		}()
	}

	return err
}

func PushImage(sess *session.Session, imageId, imageName string) error {
	domain, credentials, err := getECRDetails(sess)
	if err != nil {
		return err
	}

	client, err := dclient.NewEnvClient()
	if err != nil {
		return err
	}

	// alternative: use client.RegistryLogin() for auth?

	imgRef := fmt.Sprintf("%v/%v:latest", domain, imageName)
	// We need to tag the image with the repo tag before pushing
	err = tagImage(client, imageId, imgRef)
	if err != nil {
		return err
	}

	err = pushImage(client, credentials, imgRef)
	if err != nil {
		return err
	}

	log.Info("Image", "Image successfully pushed to %v", imgRef)
	return nil
}

func tagImage(client *dclient.Client, imageId, imgRef string) error {
	return client.ImageTag(context.Background(), imageId, imgRef)
}
