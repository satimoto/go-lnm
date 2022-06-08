package session

import (
	"context"
	"errors"

	"github.com/satimoto/go-ocpi-api/ocpirpc"
)

func (r *RpcSessionResolver) SessionCreated(ctx context.Context, input *ocpirpc.SessionCreatedRequest) (*ocpirpc.SessionCreatedResponse, error) {
	if input != nil {
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			return nil, errors.New("session not found")
		}

		go r.SessionResolver.MonitorSession(ctx, session)

		return &ocpirpc.SessionCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
