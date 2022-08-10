package notification

import (
	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/util"
)

type Notification interface {
	SendNotification(*fcm.Message) (*fcm.Response, error)
	SendNotificationWithRetry(message *fcm.Message, retries int) (*fcm.Response, error)
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
	return s.client.Send(message)
}

func (s *NotificationService) SendNotificationWithRetry(message *fcm.Message, retries int) (*fcm.Response, error) {
	return s.client.SendWithRetry(message, retries)
}
