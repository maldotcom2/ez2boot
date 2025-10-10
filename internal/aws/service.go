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

// Returned error is consumed only when called from endpoint, not Go routine
func GetEC2Instances(repo *repository.Repository, cfg model.Config, logger *slog.Logger) error {
	awsCFG, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.AWSRegion))
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		return err
	}

	ec2Client := getEC2Client(awsCFG)

	input := getDescribeInstancesInput(cfg.TagKey)

	// Max 1000 responses without pagination
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		logger.Error("Failed to describe EC2 instances", "error", err)
		return err
	}

	servers := []model.Server{}
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {

			// Add to struct
			var i = model.Server{
				UniqueID:    aws.ToString(inst.InstanceId),
				Name:        getTagValue(inst, cfg.TagKey),
				State:       string(inst.State.Name),
				ServerGroup: getTagValue(inst, cfg.TagKey),
				TimeAdded:   time.Now().Unix(),
			}

			servers = append(servers, i)
		}
	}

	// Check number of servers returned from scrape
	if len(servers) > 0 {
		logger.Info("Scraper found matching servers", "count", len(servers))
		repo.UpdateServers(servers)
	} else {
		logger.Info("Scraper found no matching servers")
	}

	return nil
}
