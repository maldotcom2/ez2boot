package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getTagValue(inst ec2types.Instance, tagKey string) string {
	for _, tag := range inst.Tags {
		if aws.ToString(tag.Key) == tagKey {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}
