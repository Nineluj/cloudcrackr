package iam

import (
	"cloudcrackr/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	ECSRoleName              = "ecsTaskExecutionRole"
	ECSManagedPolicy         = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
	AssumeRolePolicyDocument = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
)

func getTags() []*iam.Tag {
	return []*iam.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

func GetECSRoleArn(sess *session.Session) (string, error) {
	client := iam.New(sess)
	result, err := client.GetRole(&iam.GetRoleInput{RoleName: aws.String(ECSRoleName)})

	if err != nil {
		return "", err
	}

	return *result.Role.Arn, nil
}

// TODO: Create and use an IAM role
func SetupIAM(sess *session.Session, path string) error {
	client := iam.New(sess)

	_, err := client.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
		Path:                     aws.String(path),
		RoleName:                 aws.String(ECSRoleName),
		Tags:                     getTags(),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// Ignore this to make the function idempotent
			case iam.ErrCodeEntityAlreadyExistsException:
				break
			default:
				return err
			}
		} else {
			return err
		}
	}

	_, err = client.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(ECSManagedPolicy),
		RoleName:  aws.String(ECSRoleName),
	})

	return err
}
