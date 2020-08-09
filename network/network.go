package network

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetSecurityGroup(sess *session.Session) {
	client := ec2.New(sess)

	client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description:       nil,
		DryRun:            nil,
		GroupName:         nil,
		TagSpecifications: nil,
		VpcId:             nil,
	})
}
