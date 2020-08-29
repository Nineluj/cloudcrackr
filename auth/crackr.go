package auth

import (
	"cloudcrackr/auth/permissionpolicy"
	"cloudcrackr/auth/trustpolicy"
	"cloudcrackr/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"strings"
)

const (
	CrackrRoleName = "crackrRole"
	// TODO: handle rotation of credentials?
	AssumeRoleDuration = 43200 // 12 hours
)

func getSTSTags() []*sts.Tag {
	return []*sts.Tag{
		{
			Key:   aws.String(constants.TagKey),
			Value: aws.String(constants.TagValue),
		},
	}
}

func setupCrackrRole(client *iam.IAM, path, userArn string) error {
	// We need to pass a more complicated assume policy document since the user will
	// need to assume this role in order to retrieve the credentials that are needed
	// for the crackr to retrieve and upload files to the S3 bucket
	userAllowedAssumePolicyDocument, err := trustpolicy.BuildUserAssumePolicyDocument(userArn)
	if err != nil {
		return err
	}

	return createRole(client, path, CrackrRoleName, userAllowedAssumePolicyDocument)
}

func deleteCrackrRole(client *iam.IAM) error {
	_, err := client.DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String(CrackrRoleName)})
	return err
}

func getCrackrRoleArn(sess *session.Session) (string, error) {
	return getRoleArn(sess, CrackrRoleName)
}

func s3ArnFormat(name string) string {
	bucketWithKey := strings.TrimLeft(name, "s3://")
	return "arn:aws:s3:::" + bucketWithKey
}

// Retrieves the temporary image credentials
func GetCrackrCredentials(sess *session.Session, dictionaryS3, hashS3, outputS3 string) ([]*string, error) {
	client := sts.New(sess)

	roleArn, err := getCrackrRoleArn(sess)

	builtPolicy, err := permissionpolicy.BuildCrackrPolicy(
		s3ArnFormat(dictionaryS3), s3ArnFormat(hashS3), s3ArnFormat(outputS3),
	)
	if err != nil {
		return nil, err
	}

	result, err := client.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds:   aws.Int64(AssumeRoleDuration),
		Policy:            aws.String(builtPolicy),
		RoleArn:           aws.String(roleArn),
		RoleSessionName:   aws.String(CrackrRoleName),
		Tags:              getSTSTags(),
		TransitiveTagKeys: []*string{aws.String(constants.TagKey)},
	})

	if err != nil {
		return nil, err
	}

	return []*string{
		result.Credentials.AccessKeyId,
		result.Credentials.SecretAccessKey,
		result.Credentials.SessionToken,
	}, nil
}
