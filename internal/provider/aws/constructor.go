package aws

import (
	"context"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"fmt"
	"log/slog"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func NewService(awsRepo *Repository, cfg *config.Config, serverService *server.Service, logger *slog.Logger) (*Service, error) {
	awsCFG, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, fmt.Errorf("Failed to load AWS config %w", err)
	}

	ec2Client := ec2.NewFromConfig(awsCFG)

	return &Service{
		Repo:          awsRepo,
		Config:        cfg,
		ServerService: serverService,
		EC2Client:     ec2Client,
		Logger:        logger,
	}, nil
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
