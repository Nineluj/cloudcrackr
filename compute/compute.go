// Package for handling compute for cloudcrackr
package compute

import (
	"cloudcrackr/constants"
	"cloudcrackr/iam"
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
	// Memory allocated for the instance
	memoryLimit = "0.5GB"
	// vCPUs allocated for the instance
	allocatedVCpus = ".25 vCPU"
)

var (
	// Settings that need to be enabled to allow tag forwarding to work properly
	EnabledSettings = [...]string{
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

	return deployId + "-" + parts[0][1:], parts[0][1:], nil
}

func CreateCluster(sess *session.Session, clusterName string) error {
	// TODO: seems to create IAM role?
	client := ecs.New(sess)

	// Settings need to be set to allow tag propagation with ECS.
	// This only needs to be done once per account per region (once for cloudcrackr)
	for _, setting := range EnabledSettings {
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

func DeployContainer(sess *session.Session, clusterName, imageURI, bucketName, dictionary, hash string, useGpu bool) error {
	client := ecs.New(sess)

	deployId, imageName, err := getDeployId(imageURI)
	if err != nil {
		return err
	}

	// ...
	// Extract last part of Image URI

	// Create environment variables for task to bootstrap running
	envVars := getEnvVars(bucketName, dictionary, hash, ProcPrefix+deployId)

	// Get the IAM role arn for the task
	ecsTaskRoleArn, err := iam.GetECSRoleArn(sess)
	if err != nil {
		return err
	}

	//
	taskArn, err := registerTask(client, ecsTaskRoleArn, imageURI, deployId, imageName, envVars, useGpu)
	if err != nil {
		return err
	}

	err = runTask(client, clusterName, taskArn, deployId)
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

	// These could also be written to a file and passed using EnvironmentFile
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
			Value: aws.String(base + output),
		},
	}
}

func runTask(client *ecs.ECS, clusterName, taskArn, deployId string) error {
	result, err := client.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(clusterName),
		TaskDefinition: aws.String(taskArn),
		Count:          aws.Int64(1),
		LaunchType:     aws.String(ecs.LaunchTypeEc2),
		ReferenceId:    aws.String(deployId),

		// Not doing anything with the VPC but need to set this up in order
		// to run a Fargate container
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpDisabled),
				SecurityGroups: nil,
				Subnets:        []*string{aws.String("127.0.0.1/32")},
			},
		},

		// Tag related
		PropagateTags:        aws.String(ecs.PropagateTagsTaskDefinition),
		EnableECSManagedTags: aws.Bool(true),
		Tags:                 getTags(),
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

func registerTask(client *ecs.ECS, ecsTaskRoleArn, imageURI, deployId, imageName string, envVars []*ecs.KeyValuePair, useGpu bool) (string, error) {
	//var resourceReqs []*ecs.ResourceRequirement
	//if useGpu {
	//	resourceReqs = []*ecs.ResourceRequirement{
	//		{
	//			Type:  aws.String("GPU"),
	//			Value: aws.String("1"),
	//		},
	//	}
	//}

	result, err := client.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
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
				// -- Main params
				Image: aws.String(imageURI),
				Name:  aws.String(deployId),
				// We use these environment variables to tell the host where
				// it can download the files from on S3 and where it should
				// upload the results after it's done
				Environment: envVars,

				// Define GPU usage here
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
