package session

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
)

func (r *SessionResolver) SendSessionInvoiceNotification(user db.User, session db.Session, sessionInvoice db.SessionInvoice) {
	dto := CreateSessionInvoiceNotificationDto(session, sessionInvoice)
	
	r.sendNotification(user, dto)
}

func (r *SessionResolver) SendSessionUpdateNotification(user db.User, session db.Session) {
	dto := CreateSessionUpdateNotificationDto(session)
	
	r.sendNotification(user, dto)
}

func (r *SessionResolver) sendNotification(user db.User, data NotificationDto) {
	message := &fcm.Message{
		To:               user.DeviceToken,
		ContentAvailable: true,
		Data:             data,
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		util.LogOnError("LSP059", "Error sending notification", err)
		log.Printf("LSP059: Message=%v", message)
	}
}
