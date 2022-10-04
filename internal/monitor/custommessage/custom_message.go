package custommessage

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomMessageHandler func(lnrpc.CustomMessage, string)

type CustomMessageMonitor struct {
	LightningService      lightningnetwork.LightningNetwork
	CustomMessagesClient  lnrpc.Lightning_SubscribeCustomMessagesClient
	CustomMessageHandlers map[string]CustomMessageHandler
	nodeID                int64
}

func NewCustomMessageMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *CustomMessageMonitor {
	return &CustomMessageMonitor{
		LightningService:      services.LightningService,
		CustomMessageHandlers: make(map[string]CustomMessageHandler),
	}
}

func (m *CustomMessageMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Custom Messages")
	customMessageChan := make(chan lnrpc.CustomMessage)

	m.nodeID = nodeID
	go m.waitForCustomMessages(shutdownCtx, waitGroup, customMessageChan)
	go m.subscribeCustomMessages(customMessageChan)
}

func (m *CustomMessageMonitor) AddHandler(handler CustomMessageHandler) {
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
	customMessagesClient, err := m.waitForSubscribeCustomMessagesClient(0, 1000)
	util.PanicOnError("LSP014", "Error creating Custom Messages client", err)

	m.CustomMessagesClient = customMessagesClient

	for {
		channelEvent, err := m.CustomMessagesClient.Recv()

		if err == nil {
			customMessageChan <- *channelEvent
		} else {
			m.CustomMessagesClient, err = m.waitForSubscribeCustomMessagesClient(100, 1000)
			util.PanicOnError("LSP015", "Error creating Custom Messages client", err)
		}
	}
}

func (m *CustomMessageMonitor) waitForCustomMessages(shutdownCtx context.Context, waitGroup *sync.WaitGroup, customMessageChan chan lnrpc.CustomMessage) {
	waitGroup.Add(1)
	defer close(customMessageChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Custom Messages")
			return
		case customMessage := <-customMessageChan:
			m.handleCustomMessage(customMessage)
		}
	}
}

func (m *CustomMessageMonitor) waitForSubscribeCustomMessagesClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeCustomMessagesClient, err := m.LightningService.SubscribeCustomMessages(&lnrpc.SubscribeCustomMessagesRequest{})

		if err == nil {
			return subscribeCustomMessagesClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Custom Messages client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
