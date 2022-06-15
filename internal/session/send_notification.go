package session

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
)

func (r *SessionResolver) SendNotification(user db.User, session db.Session, sessionInvoice db.SessionInvoice) {
	message := &fcm.Message{
		To:               user.DeviceToken,
		ContentAvailable: true,
		Data:             CreatePaymentRequestNotificationDto(sessionInvoice),
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		util.LogOnError("LSP059", "Error sending notification", err)
		log.Printf("LSP059: Message=%v", message)
	}
}
