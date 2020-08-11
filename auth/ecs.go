package auth

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	ECSRoleName         = "ecsTaskExecutionRole"
	ECSManagedPolicyArn = "arn:aws:auth::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
)

func createECSRole(client *iam.IAM, path string) error {
	return createRole(client, path, ECSRoleName, ECSManagedPolicyArn)
}

func GetECSRoleArn(sess *session.Session) (string, error) {
	return getRoleArn(sess, ECSRoleName)
}
