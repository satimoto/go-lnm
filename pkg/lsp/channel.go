package lsp

import (
	"context"

	"github.com/satimoto/go-lsp/lsprpc"
	"google.golang.org/grpc"
)

func (s *LspService) OpenChannel(ctx context.Context, in *lsprpc.OpenChannelRequest, opts ...grpc.CallOption) (*lsprpc.OpenChannelResponse, error) {
	return s.getChannelClient().OpenChannel(ctx, in, opts...)
}

func (s *LspService) ListChannels(ctx context.Context, in *lsprpc.ListChannelsRequest, opts ...grpc.CallOption) (*lsprpc.ListChannelsResponse, error) {
	return s.getChannelClient().ListChannels(ctx, in, opts...)
}

func (s *LspService) getChannelClient() lsprpc.ChannelServiceClient {
	if s.channelClient == nil {
		client := lsprpc.NewChannelServiceClient(s.clientConn)
		s.channelClient = &client
	}

	return *s.channelClient
}
