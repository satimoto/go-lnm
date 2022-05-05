package mocks

type MockNotificationService struct{}

func NewService() *MockNotificationService {
	return &MockNotificationService{}
}

func (s *MockNotificationService) SendNotification() {
}
