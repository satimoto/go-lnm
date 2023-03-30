package cdr

import (
	"context"
	"log"

	metrics "github.com/satimoto/go-lnm/internal/metric"
)

func (r *CdrResolver) Startup(nodeID int64) {
	ctx := context.Background()
	cdrs, err := r.Repository.ListCdrsByCompletedSessionStatus(ctx, nodeID)

	if err != nil {
		metrics.RecordError("LNM134", "Error listing cdrs", err)
		log.Printf("LNM134: NodeID=%v", nodeID)
	}

	for _, cdr := range cdrs {
		log.Printf("Processing cdr %s after restart", cdr.Uid)
		go r.ProcessCdr(cdr)
	}
}
