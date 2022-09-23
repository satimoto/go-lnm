package session

import (
	"encoding/hex"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/notification"
)

type NotificationDto map[string]interface{}

func CreateSessionInvoiceNotificationDto(session db.Session, sessionInvoice db.SessionInvoice) NotificationDto {
	response := map[string]interface{}{
		"type":             notification.SESSION_INVOICE,
		"paymentRequest":   sessionInvoice.PaymentRequest,
		"signature":        hex.EncodeToString(sessionInvoice.Signature),
		"sessionUid":       session.Uid,
		"sessionInvoiceId": sessionInvoice.ID,
		"status":           session.Status,
		"startDatetime":    session.StartDatetime.Format(time.RFC3339),
	}

	if session.EndDatetime.Valid {
		response["endDatetime"] = session.EndDatetime.Time.Format(time.RFC3339)
	}

	return response
}

func CreateSessionUpdateNotificationDto(session db.Session) NotificationDto {
	response := map[string]interface{}{
		"type":       notification.SESSION_UPDATE,
		"sessionUid": session.Uid,
		"status":     session.Status,
	}

	return response
}
