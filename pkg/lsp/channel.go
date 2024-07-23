package lsp

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-lnm/lsprpc"
	"google.golang.org/grpc"
)

func (s *LspService) OpenChannel(ctx context.Context, in *lsprpc.OpenChannelRequest, opts ...grpc.CallOption) (*lsprpc.OpenChannelResponse, error) {
	timerStart := time.Now()
	response, err := s.getChannelClient().OpenChannel(ctx, in, opts...)
	timerStop := time.Now()

	log.Printf("OpenChannel responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LspService) ListChannels(ctx context.Context, in *lsprpc.ListChannelsRequest, opts ...grpc.CallOption) (*lsprpc.ListChannelsResponse, error) {
	timerStart := time.Now()
	response, err := s.getChannelClient().ListChannels(ctx, in, opts...)
	timerStop := time.Now()

	log.Printf("ListChannels responded in %f seconds", timerStop.Sub(timerStart).Seconds())

	return response, err
}

func (s *LspService) getChannelClient() lsprpc.ChannelServiceClient {
	if s.channelClient == nil {
		client := lsprpc.NewChannelServiceClient(s.clientConn)
		s.channelClient = &client
	}

	return *s.channelClient
}
