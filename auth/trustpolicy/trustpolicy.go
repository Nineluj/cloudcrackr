// Package for writing AWS compliant trust policies.
// A trust policy specifies which trusted account members are allowed to assume the role
package trustpolicy

import "encoding/json"

type (
	PolicyDocument struct {
		Version   string
		Statement []StatementEntry
	}

	StatementEntry struct {
		Effect    string
		Action    []string
		Principal Principal
	}

	Principal struct {
		AWSPrincipal string `json:"AWS"`
	}
)

// Builds the policy required by the role whose credentials are given to the cracking
// instance. Format used taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user.html
func BuildUserAssumePolicyDocument(userArn string) (string, error) {
	policy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []StatementEntry{
			{
				Effect:    "Allow",
				Action:    []string{"sts:AssumeRole", "sts:TagSession"},
				Principal: Principal{AWSPrincipal: userArn},
			},
		},
	}

	b, err := json.Marshal(&policy)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
