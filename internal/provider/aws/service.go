package aws

import (
	"context"
	"ez2boot/internal/server"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Scrape AWS to retrieve servers.
func (s *Service) Scrape() error {
	s.Logger.Debug("Scraping AWS")

	input := getDescribeInstancesInput(s.Config.TagKey) // Target tagged instances

	// Describe instances from AWS. Max 1000 responses without pagination
	result, err := s.EC2Client.DescribeInstances(context.Background(), input)
	if err != nil {
		s.Logger.Error("Failed to describe EC2 instances", "error", err)
		return err
	}

	servers := []server.Server{}
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {

			// Add to struct
			var svr = server.Server{
				UniqueID:    aws.ToString(inst.InstanceId),
				Name:        getTagValue(inst, "Name"),
				State:       mapState(string(inst.State.Name)),
				ServerGroup: getTagValue(inst, s.Config.TagKey),
				TimeAdded:   time.Now().Unix(),
			}

			servers = append(servers, svr)
		}
	}

	s.Logger.Debug("Scraped and found number of matching servers", "count", len(servers))
	s.ServerService.UpdateServers(servers)

	return nil
}

// Start required AWS servers
func (s *Service) Start() error {
	s.Logger.Debug("Starting requested AWS servers")

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

		result, err := s.EC2Client.StartInstances(context.Background(), input)
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

// Stop no longer required AWS servers
func (s *Service) Stop() error {
	s.Logger.Debug("Stopping requested AWS servers")

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

		result, err := s.EC2Client.StopInstances(context.Background(), input)
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
