package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/channelrequest"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (i *Intercept) SubscribeHtlcEvents() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	subscribeHtlcEventsClient, err := i.waitForSubscribeHtlcEventsClient(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating SubscribeHtlcEvents client", err)

	for {
		if err := i.SubscribeHtlcEvent(subscribeHtlcEventsClient); err != nil {
			subscribeHtlcEventsClient, err = i.waitForSubscribeHtlcEventsClient(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating SubscribeHtlcEvents client", err)
		}
	}
}

func (i *Intercept) SubscribeHtlcEvent(subscribeHtlcEventsClient routerrpc.Router_SubscribeHtlcEventsClient) error {
	htlcEvent, err := subscribeHtlcEventsClient.Recv()

	if err != nil {
		log.Printf("Error receiving HTLC event: %v", status.Code(err))

		return err
	}

	/** HTLC Event received.
	 *  Check that the event type is a Forward event and that is is successful.
	 *  Find the Channel Request HTLC by the circuit key params.
	 *  If the Channel Request HTLC exists, set it as settled.
	 */

	if htlcEvent.EventType == routerrpc.HtlcEvent_FORWARD {
		successEvent := htlcEvent.GetSettleEvent()
		forwardEvent := htlcEvent.GetForwardEvent()

		if forwardEvent != nil {
			log.Printf("Forward HTLC")
			log.Printf("IncomingAmtMsat: %v", forwardEvent.Info.IncomingAmtMsat)
			log.Printf("OutgoingAmtMsat: %v", forwardEvent.Info.OutgoingAmtMsat)
		}

		if successEvent != nil {
			ctx := context.Background()
			getChannelRequestHtlcByCircuitKeyParams := db.GetChannelRequestHtlcByCircuitKeyParams{
				ChanID: int64(htlcEvent.IncomingHtlcId),
				HtlcID: int64(htlcEvent.IncomingHtlcId),
			}

			if channelRequestHtlc, err := i.ChannelRequestResolver.Repository.GetChannelRequestHtlcByCircuitKey(ctx, getChannelRequestHtlcByCircuitKeyParams); err == nil {
				i.ChannelRequestResolver.Repository.UpdateChannelRequestHtlc(ctx, db.UpdateChannelRequestHtlcParams{
					ID:        channelRequestHtlc.ID,
					IsSettled: true,
				})

				unsettledChannelRequestHtlcs, _ := i.ChannelRequestResolver.Repository.ListUnsettledChannelRequestHtlcs(ctx, channelRequestHtlc.ChannelRequestID)

				if len(unsettledChannelRequestHtlcs) == 0 {
					// TODO: Add creating channel task to worker group
					if channelRequest, err := i.ChannelRequestResolver.Repository.GetChannelRequest(ctx, channelRequestHtlc.ChannelRequestID); err == nil {
						if pubkeyBytes, err := hex.DecodeString(channelRequest.Pubkey); err == nil {
							pushSat := int64(lnwire.MilliSatoshi(channelRequest.AmountMsat).ToSatoshis())
							openChannelRequest := &lnrpc.OpenChannelRequest{
								NodePubkey:         pubkeyBytes,
								LocalFundingAmount: 0,
								PushSat:            pushSat,
								TargetConf:         1,
								MinConfs:           0,
								Private:            true,
								SpendUnconfirmed:   true,
							}

							if channelPoint, err := i.LightningClient.OpenChannelSync(ctx, openChannelRequest); err == nil {
								updateChannelRequestParams := channelrequest.NewUpdateChannelRequestParams(channelRequest)
								updateChannelRequestParams.FundingTxID = channelPoint.GetFundingTxidBytes()
								updateChannelRequestParams.OutputIndex = util.SqlNullInt64(channelPoint.OutputIndex)

								i.ChannelRequestResolver.Repository.UpdateChannelRequest(ctx, updateChannelRequestParams)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (i *Intercept) waitForSubscribeHtlcEventsClient(ctx context.Context, initialDelay, retryDelay time.Duration) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeHtlcEventsClient, err := i.RouterClient.SubscribeHtlcEvents(ctx, &routerrpc.SubscribeHtlcEventsRequest{})

		if err == nil {
			return subscribeHtlcEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeHtlcEvents client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
