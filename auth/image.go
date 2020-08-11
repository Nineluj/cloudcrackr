package auth

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	ImageRoleName = "clientRole"
	//ECSManagedPolicyArn = "arn:aws:auth::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
)

// Retrieves the temporary image credentials
func GetImageCredentials(sess *session.Session) ([]*string, error) {
	return nil, errors.New("not implemented")
	client := sts.New(sess)

	result, err := client.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds:   nil,
		ExternalId:        nil,
		Policy:            nil,
		PolicyArns:        nil,
		RoleArn:           nil,
		RoleSessionName:   nil,
		SerialNumber:      nil,
		Tags:              nil,
		TokenCode:         nil,
		TransitiveTagKeys: nil,
	})

	_ = err

	return []*string{
		result.Credentials.AccessKeyId,
		result.Credentials.SecretAccessKey,
		result.Credentials.SessionToken,
	}, nil
}
