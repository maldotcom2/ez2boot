package aws

import (
	"context"
	"ez2boot/internal/server"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Returned error is consumed only when called from endpoint, not Go routine
func (s *Service) Scrape() error {
	s.Logger.Debug("Scraping AWS")
	awsCFG, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s.Config.AWSRegion))
	if err != nil {
		s.Logger.Error("Failed to load AWS config", "error", err)
		return err
	}

	ec2Client := getEC2Client(awsCFG)

	input := getDescribeInstancesInput(s.Config.TagKey)

	// Describe instances from AWS. Max 1000 responses without pagination
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		s.Logger.Error("Failed to describe EC2 instances", "error", err)
		return err
	}

	servers := []server.Server{}
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {

			// Add to struct
			var s = server.Server{
				UniqueID:    aws.ToString(inst.InstanceId),
				Name:        getTagValue(inst, "Name"),
				State:       string(inst.State.Name),
				ServerGroup: getTagValue(inst, s.Config.TagKey),
				TimeAdded:   time.Now().Unix(),
			}

			servers = append(servers, s)
		}
	}

	// Check number of servers returned from scrape
	if len(servers) > 0 {
		s.Logger.Debug("Scraper found matching servers", "count", len(servers))
		s.ServerService.UpdateServers(servers)
	} else {
		s.Logger.Debug("Scraper found no matching servers")
	}

	return nil
}
