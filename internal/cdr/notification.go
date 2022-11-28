package cdr

import (
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/notification"
)

func (r *CdrResolver) SendInvoiceRequestNotification(user db.User, invoiceRequest db.InvoiceRequest) {
	dto := notification.CreateInvoiceRequestNotificationDto(invoiceRequest)

	r.NotificationService.SendUserNotification(user, dto, notification.INVOICE_REQUEST)
}
