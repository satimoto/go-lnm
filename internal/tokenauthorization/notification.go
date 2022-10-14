package tokenauthorization

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/notification"
)

func (r *TokenAuthorizationResolver) SendTokenAuthorizeNotification(user db.User, tokenAuthorization db.TokenAuthorization) {
	dto := notification.CreateTokenuthorizeNotificationDto(tokenAuthorization)
	
	r.sendNotification(user, dto)
}

func (r *TokenAuthorizationResolver) sendNotification(user db.User, data notification.NotificationDto) {
	message := &fcm.Message{
		To:               user.DeviceToken,
		ContentAvailable: true,
		Data:             data,
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		util.LogOnError("LSP140", "Error sending notification", err)
		log.Printf("LSP140: Message=%v", message)
	}
}
