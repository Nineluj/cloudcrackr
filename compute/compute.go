// Package for handling compute for cloudcrackr. Uses EC2 over fargate since
// while Fargate provides easier management, it does not allow for file system mounting at this moment
package compute

import (
	"cloudcrackr/constants"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
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

func DeployContainer(sess *session.Session, imageURI string, useGpu bool) error {
	client := ecs.New(sess)

	taskArn, err := registerTask(client, imageURI, useGpu)
	if err != nil {
		return err
	}

	err = runTask(client, taskArn)
	if err != nil {
		return err
	}

	//err = deregisterTask(client, taskArn)
	//if err != nil {
	//	return err
	//}

	return nil
}

func runTask(client *ecs.ECS, taskArn string) error {
	result, err := client.RunTask(&ecs.RunTaskInput{
		Cluster:              aws.String("default"),
		Count:                aws.Int64(1),
		EnableECSManagedTags: aws.Bool(true),
		LaunchType:           aws.String(ecs.LaunchTypeEc2),
		NetworkConfiguration: nil,
		Overrides:            nil,
		PlacementConstraints: nil,
		PlacementStrategy:    nil,
		PropagateTags:        aws.String(ecs.PropagateTagsTaskDefinition),
		ReferenceId:          nil, // TODO: what is this?
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

func registerTask(client *ecs.ECS, imageURI string, useGpu bool) (string, error) {
	// TODO: set this properly
	imageName := "TODO"

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
		// Needed for proper IAM usage
		ExecutionRoleArn: nil,
		Family:           aws.String("ccr-" + imageName),
		// Bridge should be fine
		NetworkMode:          aws.String(ecs.NetworkModeBridge),
		PidMode:              nil,
		PlacementConstraints: nil,
		Tags:                 getTags(),
		TaskRoleArn:          nil,

		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				// Main params
				Command:    nil,
				EntryPoint: nil,

				//
				Cpu:                   nil,
				DisableNetworking:     nil,
				DockerSecurityOptions: nil,
				Environment:           nil,
				EnvironmentFiles:      nil,
				// Marks this container as essential and will cease task when it stops
				Essential: aws.Bool(true),
				// TODO: add logs
				FirelensConfiguration: nil,
				HealthCheck:           nil,
				Hostname:              nil,
				Image:                 aws.String(imageURI),

				Memory:            aws.Int64(MemoryHardLimit),
				MemoryReservation: aws.Int64(MemorySoftLimit),
				// Probably not useful for storage gateway
				MountPoints: nil,
				Name:        nil,
				// Likely not needed
				PortMappings: nil,
				// Might be needed in some cases
				Privileged: nil,

				// Might be required for bash file
				PseudoTerminal: aws.Bool(true),
				Interactive:    nil,

				ReadonlyRootFilesystem: aws.Bool(false),
				RepositoryCredentials:  nil,
				// Define GPU usage here
				ResourceRequirements: resourceReqs,
				// Might be needed to pass info used to mount Storage Gateway
				Secrets: nil,
				// Could tweak kernel parameters to speed up container speeds
				SystemControls: nil,
				User:           nil,
				// Maybe I can plug in Storage Gateway here?
				VolumesFrom: nil,
				// Equivalent of docker run's workdir
				WorkingDirectory: nil,
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
