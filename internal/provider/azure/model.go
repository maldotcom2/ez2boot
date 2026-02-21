package azure

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo          *Repository
	Config        *config.Config
	ServerService *server.Service
	VMClient      *armcompute.VirtualMachinesClient
	Logger        *slog.Logger
}
