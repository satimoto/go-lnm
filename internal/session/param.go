package session

import (
	"time"

	"github.com/satimoto/go-datastore/db"
)

func NewCreateSessionInvoiceParams(sessionID int64) db.CreateSessionInvoiceParams {
	return db.CreateSessionInvoiceParams{
		SessionID:   sessionID,
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
