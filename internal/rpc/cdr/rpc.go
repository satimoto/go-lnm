package cdr

import (
	"context"
	"errors"
	"log"

	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcCdrResolver) CdrCreated(ctx context.Context, input *ocpirpc.CdrCreatedRequest) (*ocpirpc.CdrCreatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		cdr, err := r.CdrResolver.Repository.GetCdrByUid(ctx, input.CdrUid)

		if err != nil {
			metrics.RecordError("LSP055", "Error retrieving cdr", err)
			log.Printf("LSP055: CdrUid=%v", input.CdrUid)
			return nil, errors.New("cdr not found")
		}

		go r.CdrResolver.ProcessCdr(cdr)

		return &ocpirpc.CdrCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
