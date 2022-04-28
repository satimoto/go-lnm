package notification

type Notification interface {
	SendNotification()
}

type NotificationService struct{}

// TODO: Implement notification service
func NewService() Notification {
	return &NotificationService{}
}

func (s *NotificationService) SendNotification() {
}
