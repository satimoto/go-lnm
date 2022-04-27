package notification

type Notification interface{}

type NotificationService struct{}

// TODO: Implement notification service
func NewService() Notification {
	return &NotificationService{}
}
