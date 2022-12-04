package notification

import (
	"encoding/hex"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
)

type NotificationDto map[string]interface{}

func CreateInvoiceRequestNotificationDto(invoiceRequest db.InvoiceRequest) NotificationDto {
	response := map[string]interface{}{
		"type": INVOICE_REQUEST,
	}

	return response
}

func CreateSessionInvoiceNotificationDto(session db.Session, sessionInvoice db.SessionInvoice) NotificationDto {
	response := map[string]interface{}{
		"type":             SESSION_INVOICE,
		"estimatedEnergy":  sessionInvoice.EstimatedEnergy,
		"estimatedTime":    sessionInvoice.EstimatedTime,
		"meteredEnergy":    sessionInvoice.MeteredEnergy,
		"meteredTime":      sessionInvoice.MeteredTime,
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
		"type":       SESSION_UPDATE,
		"sessionUid": session.Uid,
		"status":     session.Status,
	}

	return response
}
