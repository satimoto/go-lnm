package cdr

import (
	"context"
	"errors"
	"log"

	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcCdrResolver) CdrCreated(reqCtx context.Context, input *ocpirpc.CdrCreatedRequest) (*ocpirpc.CdrCreatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		ctx := context.Background()
		cdr, err := r.CdrResolver.Repository.GetCdrByUid(ctx, input.CdrUid)

		if err != nil {
			metrics.RecordError("LNM055", "Error retrieving cdr", err)
			log.Printf("LNM055: CdrUid=%v", input.CdrUid)
			return nil, errors.New("cdr not found")
		}

		go r.CdrResolver.ProcessCdr(cdr)

		return &ocpirpc.CdrCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
