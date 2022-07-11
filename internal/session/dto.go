package session

import (
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/notification"
)

func CreatePaymentRequestNotificationDto(session db.Session, sessionInvoice db.SessionInvoice) map[string]interface{} {
	response := map[string]interface{}{
		"type": notification.SESSION_INVOICE,
		"paymentRequest": sessionInvoice.PaymentRequest,
		"sessionUid": session.Uid,
		"sessionInvoiceId": sessionInvoice.ID,
		"status": session.Status,
		"startDatetime": session.StartDatetime.Format(time.RFC3339),
	}

	if session.EndDatetime.Valid {
		response["endDatetime"] = session.EndDatetime.Time.Format(time.RFC3339)
	}

	return response
}
