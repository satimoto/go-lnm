package rpc

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

func (r *RpcResolver) TestConnection(ctx context.Context, input *ocpirpc.TestConnectionRequest) (*ocpirpc.TestConnectionResponse, error) {
	if input != nil {
		ocpiService := ocpi.NewService(input.Addr)
		message := strconv.FormatInt(time.Now().Unix(), 16)
		testMessageReponse, err := ocpiService.TestMessage(ctx, &ocpirpc.TestMessageRequest{
			Message: message,
		})

		if err != nil {
			util.LogOnError("LSP056", "Error testing connection", err)
			log.Printf("LSP056: Addr=%v", input.Addr)
			return nil, errors.New("Connection test failed")
		}

		if testMessageReponse.Message != message {
			util.LogOnError("LSP057", "Error message response mismatch", err)
			log.Printf("LSP057: Message=%v, Response=%v", message, testMessageReponse.Message)
		}

		return &ocpirpc.TestConnectionResponse{Result: "OK"}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcResolver) TestMessage(ctx context.Context, input *ocpirpc.TestMessageRequest) (*ocpirpc.TestMessageResponse, error) {
	if input != nil {
		return &ocpirpc.TestMessageResponse{
			Message: input.Message,
		}, nil
	}

	return nil, errors.New("missing request")
}
