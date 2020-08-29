package auth

import (
	"cloudcrackr/cmd/utility"
	"cloudcrackr/constants"
	"github.com/aws/aws-sdk-go/aws"
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

func ignoreNoSuchEntityError(err error) error {
	return utility.IgnoreAWSError(err, iam.ErrCodeNoSuchEntityException)
}

func DeleteIAMRoles(sess *session.Session) error {
	client := iam.New(sess)

	err := deleteECSRole(client)
	if ignoreNoSuchEntityError(err) != nil {
		return err
	}

	err = deleteCrackrRole(client)
	return ignoreNoSuchEntityError(err)
}

func createRole(client *iam.IAM, path, roleName, assumeRolePolicyDocument string) error {
	_, err := client.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
		Path:                     aws.String("/" + path + "/"),
		RoleName:                 aws.String(roleName),
		Tags:                     getIAMTags(),
		MaxSessionDuration:       aws.Int64(AssumeRoleDuration),
	})

	return utility.IgnoreAWSError(err, iam.ErrCodeEntityAlreadyExistsException)
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

// Deletes and detaches the two kinds of related policies
// RolePolicies (=inline policies) and AttachedPolicies
func clearIAMRolePolicies(client *iam.IAM, roleName string) error {
	rn := aws.String(roleName)

	var deleteErr error

	err := client.ListRolePoliciesPages(&iam.ListRolePoliciesInput{
		RoleName: rn,
	}, func(out *iam.ListRolePoliciesOutput, _ bool) bool {
		for _, policyName := range out.PolicyNames {
			_, deleteErr = client.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
				PolicyName: policyName,
				RoleName:   rn,
			})

			if deleteErr != nil {
				return false
			}
		}
		return true
	})

	if deleteErr != nil {
		return deleteErr
	}

	if err != nil {
		return err
	}

	var detachErr error
	err = client.ListAttachedRolePoliciesPages(&iam.ListAttachedRolePoliciesInput{
		RoleName: rn,
	}, func(out *iam.ListAttachedRolePoliciesOutput, _ bool) bool {
		for _, policy := range out.AttachedPolicies {
			_, detachErr = client.DetachRolePolicy(&iam.DetachRolePolicyInput{
				PolicyArn: policy.PolicyArn,
				RoleName:  rn,
			})

			if detachErr != nil {
				return false
			}
		}
		return true
	})

	if detachErr != nil {
		return detachErr
	}

	return err
}
