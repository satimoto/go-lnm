package pendingnotification

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/channelrequest"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/pendingnotification"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/notification"
	"github.com/satimoto/go-lnm/internal/service"
)

type PendingNotificationMonitor struct {
	LightningService              lightningnetwork.LightningNetwork
	NotificationService           notification.Notification
	ChannelRequestRepository      channelrequest.ChannelRequestRepository
	PendingNotificationRepository pendingnotification.PendingNotificationRepository
	shutdownCtx                   context.Context
	waitGroup                     *sync.WaitGroup
	nodeID                        int64
}

func NewPendingNotificationMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *PendingNotificationMonitor {
	return &PendingNotificationMonitor{
		LightningService:              services.LightningService,
		NotificationService:           services.NotificationService,
		ChannelRequestRepository:      channelrequest.NewRepository(repositoryService),
		PendingNotificationRepository: pendingnotification.NewRepository(repositoryService),
	}
}

func (s *PendingNotificationMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Pending Notifications")
	s.nodeID = nodeID
	s.shutdownCtx = shutdownCtx
	s.waitGroup = waitGroup

	go s.startPendingNotificationLoop()
}

func (s *PendingNotificationMonitor) startPendingNotificationLoop() {
	for {
		ctx := context.Background()

		if pendingNotifications, err := s.PendingNotificationRepository.ListPendingNotifications(ctx, s.nodeID); err == nil {
			if len(pendingNotifications) > 0 {
				ids := []int64{}
				registrationIDs := []string{}

				for _, pendingNotification := range pendingNotifications {
					if pendingNotification.DeviceToken.Valid {
						ids = append(ids, pendingNotification.ID)
						registrationIDs = append(registrationIDs, pendingNotification.DeviceToken.String)
					}
				}

				message := &fcm.Message{
					RegistrationIDs: registrationIDs,
					CollapseKey:     notification.INVOICE_REQUEST,
					Notification: &fcm.Notification{
						Title: "Ohms! Collect your satoshis!",
						Body:  "You've received some satoshis from others charging their vehicles! Open the app to collect them.",
					},
				}

				_, err := s.NotificationService.SendNotificationWithRetry(message, 10)
				// TODO go-lsp#22: Handle application uninstall

				if err != nil {
					metrics.RecordError("LNM131", "Error sending notification", err)
					log.Printf("LNM131: Message=%#v", message)
				} else {
					updatePendingNotificationsParams := db.UpdatePendingNotificationsParams{
						SendDate: time.Now().Add(time.Hour * 24),
						Ids:      ids,
					}

					err = s.PendingNotificationRepository.UpdatePendingNotifications(ctx, updatePendingNotificationsParams)

					if err != nil {
						metrics.RecordError("LNM132", "Error updating pending notifications", err)
						log.Printf("LNM132: Params=%#v", updatePendingNotificationsParams)
					}

					notification.RecordNotificationSent(notification.INVOICE_REQUEST, len(registrationIDs))
				}
			}
		}

		select {
		case <-s.shutdownCtx.Done():
			log.Printf("Shutting down Pending Notifications")
			return
		case <-time.After(time.Hour):
			continue
		}
	}
}
