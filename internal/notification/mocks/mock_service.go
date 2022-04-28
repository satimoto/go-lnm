package mocks

import (
	"github.com/satimoto/go-lsp/internal/notification"
)

type MockNotificationService struct{}

func NewService() notification.Notification {
	return &MockNotificationService{}
}

func (s *MockNotificationService) SendNotification() {
}
