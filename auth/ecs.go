package auth

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	ECSRoleName         = "ecsTaskExecutionRole"
	ECSManagedPolicyArn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
)

func setupECSRole(client *iam.IAM, path string) error {
	err := createRole(client, path, ECSRoleName, DefaultAssumeRolePolicyDocument)
	if err != nil {
		return err
	}

	err = attachPolicy(client, ECSRoleName, ECSManagedPolicyArn)
	return err
}

func GetECSRoleArn(sess *session.Session) (string, error) {
	return getRoleArn(sess, ECSRoleName)
}

func deleteECSRole(client *iam.IAM) error {
	_, err := client.DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String(ECSRoleName)})
	return err
}
