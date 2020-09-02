// Package for handling compute for cloudcrackr
package compute

import (
	"cloudcrackr/auth"
	"cloudcrackr/constants"
	"cloudcrackr/network"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/visionmedia/go-cli-log"
)

const (
	ProcPrefix = "proc/"
)

const (
	// Memory allocated for the instance
	memoryLimit = "0.5GB"
	// vCPUs allocated for the instance
	allocatedVCpus = ".25 vCPU"
)

var (
	// ErrorClusterNotFound when ECS cluster couldn't be found
	ErrorClusterNotFound = errors.New("couldn't find cluster")
	// ErrorWrongTaskDeregistration when task deregistration ID doesn't match the requested deletee's task ID
	ErrorWrongTaskDeregistration = errors.New("wrong task was de-registered")
)

var (
	// Settings that need to be enabled to allow tag forwarding to work properly
	enabledSettings = [...]string{
		ecs.SettingNameServiceLongArnFormat,
		ecs.SettingNameTaskLongArnFormat,
		ecs.SettingNameContainerInstanceLongArnFormat,
	}
)

func getTags() []*ecs.Tag {
	return []*ecs.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

func getDeployID(imageURI string) (string, string, error) {
	uri, err := url.Parse("https://" + imageURI)
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(uri.Path, ":", 2)

	// Get the date+time to create a unique identifier for this deployment
	t := time.Now()
	deployID := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return deployID + "-" + parts[0][1:], parts[0][1:], nil
}

func CreateCluster(sess *session.Session, clusterName string) error {
	client := ecs.New(sess)

	// Settings need to be set to allow tag propagation with ECS.
	// This only needs to be done once per account per region (once for cloudcrackr)
	for _, setting := range enabledSettings {
		_, err := client.PutAccountSetting(&ecs.PutAccountSettingInput{
			Name:  aws.String(setting),
			Value: aws.String("enabled"),
		})

		if err != nil {
			return err
		}
	}

	_, err := client.CreateCluster(&ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		Tags:        getTags(),
	})

	return err
}

func DeleteCluster(sess *session.Session, clusterName string) error {
	client := ecs.New(sess)

	clusterArn, err := getClusterArn(client, clusterName)
	if err != nil {
		return err
	}

	_, err = client.DeleteCluster(&ecs.DeleteClusterInput{Cluster: aws.String(clusterArn)})
	return err
}

func getClusterArn(client *ecs.ECS, clusterName string) (string, error) {
	result, err := client.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	})

	if err != nil {
		return "", err
	}

	if len(result.Clusters) == 0 {
		return "", ErrorClusterNotFound
	}

	return *result.Clusters[0].ClusterArn, nil
}

// DeployContainer deploys the image onto an ECS managed container
func DeployContainer(sess *session.Session, clusterName, imageURI, bucketName, dictionary, hash string,
	useGpu bool) error {
	client := ecs.New(sess)

	deployID, imageName, err := getDeployID(imageURI)
	if err != nil {
		return err
	}

	// Get the ECS cluster ARN, this also validates that it exists
	clusterArn, err := getClusterArn(client, clusterName)
	if err != nil {
		return err
	}

	// Get the default subnet to use
	subnetArn, err := network.GetDefaultSubnetArn(sess)
	if err != nil {
		return err
	}

	// Get the s3 locations for the execute script
	s3Targets := getS3Targets(bucketName, dictionary, hash, ProcPrefix+deployID)

	// Get the IAM role arn for the task
	ecsTaskRoleArn, err := auth.GetECSRoleArn(sess)
	if err != nil {
		return err
	}

	// Get credentials with limited privileges for AWS cli on the container
	credentials, err := auth.GetCrackrCredentials(sess,
		s3Targets.dictionaryPath, s3Targets.hashPath, s3Targets.outputPath)
	if err != nil {
		return err
	}

	taskArn, err := registerTask(client, ecsTaskRoleArn, imageURI, deployID, imageName, s3Targets, credentials, useGpu)
	if err != nil {
		return err
	}

	err = runTask(client, clusterArn, taskArn, subnetArn, deployID)
	if err != nil {
		return err
	}

	log.Info("Cracking", "Started instance for cracking with image %v", imageName)

	return nil
}

type S3Targets struct {
	dictionaryPath string
	hashPath       string
	outputPath     string
}

func getS3Targets(bucketName, dictionary, hash, output string) *S3Targets {
	base := "s3://" + bucketName + "/"

	// These could also be written to a file and passed using EnvironmentFile
	return &S3Targets{
		base + dictionary,
		base + hash,
		base + output,
	}
}

func runTask(client *ecs.ECS, clusterArn, taskArn, subnetArn, deployID string) error {
	subnetName := strings.SplitN(subnetArn, "/", 2)

	input := &ecs.RunTaskInput{
		// Should actually work with the short name? Check this
		Cluster: aws.String(clusterArn),

		TaskDefinition: aws.String(taskArn),
		Count:          aws.Int64(1),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		ReferenceId:    aws.String(deployID),

		// Not doing anything with the VPC but need to set this up in order
		// to run a Fargate container
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				//SecurityGroups: nil,
				Subnets: []*string{aws.String(subnetName[1])},
			},
		},

		// Tag related
		PropagateTags:        aws.String(ecs.PropagateTagsTaskDefinition),
		EnableECSManagedTags: aws.Bool(true),
		Tags:                 getTags(),
	}

	result, err := client.RunTask(input)

	if err != nil {
		return err
	}

	if len(result.Failures) > 0 {
		for _, fail := range result.Failures {
			return errors.New(*fail.Reason)
		}
	}

	return nil
}

func registerTask(client *ecs.ECS, ecsTaskRoleArn, imageURI, deployID, imageName string,
	s3Target *S3Targets, credentials *auth.Credentials, _ bool) (string, error) {
	// TODO: Pass region
	commandArguments := []*string{
		// first three arguments for s3
		&s3Target.dictionaryPath, &s3Target.hashPath, &s3Target.outputPath,
		// next three are AWS credentials so that the container can use the aws CLI
		// with limited permissions
		&credentials.AccessKeyID, &credentials.SecretAccessKey, &credentials.SessionToken,
		// region to use
		client.Config.Region,
	}

	input := &ecs.RegisterTaskDefinitionInput{
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityFargate)},
		Family:                  aws.String(imageName),

		Tags: getTags(),

		// Needed for proper IAM usage
		ExecutionRoleArn: aws.String(ecsTaskRoleArn),
		TaskRoleArn:      nil,

		// -- Resources for task
		Cpu:         aws.String(allocatedVCpus),
		Memory:      aws.String(memoryLimit),
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),

		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Command: commandArguments,

				// -- Main params
				Image: aws.String(imageURI),
				Name:  aws.String(deployID),

				// (not implemented yet) Define GPU usage here
				//ResourceRequirements: resourceReqs,
				// Could tweak kernel parameters to speed up container speeds
				SystemControls: nil,
				User:           nil,

				// -- Misc
				// Marks this container as essential and will cease task when it stops
				Essential: aws.Bool(true),
				// Might be needed in some cases
				Privileged: nil,
				// Might be required for bash file
				PseudoTerminal: aws.Bool(true),
				Interactive:    nil,
			},
		},
	}

	result, err := client.RegisterTaskDefinition(input)

	if err != nil {
		return "", err
	}

	return *result.TaskDefinition.TaskDefinitionArn, nil
}

func deregisterTask(client *ecs.ECS, taskArn string) error {
	result, err := client.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(taskArn),
	})

	if err != nil {
		return err
	}

	if taskArn != *result.TaskDefinition.TaskDefinitionArn {
		return ErrorWrongTaskDeregistration
	}

	return nil
}
