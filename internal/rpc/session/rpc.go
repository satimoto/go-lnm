package session

import (
	"context"
	"errors"
	"log"

	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcSessionResolver) SessionCreated(ctx context.Context, input *ocpirpc.SessionCreatedRequest) (*ocpirpc.SessionCreatedResponse, error) {
	if input != nil {
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			util.LogOnError("LSP058", "Error retrieving session", err)
			log.Printf("LSP058: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.StartSessionMonitor(ctx, session)

		return &ocpirpc.SessionCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcSessionResolver) SessionUpdated(ctx context.Context, input *ocpirpc.SessionUpdatedRequest) (*ocpirpc.SessionUpdatedResponse, error) {
	if input != nil {
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			util.LogOnError("LSP050", "Error retrieving session", err)
			log.Printf("LSP050: SessionUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.UpdateSession(ctx, session)

		return &ocpirpc.SessionUpdatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}