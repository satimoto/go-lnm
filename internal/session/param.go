package session

import (
	"time"

	"github.com/satimoto/go-datastore/db"
)

func NewCreateSessionInvoiceParams(session db.Session) db.CreateSessionInvoiceParams {
	return db.CreateSessionInvoiceParams{
		SessionID:   session.ID,
		Currency:    session.Currency,
		IsSettled:   false,
		IsExpired:   false,
		LastUpdated: time.Now(),
	}
}

func NewUpdateSessionInvoiceParams(sessionInvoice db.SessionInvoice) db.UpdateSessionInvoiceParams {
	return db.UpdateSessionInvoiceParams{
		ID:          sessionInvoice.ID,
		IsSettled:   sessionInvoice.IsSettled,
		IsExpired:   sessionInvoice.IsExpired,
		LastUpdated: time.Now(),
	}
}
