package channel

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strconv"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/lsprpc"
	"github.com/satimoto/go-lsp/pkg/util"
)

func (r *RpcChannelResolver) OpenChannel(ctx context.Context, input *lsprpc.OpenChannelRequest) (*lsprpc.OpenChannelResponse, error) {
	if input != nil {
		maxFeeSatPerVByte := uint64(dbUtil.GetEnvInt32("PSBT_MAX_FEE_SAT_VBYTE", 64))
		satPerVByte := uint64(2)

		if estimateFeeResponse, err := r.LightningService.EstimateFee(&walletrpc.EstimateFeeRequest{ConfTarget: 6}); err == nil {
			satPerVByte = util.FeePerVByte(estimateFeeResponse.SatPerKw)
		}

		if satPerVByte > maxFeeSatPerVByte {
			metrics.RecordError("LSP154", "Error fees too high", errors.New("error fees too high"))
			log.Printf("LSP154: Sats/VByte=%v", satPerVByte)
			return nil, errors.New("error fees too high")
		}

		walletBalance, err := r.LightningService.WalletBalance(&lnrpc.WalletBalanceRequest{})

		if err != nil {
			metrics.RecordError("LSP109", "Error retreiving wallet balance", err)
			log.Printf("LSP109: OpenChannelRequest=%#v", input)
			return nil, errors.New("error retreiving wallet balance")
		}

		localFundingAmount := util.CalculateLocalFundingAmount(input.Amount)

		if localFundingAmount >= walletBalance.TotalBalance {
			// TODO: Report low balance
			log.Printf("LSP110: Error funding channel request")
			log.Printf("LSP110: LocalFundingAmount=%v TotalBalance=%v", localFundingAmount, walletBalance.TotalBalance)
			return nil, errors.New("error funding channel request")
		}

		alias, err := r.LightningService.AllocateAlias(&lnrpc.AllocateAliasRequest{})

		if err != nil {
			metrics.RecordError("LSP107", "Error allocating alias", err)
			log.Printf("LSP107: OpenChannelRequest=%#v", input)
			return nil, errors.New("error allocating alias")
		}

		pendingChanId := r.generatePendingChanId(ctx)
		shortChanID := lnwire.NewShortChanIDFromInt(alias.Scid)
		log.Printf("Allocating alias ShortChannelID: %v", shortChanID.String())

		baseFeeMsat := int64(dbUtil.GetEnvInt32("BASE_FEE_MSAT", 0))
		feeRatePpm := uint32(dbUtil.GetEnvInt32("FEE_RATE_PPM", 0))
		timeLockDelta := uint32(dbUtil.GetEnvInt32("TIME_LOCK_DELTA", 100))

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

func (r *RpcChannelResolver) ListChannels(ctx context.Context, input *lsprpc.ListChannelsRequest) (*lsprpc.ListChannelsResponse, error) {
	if input != nil {
		var channelIds []string
		listChannels, err := r.LightningService.ListChannels(&lnrpc.ListChannelsRequest{
			ActiveOnly: true,
			PublicOnly: true,
		})

		if err != nil {
			metrics.RecordError("LSP153", "Error listing channels", err)
			return nil, errors.New("error retreiving wallet balance")
		}

		for _, channel := range listChannels.Channels {
			channelIds = append(channelIds, strconv.FormatUint(channel.ChanId, 10))
		}

		return &lsprpc.ListChannelsResponse{
			ChannelIds: channelIds,
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
