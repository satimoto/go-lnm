package tokenauthorization

import (
	"context"
	"errors"

	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *RpcTokenAuthorizationResolver) TokenAuthorizationCreated(ctx context.Context, input *ocpirpc.TokenAuthorizationCreatedRequest) (*ocpirpc.TokenAuthorizationCreatedResponse, error) {
	if input != nil {
		// TODO: This RPC call should be handled asynchronously
		go r.TokenAuthorizationResolver.StartTokenAuthorizationMonitor(input.AuthorizationId)

		return &ocpirpc.TokenAuthorizationCreatedResponse{}, nil
	}

	return nil, errors.New("missing request")
}
