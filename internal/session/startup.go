package session

import (
	"context"
	"log"

	dbUtil "github.com/satimoto/go-datastore/pkg/util"
)

func (r *SessionResolver) Startup(nodeID int64) {
	ctx := context.Background()
	sessions, err := r.Repository.ListInProgressSessionsByNodeID(ctx, dbUtil.SqlNullInt64(nodeID))

	if err != nil {
		dbUtil.LogOnError("LSP135", "Error listing sessions", err)
		log.Printf("LSP135: NodeID=%v", nodeID)
	}

	for _, session := range sessions {
		log.Printf("Monitoring session %s after restart", session.Uid)
		go r.StartSessionMonitor(session)
	}
}
