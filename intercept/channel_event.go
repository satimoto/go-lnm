package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (i *Intercept) SubscribeChannelEvents() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	subscribeChannelEventsClient, err := i.waitForSubscribeChannelEventsClient(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating SubscribeChannelEvents client", err)

	for {
		if err := i.SubscribeChannelEvent(subscribeChannelEventsClient); err != nil {
			subscribeChannelEventsClient, err = i.waitForSubscribeChannelEventsClient(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating SubscribeChannelEvents client", err)
		}
	}
}

func (i *Intercept) SubscribeChannelEvent(subscribeChannelEventsClient lnrpc.Lightning_SubscribeChannelEventsClient) error {
	channelEvent, err := subscribeChannelEventsClient.Recv()

	if err != nil {
		log.Printf("Error receiving channel event: %v", status.Code(err))

		return err
	}

	/** Channel Event received.
	 *  Find the Channel Request by the channel point params.
	 *  Update the Channel Request status depending on the event type.
	 */

	log.Printf("Channel Event: %v", channelEvent.Type)

	ctx := context.Background()

	switch channelEvent.Type {
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		pendingOpenChannel := channelEvent.GetPendingOpenChannel()
		log.Printf("Txid: %v", hex.EncodeToString(pendingOpenChannel.Txid))
		log.Printf("OutputIndex: %v", pendingOpenChannel.OutputIndex)

		i.ChannelRequestResolver.Repository.UpdateChannelRequestByChannelPoint(ctx, db.UpdateChannelRequestByChannelPointParams{
			FundingTxID: pendingOpenChannel.Txid,
			OutputIndex: util.SqlNullInt64(pendingOpenChannel.OutputIndex),
			Status:      db.ChannelRequestStatusOPENINGCHANNEL,
		})
		break
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		openChannel := channelEvent.GetOpenChannel()
		txid, outputIndex, _ := util.ConvertChannelPoint(openChannel.ChannelPoint)
		log.Printf("Txid: %v", hex.EncodeToString(txid))
		log.Printf("OutputIndex: %v", outputIndex)

		i.ChannelRequestResolver.Repository.UpdateChannelRequestByChannelPoint(ctx, db.UpdateChannelRequestByChannelPointParams{
			FundingTxID: txid,
			OutputIndex: util.SqlNullInt64(outputIndex),
			Status:      db.ChannelRequestStatusCOMPLETED,
		})
		break
	}

	return nil
}

func (i *Intercept) waitForSubscribeChannelEventsClient(ctx context.Context, initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeChannelEventsClient, err := i.LightningClient.SubscribeChannelEvents(ctx, &lnrpc.ChannelEventSubscription{})

		if err == nil {
			return subscribeChannelEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeChannelEvents client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
