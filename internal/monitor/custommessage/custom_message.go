package custommessage

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type customMessageHandler func(lnrpc.CustomMessage, string)

type CustomMessageMonitor struct {
	LightningService      lightningnetwork.LightningNetwork
	CustomMessagesClient  lnrpc.Lightning_SubscribeCustomMessagesClient
	CustomMessageHandlers map[string]customMessageHandler
}

func NewCustomMessageMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *CustomMessageMonitor {
	return &CustomMessageMonitor{
		LightningService:      lightningService,
		CustomMessageHandlers: make(map[string]customMessageHandler),
	}
}

func (m *CustomMessageMonitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Custom Messages")
	customMessageChan := make(chan lnrpc.CustomMessage)

	go m.waitForCustomMessages(ctx, waitGroup, customMessageChan)
	go m.subscribeCustomMessages(customMessageChan)
}

func (m *CustomMessageMonitor) AddHandler(handler customMessageHandler) {
	index := strconv.FormatInt(time.Now().UnixNano(), 10)
	m.CustomMessageHandlers[index] = handler
}

func (m *CustomMessageMonitor) RemoveHandler(index string) {
	delete(m.CustomMessageHandlers, index)
}

func (m *CustomMessageMonitor) handleCustomMessage(customMessage lnrpc.CustomMessage) {
	for index, handler := range m.CustomMessageHandlers {
		handler(customMessage, index)
	}
}

func (m *CustomMessageMonitor) subscribeCustomMessages(customMessageChan chan<- lnrpc.CustomMessage) {
	customMessagesClient, err := m.waitForSubscribeCustomMessagesClient(m.LightningService.GetMacaroonCtx(), 0, 1000)
	util.PanicOnError("LSP014", "Error creating Custom Messages client", err)

	m.CustomMessagesClient = customMessagesClient

	for {
		channelEvent, err := m.CustomMessagesClient.Recv()

		if err == nil {
			customMessageChan <- *channelEvent
		} else {
			m.CustomMessagesClient, err = m.waitForSubscribeCustomMessagesClient(m.LightningService.GetMacaroonCtx(), 100, 1000)
			util.PanicOnError("LSP015", "Error creating Custom Messages client", err)
		}
	}
}

func (m *CustomMessageMonitor) waitForCustomMessages(ctx context.Context, waitGroup *sync.WaitGroup, customMessageChan chan lnrpc.CustomMessage) {
	waitGroup.Add(1)
	defer close(customMessageChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Custom Messages")
			return
		case customMessage := <-customMessageChan:
			m.handleCustomMessage(customMessage)
		}
	}
}

func (m *CustomMessageMonitor) waitForSubscribeCustomMessagesClient(ctx context.Context, initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeCustomMessagesClient, err := m.LightningService.GetLightningClient().SubscribeCustomMessages(ctx, &lnrpc.SubscribeCustomMessagesRequest{})

		if err == nil {
			return subscribeCustomMessagesClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Custom Messages client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
