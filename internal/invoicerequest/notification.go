package invoicerequest

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/notification"
)

func (r *InvoiceRequestResolver) SendInvoiceRequestNotification(user db.User, invoiceRequest db.InvoiceRequest) {
	dto := notification.CreateInvoiceRequestNotificationDto(invoiceRequest)

	r.sendNotification(user, dto)
}

func (r *InvoiceRequestResolver) sendNotification(user db.User, data notification.NotificationDto) {
	message := &fcm.Message{
		To:               user.DeviceToken.String,
		ContentAvailable: true,
		Priority:         "high",
		Data:             data,
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		metrics.RecordError("LNM141", "Error sending notification", err)
		log.Printf("LNM141: Message=%v", message)
	}

	notification.RecordNotificationSent(notification.INVOICE_REQUEST, 1)
}
