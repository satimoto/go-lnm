package rpc

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-ocpi/ocpirpc"
	"github.com/satimoto/go-ocpi/pkg/ocpi"
)

func (r *RpcResolver) TestConnection(reqCtx context.Context, input *ocpirpc.TestConnectionRequest) (*ocpirpc.TestConnectionResponse, error) {
	if input != nil {
		ctx := context.Background()
		ocpiService := ocpi.NewService(input.Addr)
		message := strconv.FormatInt(time.Now().Unix(), 16)
		testMessageReponse, err := ocpiService.TestMessage(ctx, &ocpirpc.TestMessageRequest{
			Message: message,
		})

		if err != nil {
			metrics.RecordError("LSP056", "Error testing connection", err)
			log.Printf("LSP056: Addr=%v", input.Addr)
			return nil, errors.New("Connection test failed")
		}

		if testMessageReponse.Message != message {
			metrics.RecordError("LSP057", "Error message response mismatch", err)
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
