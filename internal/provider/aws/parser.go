package aws

import (
	"ez2boot/internal/server"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Iterate the tags on the instance to find the value
func getTagValue(inst ec2types.Instance, tagKey string) string {
	for _, tag := range inst.Tags {
		if aws.ToString(tag.Key) == tagKey {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// Map provider specific states to generic
func mapState(state string) server.ServerState {
	switch state {
	case "running":
		return server.ServerOn
	case "stopping", "pending":
		return server.ServerTransitioning
	default:
		return server.ServerOff
	}
}
