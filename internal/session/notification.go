package session

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/notification"
)

func (r *SessionResolver) SendSessionInvoiceNotification(user db.User, session db.Session, sessionInvoice db.SessionInvoice) {
	dto := notification.CreateSessionInvoiceNotificationDto(session, sessionInvoice)

	r.sendNotification(user, dto, notification.SESSION_INVOICE)
}

func (r *SessionResolver) SendSessionUpdateNotification(user db.User, session db.Session) {
	dto := notification.CreateSessionUpdateNotificationDto(session)

	r.sendNotification(user, dto, notification.SESSION_UPDATE)
}

func (r *SessionResolver) sendNotification(user db.User, data notification.NotificationDto, notificationType string) {
	message := &fcm.Message{
		To:               user.DeviceToken.String,
		ContentAvailable: true,
		Data:             data,
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		metrics.RecordError("LSP059", "Error sending notification", err)
		log.Printf("LSP059: Message=%v", message)
	}

	notification.RecordNotificationSent(notificationType, 1)
}
