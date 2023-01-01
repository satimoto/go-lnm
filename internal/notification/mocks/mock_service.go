package mocks

import (
	"errors"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/notification"
)

type MockNotificationService struct {
	sendNotificationMessageMockData  []*fcm.Message
	sendNotificationResponseMockData []*fcm.Response
}

func NewService() *MockNotificationService {
	return &MockNotificationService{}
}

func (s *MockNotificationService) SendNotification(message *fcm.Message) (*fcm.Response, error) {
	s.sendNotificationMessageMockData = append(s.sendNotificationMessageMockData, message)

	if len(s.sendNotificationResponseMockData) == 0 {
		return &fcm.Response{}, errors.New("NotFound")
	}

	response := s.sendNotificationResponseMockData[0]
	s.sendNotificationResponseMockData = s.sendNotificationResponseMockData[1:]
	return response, nil
}

func (s *MockNotificationService) SendNotificationWithRetry(message *fcm.Message, retries int) (*fcm.Response, error) {
	s.sendNotificationMessageMockData = append(s.sendNotificationMessageMockData, message)

	if len(s.sendNotificationResponseMockData) == 0 {
		return &fcm.Response{}, errors.New("NotFound")
	}

	response := s.sendNotificationResponseMockData[0]
	s.sendNotificationResponseMockData = s.sendNotificationResponseMockData[1:]
	return response, nil
}

func (s *MockNotificationService) SetSendNotificationMockData(message *fcm.Response) {
	s.sendNotificationResponseMockData = append(s.sendNotificationResponseMockData, message)
}

func (s *MockNotificationService) SendUserNotification(user db.User, data notification.NotificationDto, notificationType string) {
}
