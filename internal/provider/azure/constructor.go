package azure

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

func NewService(azureRepo *Repository, cfg *config.Config, serverService *server.Service, logger *slog.Logger) (*Service, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to load Azure credentials %w", err)
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(cfg.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %w", err)
	}

	return &Service{
		Repo:          azureRepo,
		Config:        cfg,
		ServerService: serverService,
		VMClient:      vmClient,
		Logger:        logger,
	}, nil
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
