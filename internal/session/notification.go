package session

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/notification"
)

func (r *SessionResolver) SendSessionInvoiceNotification(user db.User, session db.Session, sessionInvoice db.SessionInvoice) {
	dto := notification.CreateSessionInvoiceNotificationDto(session, sessionInvoice)

	r.NotificationService.SendUserNotification(user, dto, notification.SESSION_INVOICE)
}

func (r *SessionResolver) SendSessionUpdateNotification(user db.User, session db.Session) {
	dto := notification.CreateSessionUpdateNotificationDto(session)

	r.NotificationService.SendUserNotification(user, dto, notification.SESSION_UPDATE)
}
