package session

import (
	"context"
	"errors"
	"log"

	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcSessionResolver) SessionCreated(reqCtx context.Context, input *ocpirpc.SessionCreatedRequest) (*ocpirpc.SessionCreatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		ctx := context.Background()
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			metrics.RecordError("LNM058", "Error retrieving session", err)
			log.Printf("LNM058: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.StartSessionMonitor(session)

		return &ocpirpc.SessionCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcSessionResolver) SessionUpdated(reqCtx context.Context, input *ocpirpc.SessionUpdatedRequest) (*ocpirpc.SessionUpdatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		ctx := context.Background()
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			metrics.RecordError("LNM050", "Error retrieving session", err)
			log.Printf("LNM050: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.UpdateSession(session)

		return &ocpirpc.SessionUpdatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
