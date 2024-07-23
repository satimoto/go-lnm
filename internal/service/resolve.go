package service

import (
	"os"

	"github.com/satimoto/go-lnm/internal/ferp"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	"github.com/satimoto/go-lnm/internal/notification"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

type ServiceResolver struct {
	FerpService         ferp.Ferp
	LightningService    lightningnetwork.LightningNetwork
	NotificationService notification.Notification
	OcpiService         ocpi.Ocpi
}

func NewService() *ServiceResolver {
	ferpService := ferp.NewService(os.Getenv("FERP_RPC_ADDRESS"))
	lightningService := lightningnetwork.NewService()
	notificationService := notification.NewService(os.Getenv("FCM_API_KEY"))
	ocpiService := ocpi.NewService(os.Getenv("OCPI_RPC_ADDRESS"))

	return &ServiceResolver{
		FerpService:         ferpService,
		LightningService:    lightningService,
		OcpiService:         ocpiService,
		NotificationService: notificationService,
	}
}
