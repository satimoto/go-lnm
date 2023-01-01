package session

import (
	"context"
	"errors"
	"log"

	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcSessionResolver) SessionCreated(ctx context.Context, input *ocpirpc.SessionCreatedRequest) (*ocpirpc.SessionCreatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			metrics.RecordError("LSP058", "Error retrieving session", err)
			log.Printf("LSP058: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.StartSessionMonitor(session)

		return &ocpirpc.SessionCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcSessionResolver) SessionUpdated(ctx context.Context, input *ocpirpc.SessionUpdatedRequest) (*ocpirpc.SessionUpdatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			metrics.RecordError("LSP050", "Error retrieving session", err)
			log.Printf("LSP050: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.UpdateSession(session)

		return &ocpirpc.SessionUpdatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
