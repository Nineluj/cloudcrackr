package network

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetDefaultSubnetArn(sess *session.Session) (string, error) {
	client := ec2.New(sess)

	var subnets []*ec2.Subnet

	describeSubnetInput := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("default-for-az"),
				Values: []*string{aws.String("true")},
			},
		},
	}

	for {
		result, err := client.DescribeSubnets(describeSubnetInput)

		if err != nil {
			return "", err
		}

		subnets = append(subnets, result.Subnets...)

		if result.NextToken == nil || len(result.Subnets) != 0 {
			break
		}

		describeSubnetInput.NextToken = result.NextToken
	}

	// it's probably fine to have more than 1 subnet found
	if len(subnets) == 0 {
		return "", errors.New("found no default subnet to use")
	}

	return *subnets[0].SubnetArn, nil
}
