package aws

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo          *Repository
	Config        *config.Config
	ServerService *server.Service
	EC2Client     *ec2.Client
	Logger        *slog.Logger
}
