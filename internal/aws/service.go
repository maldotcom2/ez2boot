package aws

import (
	"context"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func GetEC2Instances(repo *repository.Repository, tagKey string, logger *slog.Logger) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-2"))
	if err != nil {
		// TODO Logger
		return err
	}

	ec2Client := getEC2Client(cfg)

	input := getDescribeInstancesInput(tagKey)

	// Max 1000 responses without pagination
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		// TODO Logger
		return err
	}

	servers := []model.Server{}
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {

			// Add to struct
			var i = model.Server{
				UniqueID:    aws.ToString(inst.InstanceId),
				Name:        getTagValue(inst, tagKey),
				State:       string(inst.State.Name),
				ServerGroup: getTagValue(inst, tagKey),
				TimeAdded:   time.Now().Unix(),
			}

			servers = append(servers, i)
		}
	}

	repo.AddOrUpdateServers(servers, logger)

	return nil
}
