package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func getEC2Client(cfg aws.Config) *ec2.Client {
	ec2Client := ec2.NewFromConfig(cfg)

	return ec2Client
}
