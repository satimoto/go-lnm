package channel

import (
	"context"
	"crypto/rand"
	"errors"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/lsprpc"
)

func (r *RpcChannelResolver) OpenChannel(ctx context.Context, input *lsprpc.OpenChannelRequest) (*lsprpc.OpenChannelResponse, error) {
	if input != nil {
		// TODO: Check funds are available to create this channel
		pendingChanId := r.generatePendingChanId(ctx)

		alias, err := r.LightningService.AllocateAlias(&lnrpc.AllocateAliasRequest{})

		if err != nil {
			util.LogOnError("LSP107", "Error allocating alias", err)
			log.Printf("LSP107: OpenChannelRequest=%#v", input)
			return nil, errors.New("error allocating alias")
		}

		baseFeeMsat := int64(util.GetEnvInt32("BASE_FEE_MSAT", 0))
		feeRatePpm := uint32(util.GetEnvInt32("FEE_RATE_PPM", 10))
		timeLockDelta := uint32(util.GetEnvInt32("TIME_LOCK_DELTA", 100))

		return &lsprpc.OpenChannelResponse{
			PendingChanId:             pendingChanId,
			Scid:                      alias.Scid,
			FeeBaseMsat:               baseFeeMsat,
			FeeProportionalMillionths: feeRatePpm,
			CltvExpiryDelta:           timeLockDelta,
		}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcChannelResolver) generatePendingChanId(ctx context.Context) []byte {
	pendingChanId := make([]byte, 32)

	for {
		if _, err := rand.Read(pendingChanId); err == nil {
			if _, err := r.ChannelRequestResolver.Repository.GetChannelRequestByPendingChanId(ctx, pendingChanId); err != nil {
				return pendingChanId
			}
		}
	}
}
