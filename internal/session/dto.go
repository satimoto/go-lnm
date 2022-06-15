package session

import "github.com/satimoto/go-datastore/pkg/db"

func CreatePaymentRequestNotificationDto(sessionInvoice db.SessionInvoice) map[string]interface{} {
	return map[string]interface{}{
		"payReq": sessionInvoice.PaymentRequest,
	}
}
