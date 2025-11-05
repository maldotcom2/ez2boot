package aws

import (
	"context"
	"ez2boot/internal/server"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (s *Service) Start() error {
	s.Logger.Debug("Starting requested AWS servers")
	awsCFG, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s.Config.AWSRegion))
	if err != nil {
		s.Logger.Error("Failed to load AWS config", "error", err)
		return err
	}

	ec2Client := getEC2Client(awsCFG)

	// Get start instance IDs
	instanceIDs, err := s.ServerService.GetPending("off", "on")
	if err != nil {
		s.Logger.Error("Failed to get instance IDs pending on", "error", err)
		return err
	}

	// Nothing to do
	if len(instanceIDs) == 0 {
		s.Logger.Debug("No instances to start")
		return nil
	}

	// Loop and turn each on
	for _, id := range instanceIDs {
		s.Logger.Debug("Starting", "id", id)
		input := &ec2.StartInstancesInput{
			InstanceIds: []string{id},
		}

		result, err := ec2Client.StartInstances(context.TODO(), input)
		if err != nil {
			s.Logger.Error("failed to start instance", "id", id, "error", err)
			continue
		}

		for _, instance := range result.StartingInstances {
			s.Logger.Info("Instance start initiated",
				"id", aws.ToString(instance.InstanceId),
				"previous_state", instance.PreviousState.Name,
				"current_state", instance.CurrentState.Name,
			)
		}
	}

	return nil
}

func (s *Service) Stop() error {
	s.Logger.Debug("Stopping requested AWS servers")
	awsCFG, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s.Config.AWSRegion))
	if err != nil {
		s.Logger.Error("Failed to load AWS config", "error", err)
		return err
	}

	ec2Client := getEC2Client(awsCFG)

	// Get start instance IDs
	instanceIDs, err := s.ServerService.GetPending("on", "off")
	if err != nil {
		s.Logger.Error("Failed to get instance IDs pending off", "error", err)
		return err
	}

	// Nothing to do
	if len(instanceIDs) == 0 {
		s.Logger.Debug("No instances to stop")
		return nil
	}

	// Loop and turn each on
	for _, id := range instanceIDs {
		s.Logger.Debug("Stopping", "id", id)
		input := &ec2.StopInstancesInput{
			InstanceIds: []string{id},
		}

		result, err := ec2Client.StopInstances(context.TODO(), input)
		if err != nil {
			s.Logger.Error("Failed to stop instance", "id", id, "error", err)
			continue
		}

		for _, instance := range result.StoppingInstances {
			s.Logger.Info("Instance stop initiated",
				"id", aws.ToString(instance.InstanceId),
				"previous_state", instance.PreviousState.Name,
				"current_state", instance.CurrentState.Name,
			)
		}
	}

	return nil
}
