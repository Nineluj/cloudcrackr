package auth

import (
	"cloudcrackr/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	DefaultAssumeRolePolicyDocument = `{
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

func getIAMTags() []*iam.Tag {
	return []*iam.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

func getUserArn(client *iam.IAM) (string, error) {
	//TODO: document that we use session tokens to create policy that allows the user
	// to create the STS role?
	result, err := client.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	return *result.User.Arn, nil
}

func SetupIAM(sess *session.Session, path string) error {
	client := iam.New(sess)

	userArn, err := getUserArn(client)
	if err != nil {
		return err
	}

	err = setupECSRole(client, path)
	if err != nil {
		return err
	}

	err = setupCrackrRole(client, path, userArn)

	return err
}

func DeleteIAMRoles(sess *session.Session) error {
	client := iam.New(sess)

	err := deleteECSRole(client)
	if err != nil {
		return err
	}

	err = deleteCrackrRole(client)
	return err
}

func createRole(client *iam.IAM, path, roleName, assumeRolePolicyDocument string) error {
	_, err := client.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
		Path:                     aws.String("/" + path + "/"),
		RoleName:                 aws.String(roleName),
		Tags:                     getIAMTags(),
		MaxSessionDuration:       aws.Int64(AssumeRoleDuration),
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

	return nil
}

func attachPolicy(client *iam.IAM, roleName string, managedPolicyArn string) error {
	_, err := client.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(managedPolicyArn),
		RoleName:  aws.String(roleName),
	})

	return err
}

func getRoleArn(sess *session.Session, roleName string) (string, error) {
	client := iam.New(sess)
	result, err := client.GetRole(&iam.GetRoleInput{RoleName: aws.String(roleName)})

	if err != nil {
		return "", err
	}

	return *result.Role.Arn, nil
}
