// Package for writing AWS compliant permission policies.
// A permissions policy grants roles the needed permissions to carry out the intended tasks on the resource
package permissionpolicy

import "encoding/json"

type (
	PolicyDocument struct {
		Version   string
		Statement []StatementEntry
	}

	StatementEntry struct {
		Effect   string
		Action   []string
		Resource string
	}
)

func (pd *PolicyDocument) ToString() (string, error) {
	b, err := json.Marshal(pd)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Builds the policy needed for a crackr image. Based on:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/iam-example-policies.html
func BuildCrackrPolicy(dictionarySourceArn, hashSourceArn, outputBucketArn string) (string, error) {
	policy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetObject",
				},
				Resource: dictionarySourceArn,
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetObject",
				},
				Resource: hashSourceArn,
			},

			{
				Effect: "Allow",
				Action: []string{
					"s3:PutObject",
				},
				Resource: outputBucketArn + "/*",
			},
		},
	}

	return policy.ToString()
}
