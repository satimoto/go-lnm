package session

import (
	"context"
	"errors"

	"github.com/satimoto/go-ocpi-api/ocpirpc/sessionrpc"
)

func (r *RpcSessionResolver) SessionStarted(ctx context.Context, input *sessionrpc.SessionStartedRequest) (*sessionrpc.SessionStartedResponse, error) {
	if input != nil {
		session, err := r.SessionResolver.Repository.GetSessionByUid(ctx, input.SessionUid)

		if err != nil {
			return nil, errors.New("Session not found")
		}

		go r.SessionResolver.MonitorSession(ctx, session)

		return &sessionrpc.SessionStartedResponse{}, nil
	}

	return nil, errors.New("Missing SessionStartedRequest")
}
