package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Filter for opt-in instances
func getDescribeInstancesInput(tagKey string) *ec2.DescribeInstancesInput {
	return &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []string{tagKey},
			},
		},
	}
}
