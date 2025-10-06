package aws

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getTagValue(inst ec2types.Instance, tagKey string) string {
	for _, tag := range inst.Tags {
		if tag.Key != nil && *tag.Key == tagKey {
			if tag.Value != nil {
				return *tag.Value
			}
		}
	}
	return ""
}
