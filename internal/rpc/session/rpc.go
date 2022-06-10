package session

import (
	"context"
	"errors"
	"log"

	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ocpi-api/ocpirpc"
)

func (r *RpcSessionResolver) SessionCreated(ctx context.Context, input *ocpirpc.SessionCreatedRequest) (*ocpirpc.SessionCreatedResponse, error) {
	if input != nil {
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			util.LogOnError("LSP058", "Error retrieving cdr", err)
			log.Printf("LSP058: CdrUid=%v", input.SessionUid)
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.MonitorSession(ctx, session)

		return &ocpirpc.SessionCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
