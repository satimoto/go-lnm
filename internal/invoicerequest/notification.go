package invoicerequest

import (
	"log"

	"github.com/appleboy/go-fcm"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/notification"
)

func (r *InvoiceRequestResolver) SendInvoiceRequestNotification(user db.User, invoiceRequest db.InvoiceRequest) {
	dto := notification.CreateInvoiceRequestNotificationDto(invoiceRequest)
	
	r.sendNotification(user, dto)
}

func (r *InvoiceRequestResolver) sendNotification(user db.User, data notification.NotificationDto) {
	message := &fcm.Message{
		To:               user.DeviceToken,
		ContentAvailable: true,
		Data:             data,
	}

	_, err := r.NotificationService.SendNotificationWithRetry(message, 10)

	if err != nil {
		// TODO: Cancel session?
		util.LogOnError("LSP141", "Error sending notification", err)
		log.Printf("LSP141: Message=%v", message)
	}
}
