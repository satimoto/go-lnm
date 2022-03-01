package intercept

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type customMessageHandler func(*lnrpc.CustomMessage, string)

func (i *Intercept) SubscribeCustomMesssages() {
	lndMacaroon, err := base64.StdEncoding.DecodeString(os.Getenv("LND_MACAROON"))
	util.PanicOnError("Invalid LND Macaroon", err)

	macaroonCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(lndMacaroon))
	subscribeCustomMessagesClient, err := i.waitForSubscribeCustomMessagesClient(macaroonCtx, 0, 1000)
	util.PanicOnError("Error creating SubscribeCustomMessages client", err)

	for {
		if err := i.SubscribeCustomMesssage(subscribeCustomMessagesClient); err != nil {
			subscribeCustomMessagesClient, err = i.waitForSubscribeCustomMessagesClient(macaroonCtx, 100, 1000)
			util.PanicOnError("Error creating SubscribeCustomMessages client", err)
		}
	}
}

func (i *Intercept) SubscribeCustomMesssage(subscribeCustomMessagesClient lnrpc.Lightning_SubscribeCustomMessagesClient) error {
	customMessage, err := subscribeCustomMessagesClient.Recv()

	if err != nil {
		log.Printf("Error receiving custom message: %v", status.Code(err))

		return err
	}

	for index, handler := range i.customMessageHandlers {
		handler(customMessage, index)
	}

	return nil
}

func (i *Intercept) AddCustomMessageHandler(handler customMessageHandler) {
	index := strconv.FormatInt(time.Now().UnixNano(), 10)
	i.customMessageHandlers[index] = handler
}

func (i *Intercept) RemoveCustomMessageHandler(index string) {
	delete(i.customMessageHandlers, index)
}

func (i *Intercept) waitForSubscribeCustomMessagesClient(ctx context.Context, initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeCustomMessagesClient, err := i.LightningClient.SubscribeCustomMessages(ctx, &lnrpc.SubscribeCustomMessagesRequest{})

		if err == nil {
			return subscribeCustomMessagesClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeCustomMessages client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
