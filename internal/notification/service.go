package notification

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
)

type Notification interface {
	SendNotification(*fcm.Message) (*fcm.Response, error)
	SendNotificationWithRetry(message *fcm.Message, retries int) (*fcm.Response, error)
	SendUserNotification(user db.User, data NotificationDto, notificationType string)
}

type NotificationService struct {
	client *fcm.Client
}

func NewService(apiKey string) Notification {
	client, err := fcm.NewClient(apiKey)
	util.PanicOnError("LSP034", "Invalid FCM API key", err)

	return &NotificationService{
		client: client,
	}
}

func (s *NotificationService) SendNotification(message *fcm.Message) (*fcm.Response, error) {
	log.Printf("Sending notification: %v", message.To)
	log.Printf("Data=%#v", message.Data)
	return s.client.Send(message)
}

func (s *NotificationService) SendNotificationWithRetry(message *fcm.Message, retries int) (*fcm.Response, error) {
	log.Printf("Sending notification with retry: %v", message.To)
	log.Printf("Data=%#v", message.Data)
	return s.client.SendWithRetry(message, retries)
}

func (s *NotificationService) SendUserNotification(user db.User, data NotificationDto, notificationType string) {
	message := &fcm.Message{
		To:               user.DeviceToken.String,
		ContentAvailable: true,
		Data:             data,
	}

	_, err := s.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		metrics.RecordError("LSP059", "Error sending notification", err)
		log.Printf("LSP059: Message=%v", message)
	}

	RecordNotificationSent(notificationType, 1)
}
