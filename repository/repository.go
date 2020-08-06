// Provides functions for interacting with the Container Registry
package repository

import (
	"cloudcrackr/constants"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
	log "github.com/visionmedia/go-cli-log"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	TagLookupTimeout = 10
)

func getTags() []*ecr.Tag {
	return []*ecr.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

// Initiates a new repository on AWS that cloudcrackr can use
func createRepository(sess *session.Session, name string) error {
	client := ecr.New(sess)

	_, err := client.CreateRepository(&ecr.CreateRepositoryInput{
		EncryptionConfiguration: nil,
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(false),
		},
		ImageTagMutability: nil,
		RepositoryName:     aws.String(name),
		Tags:               getTags(),
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

// Verifies the presence of the imageName and returns its URI
func GetImageURI(sess *session.Session, imageName string) (string, error) {
	client := ecr.New(sess)
	result, err := client.DescribeImages(&ecr.DescribeImagesInput{
		RepositoryName: aws.String(imageName),
	})

	if err != nil {
		return "", err
	}

	if len(result.ImageDetails) == 0 {
		return "", errors.New("couldn't find imageName")
	}

	domain, _, err := getECRDetails(sess)

	if err != nil {
		return "", err
	}

	imgRef := fmt.Sprintf("%v/%v:latest", domain, imageName)

	return imgRef, nil
}

func collector(allRepos []*ecr.Repository, nTagCheckers int, recv <-chan tagCheckedRepository) (taggedRepoNames []string) {
	// map for repos so that once we get responses we can find them in O(1) time
	repoMap := make(map[string]*ecr.Repository)
	for _, repo := range allRepos {
		repoMap[*repo.RepositoryArn] = repo
	}

	nResponses := 0

ListenLoop:
	for {
		select {
		case res := <-recv:
			nResponses++

			if res.hasTag {
				taggedRepoNames = append(taggedRepoNames, *repoMap[res.arn].RepositoryName)
			}

			if nResponses == nTagCheckers {
				return
			}
		case <-time.After(time.Second * TagLookupTimeout):
			log.Warn("Gave up on finding more repositories, time out")
			break ListenLoop
		}
	}

	return
}

type tagCheckedRepository struct {
	arn    string
	hasTag bool
}

func tagChecker(client *ecr.ECR, arn string, resp chan<- tagCheckedRepository) {
	result, err := client.ListTagsForResource(&ecr.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	})

	if err != nil {
		return
	}

	for _, tag := range result.Tags {
		if *tag.Key == constants.TagKey && *tag.Value == constants.TagValue {
			resp <- tagCheckedRepository{
				arn:    arn,
				hasTag: true,
			}

			return
		}
	}

	resp <- tagCheckedRepository{
		arn:    arn,
		hasTag: false,
	}
}

func ListImages(sess *session.Session) ([]string, error) {
	client := ecr.New(sess)

	// TODO: should use result and nextToken to check if there were >100 results
	// another call should be made
	result, err := client.DescribeRepositories(&ecr.DescribeRepositoriesInput{})

	if err != nil {
		return nil, err
	}

	repoChan := make(chan tagCheckedRepository)
	nTagCheckers := len(result.Repositories)
	for _, repo := range result.Repositories {
		go tagChecker(client, *repo.RepositoryArn, repoChan)
	}

	repos := collector(result.Repositories, nTagCheckers, repoChan)
	close(repoChan)

	return repos, nil
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

func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	jsonBuf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(jsonBuf), nil
}

// TODO: Would like to use the below code but only Stdout works at the moment
type StatusMessage struct {
	Status          string `json:"status"`
	ProgressDetails struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	ProgressBar string `json:"progress"`
	Id          string `json:"id"`
}

type prettyStatusWriter struct{}

func (psw *prettyStatusWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return
	}

	var status StatusMessage

	err = json.Unmarshal(p, &status)

	if err != nil {
		return n, err
	}

	if status.Id != "" {
		log.Info(status.Status+" - "+status.Id,
			fmt.Sprintf("%v/%v [%v%%]",
				status.ProgressDetails.Current, status.ProgressDetails.Total, 0))
	} else {
		log.Info("Upload", status.Status)
	}

	return
}

func pushImage(client *dclient.Client, username, password, imageRef string) error {
	registryAuth, err := encodeAuthToBase64(types.AuthConfig{
		Username: username,
		Password: password,
	})

	if err != nil {
		return err
	}

	// TODO: check tags?
	pushOptions := types.ImagePushOptions{
		RegistryAuth: registryAuth,
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
	} else {
		return err
	}

	_, err = io.Copy(os.Stdin, readCloser)

	return err
}

func PushImage(sess *session.Session, imageId, imageName string) error {
	domain, credentials, err := getECRDetails(sess)
	if err != nil {
		return err
	}

	client, err := dclient.NewClientWithOpts()
	if err != nil {
		return err
	}

	username, password, err := parseCredentials(&credentials)
	if err != nil {
		return err
	}

	err = createRepository(sess, imageName)
	if err != nil {
		return err
	}

	imgRef := fmt.Sprintf("%v/%v:latest", domain, imageName)
	// We need to tag the image with the repo tag before pushing
	err = tagImage(client, imageId, imgRef)
	if err != nil {
		return err
	}

	err = pushImage(client, username, password, imgRef)
	if err != nil {
		return err
	}

	log.Info("Image", "Image successfully pushed to %v", imgRef)
	return nil
}

func tagImage(client *dclient.Client, imageId, imgRef string) error {
	return client.ImageTag(context.Background(), imageId, imgRef)
}
