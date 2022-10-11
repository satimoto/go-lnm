package cdr

import (
	"context"
	"log"

	dbUtil "github.com/satimoto/go-datastore/pkg/util"
)

func (r *CdrResolver) Startup(nodeID int64) {
	ctx := context.Background()
	cdrs, err := r.Repository.ListCdrsByCompletedSessionStatus(ctx, nodeID)

	if err != nil {
		dbUtil.LogOnError("LSP134", "Error listing cdrs", err)
		log.Printf("LSP134: NodeID=%v", nodeID)
	}

	for _, cdr := range cdrs {
		log.Printf("Processing cdr %s after restart", cdr.Uid)
		go r.ProcessCdr(cdr)
	}
}
