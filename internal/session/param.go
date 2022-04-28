package session

import (
	"time"

	"github.com/satimoto/go-datastore/db"
)

func NewCreateSessionInvoiceParams(sessionID int64) db.CreateSessionInvoiceParams {
	return db.CreateSessionInvoiceParams{
		SessionID: sessionID,
		Settled: false,
		Expired: false,
		LastUpdated: time.Now(),
}
}

func NewUpdateSessionInvoiceParams(sessionInvoice db.SessionInvoice) db.UpdateSessionInvoiceParams {
	return db.UpdateSessionInvoiceParams{
		ID:          sessionInvoice.ID,
		Settled:     sessionInvoice.Settled,
		Expired:     sessionInvoice.Expired,
		LastUpdated: time.Now(),
	}
}
