// Package for handling compute for cloudcrackr. Uses EC2 over fargate since
// while Fargate provides easier management, it does not allow for file system mounting at this moment
package compute

import (
	"cloudcrackr/constants"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"net/url"
	"strings"
	"time"
)

const (
	ProcPrefix = "proc/"
)

const (
	// Amount of memory at which the system will try to minimize additional usage
	MemorySoftLimit = 256
	// Amount of memory at which the system will shut off the instance
	MemoryHardLimit = 512
)

func getTags() []*ecs.Tag {
	return []*ecs.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

func getDeployId(imageURI string) (string, string, error) {
	uri, err := url.Parse("https://" + imageURI)
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(uri.Path, ":", 2)

	// Get the date+time to create a unique identifier for this deployment
	t := time.Now()
	deployId := fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return deployId + parts[0], parts[0], nil
}

func createCluster() {
	// TODO: seems to create IAM role?
	// client.CreateCluster()
}

func DeployContainer(sess *session.Session, imageURI, bucketName, dictionary, hash string, useGpu bool) error {
	client := ecs.New(sess)

	deployId, imageName, err := getDeployId(imageURI)
	if err != nil {
		return err
	}

	// ...
	// Extract last part of Image URI

	// Create environment variables for task to bootstrap running
	envVars := getEnvVars(bucketName, dictionary, hash, ProcPrefix+deployId+"/")

	//
	taskArn, err := registerTask(client, imageURI, deployId, imageName, envVars, useGpu)
	if err != nil {
		return err
	}

	err = runTask(client, taskArn, deployId)
	if err != nil {
		return err
	}

	//err = deregisterTask(client, taskArn)
	//if err != nil {
	//	return err
	//}

	return nil
}

func getEnvVars(bucketName, dictionary, hash, output string) []*ecs.KeyValuePair {
	base := "s3://" + bucketName + "/"

	return []*ecs.KeyValuePair{
		{
			Name:  aws.String("CCR_DICTIONARY"),
			Value: aws.String(base + dictionary),
		},
		{
			Name:  aws.String("CCR_HASH"),
			Value: aws.String(base + hash),
		},
		{
			Name:  aws.String("CCR_OUTPUT"),
			Value: aws.String(base + output + "output"),
		},
	}
}

func runTask(client *ecs.ECS, taskArn, deployId string) error {
	result, err := client.RunTask(&ecs.RunTaskInput{
		Cluster:              aws.String("default"),
		Count:                aws.Int64(1),
		EnableECSManagedTags: aws.Bool(true),
		LaunchType:           aws.String(ecs.LaunchTypeEc2),
		PropagateTags:        aws.String(ecs.PropagateTagsTaskDefinition),
		ReferenceId:          aws.String(deployId),
		Tags:                 getTags(),
		TaskDefinition:       aws.String(taskArn),
	})

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

func registerTask(client *ecs.ECS, imageURI, deployId, imageName string, envVars []*ecs.KeyValuePair, useGpu bool) (string, error) {
	var resourceReqs []*ecs.ResourceRequirement
	if useGpu {
		resourceReqs = []*ecs.ResourceRequirement{
			{
				Type:  aws.String("GPU"),
				Value: aws.String("1"),
			},
		}
	}

	result, err := client.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		RequiresCompatibilities: []*string{aws.String(ecs.CompatibilityEc2)},
		Family:                  aws.String(imageName),

		Tags: getTags(),

		// Needed for proper IAM usage
		ExecutionRoleArn: nil,
		TaskRoleArn:      nil,

		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				// -- Main params
				Image: aws.String(imageURI),
				Name:  aws.String(deployId),
				// We use these environment variables to tell the host where
				// it can download the files from on S3 and where it should
				// upload the results after it's done
				Environment: envVars,

				// -- Resources for container
				Cpu:               aws.Int64(10),
				Memory:            aws.Int64(MemoryHardLimit),
				MemoryReservation: aws.Int64(MemorySoftLimit),
				// Define GPU usage here
				ResourceRequirements: resourceReqs,
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
	})

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
		return errors.New("wrong task was de-registered")
	}

	return nil
}
