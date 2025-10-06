package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func GetEC2Instances() ([]Instance, error) {
	var filterTag string = "ez2boot" // TODO: Allow way to customise this tag key
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-2"))
	if err != nil {
		// TODO Logger
		return nil, err
	}

	ec2Client := getEC2Client(cfg)

	input := getDescribeInstancesInput(filterTag)

	// Max 1000 responses without pagination
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		// TODO Logger
		return nil, err
	}

	instances := []Instance{}
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {

			// Add to struct
			var i = Instance{
				InstanceId:  aws.ToString(inst.InstanceId),
				Name:        getTagValue(inst, filterTag),
				ServerGroup: getTagValue(inst, filterTag),
				TimeAdded:   time.Now().Unix(),
			}

			instances = append(instances, i)
		}
	}

	return instances, nil
}
