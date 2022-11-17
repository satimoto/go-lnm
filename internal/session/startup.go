package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
)

func (r *SessionResolver) Startup(nodeID int64) {
	ctx := context.Background()
	utcTime := time.Now().UTC()

	sessions, err := r.Repository.ListInProgressSessionsByNodeID(ctx, dbUtil.SqlNullInt64(nodeID))

	if err != nil {
		metrics.RecordError("LSP135", "Error listing sessions", err)
		log.Printf("LSP135: NodeID=%v", nodeID)
	}

	for _, session := range sessions {
		if session.Status == db.SessionStatusTypePENDING {
			pendingDuration := utcTime.Sub(session.StartDatetime).Seconds()

			if pendingDuration > 90 {
				if user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID); err == nil {
					r.invalidateSession(ctx, user, session)
					continue	
				}
			}
		}

		log.Printf("Monitoring session %s after restart", session.Uid)
		go r.StartSessionMonitor(session)
	}
}
