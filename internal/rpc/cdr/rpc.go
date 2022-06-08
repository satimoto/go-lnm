package cdr

import (
	"context"
	"errors"

	"github.com/satimoto/go-ocpi-api/ocpirpc"
)

func (r *RpcCdrResolver) CdrCreated(ctx context.Context, input *ocpirpc.CdrCreatedRequest) (*ocpirpc.CdrCreatedResponse, error) {
	if input != nil {
		cdr, err := r.CdrResolver.Repository.GetCdrByUid(ctx, input.CdrUid)

		if err != nil {
			return nil, errors.New("cdr not found")
		}

		go r.CdrResolver.ProcessCdr(ctx, cdr)

		return &ocpirpc.CdrCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
