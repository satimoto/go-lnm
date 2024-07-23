package session

import (
	"context"
	"log"

	"github.com/satimoto/go-datastore/pkg/db"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lnm/internal/metric"
)

func (r *SessionResolver) Startup(nodeID int64) {
	ctx := context.Background()

	// Start monitoring in progress sessions
	sessions, err := r.Repository.ListInProgressSessionsByNodeID(ctx, dbUtil.SqlNullInt64(nodeID))

	if err != nil {
		metrics.RecordError("LNM135", "Error listing sessions", err)
		log.Printf("LNM135: NodeID=%v", nodeID)
	}

	for _, session := range sessions {
		log.Printf("Monitoring session %s after restart", session.Uid)
		go r.StartSessionMonitor(session)
	}

	// List session invoices to check expiry
	listSessionInvoicesParams := db.ListSessionInvoicesByNodeIDParams{
		NodeID:    dbUtil.SqlNullInt64(nodeID),
		IsExpired: false,
		IsSettled: false,
	}

	sessionInvoices, err := r.Repository.ListSessionInvoicesByNodeID(ctx, listSessionInvoicesParams)

	if err != nil {
		metrics.RecordError("LNM160", "Error listing sessions", err)
		log.Printf("LNM160: Params=%#v", listSessionInvoicesParams)
	}

	for _, sessionInvoice := range sessionInvoices {
		go r.WaitForInvoiceExpiry(sessionInvoice.PaymentRequest)
	}
}
