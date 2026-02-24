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
	s.Logger.Debug("Scraping AWS", "domain", "aws")

	input := getDescribeInstancesInput(s.Config.TagKey) // Target tagged instances

	// Describe instances from AWS. Max 1000 responses without pagination
	result, err := s.EC2Client.DescribeInstances(context.Background(), input)
	if err != nil {
		s.Logger.Error("Failed to describe EC2 instances", "domain", "aws", "error", err)
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

	s.Logger.Debug("Scraped and found number of matching instances", "domain", "aws", "count", len(servers))
	s.ServerService.UpdateServers(servers)

	return nil
}

// Start required AWS servers
func (s *Service) Start() error {
	s.Logger.Debug("Starting requested AWS instances", "domain", "aws")

	// Get start instance IDs
	instanceIDs, err := s.ServerService.GetPending("off", "on")
	if err != nil {
		s.Logger.Error("Failed to get instance IDs pending on", "domain", "aws", "error", err)
		return err
	}

	// Nothing to do
	if len(instanceIDs) == 0 {
		s.Logger.Debug("No instances to start", "domain", "aws")
		return nil
	}

	// Loop and turn each on
	for _, id := range instanceIDs {
		s.Logger.Debug("Starting", "id", id, "domain", "aws")
		input := &ec2.StartInstancesInput{
			InstanceIds: []string{id},
		}

		result, err := s.EC2Client.StartInstances(context.Background(), input)
		if err != nil {
			s.Logger.Error("Failed to start instance", "id", id, "domain", "aws", "error", err)
			continue
		}

		for _, instance := range result.StartingInstances {
			s.Logger.Info("Instance start initiated", "domain", "aws",
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
	s.Logger.Debug("Stopping requested AWS instances", "domain", "aws")

	// Get start instance IDs
	instanceIDs, err := s.ServerService.GetPending("on", "off")
	if err != nil {
		s.Logger.Error("Failed to get instance IDs pending off", "domain", "aws", "error", err)
		return err
	}

	// Nothing to do
	if len(instanceIDs) == 0 {
		s.Logger.Debug("No instances to stop", "domain", "aws")
		return nil
	}

	// Loop and turn each on
	for _, id := range instanceIDs {
		s.Logger.Debug("Stopping", "id", id, "domain", "aws")
		input := &ec2.StopInstancesInput{
			InstanceIds: []string{id},
		}

		result, err := s.EC2Client.StopInstances(context.Background(), input)
		if err != nil {
			s.Logger.Error("Failed to stop instance", "id", id, "domain", "aws", "error", err)
			continue
		}

		for _, instance := range result.StoppingInstances {
			s.Logger.Info("Instance stop initiated", "domain", "aws",
				"id", aws.ToString(instance.InstanceId),
				"previous_state", instance.PreviousState.Name,
				"current_state", instance.CurrentState.Name,
			)
		}
	}

	return nil
}
