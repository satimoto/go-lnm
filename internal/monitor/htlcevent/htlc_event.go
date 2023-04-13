package htlcevent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/routingevent"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/ferp"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HtlcEventMonitor struct {
	FerpService              ferp.Ferp
	LightningService         lightningnetwork.LightningNetwork
	HtlcEventsClient         routerrpc.Router_SubscribeHtlcEventsClient
	RoutingEventRepository   routingevent.RoutingEventRepository
	accountingCurrency       string
	nodeID                   int64
}

func NewHtlcEventMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *HtlcEventMonitor {
	return &HtlcEventMonitor{
		FerpService:              services.FerpService,
		LightningService:         services.LightningService,
		RoutingEventRepository:   routingevent.NewRepository(repositoryService),
	}
}

func (m *HtlcEventMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Htlc Events")
	htlcEventChan := make(chan routerrpc.HtlcEvent)

	m.accountingCurrency = util.GetEnv("ACCOUNTING_CURRENCY", "EUR")
	m.nodeID = nodeID

	go m.waitForHtlcEvents(shutdownCtx, waitGroup, htlcEventChan)
	go m.subscribeHtlcEventInterceptions(htlcEventChan)
}

func (m *HtlcEventMonitor) handleHtlcEvent(htlcEvent routerrpc.HtlcEvent) {
	/** HTLC Event received.
	 *  Check that the event type is a Forward event and that it is successful.
	 *  Find the Channel Request HTLC by the circuit key params.
	 *  If the Channel Request HTLC exists, set it as settled.
	 */

	if htlcEvent.EventType == routerrpc.HtlcEvent_FORWARD {
		ctx := context.Background()

		switch htlcEvent.Event.(type) {
		case *routerrpc.HtlcEvent_ForwardEvent:
			m.handleForwardHtlcEvent(ctx, htlcEvent)
		case *routerrpc.HtlcEvent_ForwardFailEvent:
			m.handleForwardFailHtlcEvent(ctx, htlcEvent)
		case *routerrpc.HtlcEvent_LinkFailEvent:
			m.handleLinkFailHtlcEvent(ctx, htlcEvent)
		case *routerrpc.HtlcEvent_SettleEvent:
			m.handleSettleHtlcEvent(ctx, htlcEvent)
		}
	}
}

func (m *HtlcEventMonitor) handleForwardHtlcEvent(ctx context.Context, htlcEvent routerrpc.HtlcEvent) {
	forwardEvent := htlcEvent.GetForwardEvent()

	currencyRate, err := m.FerpService.GetRate(m.accountingCurrency)

	if err != nil {
		metrics.RecordError("LNM071", "Error getting FERP rate", err)
		log.Printf("LNM071: Currency=%v", m.accountingCurrency)
		return
	}

	feeMsat := int64(forwardEvent.Info.IncomingAmtMsat) - int64(forwardEvent.Info.OutgoingAmtMsat)
	createRoutingEventParams := db.CreateRoutingEventParams{
		NodeID:           m.nodeID,
		EventType:        db.RoutingEventTypeFORWARD,
		EventStatus:      db.RoutingEventStatusINFLIGHT,
		Currency:         m.accountingCurrency,
		CurrencyRate:     currencyRate.Rate,
		CurrencyRateMsat: currencyRate.RateMsat,
		IncomingChanID:   int64(htlcEvent.IncomingChannelId),
		IncomingHtlcID:   int64(htlcEvent.IncomingHtlcId),
		IncomingFiat:     float64(forwardEvent.Info.IncomingAmtMsat) / float64(currencyRate.RateMsat),
		IncomingMsat:     int64(forwardEvent.Info.IncomingAmtMsat),
		OutgoingChanID:   int64(htlcEvent.OutgoingChannelId),
		OutgoingHtlcID:   int64(htlcEvent.IncomingHtlcId),
		OutgoingFiat:     float64(forwardEvent.Info.OutgoingAmtMsat) / float64(currencyRate.RateMsat),
		OutgoingMsat:     int64(forwardEvent.Info.OutgoingAmtMsat),
		FeeFiat:          float64(feeMsat) / float64(currencyRate.RateMsat),
		FeeMsat:          feeMsat,
		LastUpdated:      time.Unix(0, int64(htlcEvent.TimestampNs)),
	}

	_, err = m.RoutingEventRepository.CreateRoutingEvent(ctx, createRoutingEventParams)

	if err != nil {
		metrics.RecordError("LNM072", "Error creating routing event", err)
		log.Printf("LNM072: Params=%#v", createRoutingEventParams)
	}
}

func (m *HtlcEventMonitor) handleForwardFailHtlcEvent(ctx context.Context, htlcEvent routerrpc.HtlcEvent) {
	incomingChannelId := int64(htlcEvent.IncomingChannelId)
	incomingHtlcId := int64(htlcEvent.IncomingHtlcId)

	// Update routing event
	updateRoutingEventParams := db.UpdateRoutingEventParams{
		EventStatus:    db.RoutingEventStatusFORWARDFAIL,
		IncomingChanID: incomingChannelId,
		IncomingHtlcID: incomingHtlcId,
		OutgoingChanID: int64(htlcEvent.OutgoingChannelId),
		OutgoingHtlcID: int64(htlcEvent.IncomingHtlcId),
		LastUpdated:    time.Unix(0, int64(htlcEvent.TimestampNs)),
	}

	routingEvent, err := m.RoutingEventRepository.UpdateRoutingEvent(ctx, updateRoutingEventParams)

	if err != nil {
		metrics.RecordError("LNM081", "Error updating routing event", err)
		log.Printf("LNM081: Params=%#v", updateRoutingEventParams)
	}

	// Metrics: Increment number of failed routing events
	metricRoutingEventsFailed.Inc()

	if err == nil {
		metricRoutingEventsFailedFeeFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.FeeFiat)
		metricRoutingEventsFailedFeeSatoshis.Add(float64(routingEvent.FeeMsat / 1000))
		metricRoutingEventsFailedTotalFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.OutgoingFiat)
		metricRoutingEventsFailedTotalSatoshis.Add(float64(routingEvent.OutgoingMsat / 1000))
	}
}

func (m *HtlcEventMonitor) handleLinkFailHtlcEvent(ctx context.Context, htlcEvent routerrpc.HtlcEvent) {
	linkFailEvent := htlcEvent.GetLinkFailEvent()
	incomingChannelId := int64(htlcEvent.IncomingChannelId)
	incomingHtlcId := int64(htlcEvent.IncomingHtlcId)

	// Update routing event
	updateRoutingEventParams := db.UpdateRoutingEventParams{
		EventStatus:    db.RoutingEventStatusLINKFAIL,
		IncomingChanID: incomingChannelId,
		IncomingHtlcID: incomingHtlcId,
		OutgoingChanID: int64(htlcEvent.OutgoingChannelId),
		OutgoingHtlcID: int64(htlcEvent.IncomingHtlcId),
		WireFailure:    util.SqlNullInt32(linkFailEvent.WireFailure),
		FailureDetail:  util.SqlNullInt32(linkFailEvent.FailureDetail),
		FailureString:  util.SqlNullString(linkFailEvent.FailureString),
		LastUpdated:    time.Unix(0, int64(htlcEvent.TimestampNs)),
	}

	routingEvent, err := m.RoutingEventRepository.UpdateRoutingEvent(ctx, updateRoutingEventParams)

	if err != nil {
		createRoutingEventParams := db.CreateRoutingEventParams{
			NodeID:           m.nodeID,
			EventType:        db.RoutingEventTypeFORWARD,
			EventStatus:      db.RoutingEventStatusLINKFAIL,
			Currency:         m.accountingCurrency,
			CurrencyRate:     0,
			CurrencyRateMsat: 0,
			IncomingChanID:   int64(htlcEvent.IncomingChannelId),
			IncomingHtlcID:   int64(htlcEvent.IncomingHtlcId),
			IncomingFiat:     0,
			IncomingMsat:     0,
			OutgoingChanID:   int64(htlcEvent.OutgoingChannelId),
			OutgoingHtlcID:   int64(htlcEvent.IncomingHtlcId),
			OutgoingFiat:     0,
			OutgoingMsat:     0,
			FeeFiat:          0,
			FeeMsat:          0,
			WireFailure:      util.SqlNullInt32(linkFailEvent.WireFailure),
			FailureDetail:    util.SqlNullInt32(linkFailEvent.FailureDetail),
			FailureString:    util.SqlNullString(linkFailEvent.FailureString),
			LastUpdated:      time.Unix(0, int64(htlcEvent.TimestampNs)),
		}

		routingEvent, err = m.RoutingEventRepository.CreateRoutingEvent(ctx, createRoutingEventParams)

		if err != nil {
			metrics.RecordError("LNM082", "Error creating routing event", err)
			log.Printf("LNM082: Params=%#v", createRoutingEventParams)
		}
	}

	// Metrics: Increment number of failed routing events
	metricRoutingEventsFailed.Inc()

	if err == nil {
		metricRoutingEventsFailedFeeFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.FeeFiat)
		metricRoutingEventsFailedFeeSatoshis.Add(float64(routingEvent.FeeMsat / 1000))
		metricRoutingEventsFailedTotalFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.OutgoingFiat)
		metricRoutingEventsFailedTotalSatoshis.Add(float64(routingEvent.OutgoingMsat / 1000))
	}
}

func (m *HtlcEventMonitor) handleSettleHtlcEvent(ctx context.Context, htlcEvent routerrpc.HtlcEvent) {
	incomingChannelId := int64(htlcEvent.IncomingChannelId)
	incomingHtlcId := int64(htlcEvent.IncomingHtlcId)

	// Update routing event
	updateRoutingEventParams := db.UpdateRoutingEventParams{
		EventStatus:    db.RoutingEventStatusSETTLE,
		IncomingChanID: incomingChannelId,
		IncomingHtlcID: incomingHtlcId,
		OutgoingChanID: int64(htlcEvent.OutgoingChannelId),
		OutgoingHtlcID: int64(htlcEvent.IncomingHtlcId),
		LastUpdated:    time.Unix(0, int64(htlcEvent.TimestampNs)),
	}

	routingEvent, err := m.RoutingEventRepository.UpdateRoutingEvent(ctx, updateRoutingEventParams)

	if err != nil {
		metrics.RecordError("LNM083", "Error updating routing event", err)
		log.Printf("LNM083: Params=%#v", updateRoutingEventParams)
	}

	// Metrics: Increment number of settled routing events
	metricRoutingEventsSettled.Inc()

	if err == nil {
		metricRoutingEventsSettledFeeFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.FeeFiat)
		metricRoutingEventsSettledFeeSatoshis.Add(float64(routingEvent.FeeMsat / 1000))
		metricRoutingEventsSettledTotalFiat.WithLabelValues(routingEvent.Currency).Add(routingEvent.OutgoingFiat)
		metricRoutingEventsSettledTotalSatoshis.Add(float64(routingEvent.OutgoingMsat / 1000))
	}
}

func (m *HtlcEventMonitor) subscribeHtlcEventInterceptions(htlcEventChan chan<- routerrpc.HtlcEvent) {
	htlcEventsClient, err := m.waitForSubscribeHtlcEventsClient(0, 1000)
	util.PanicOnError("LNM018", "Error creating Htlc Events client", err)
	m.HtlcEventsClient = htlcEventsClient

	for {
		htlcInterceptRequest, err := m.HtlcEventsClient.Recv()

		if err == nil {
			htlcEventChan <- *htlcInterceptRequest
		} else {
			m.HtlcEventsClient, err = m.waitForSubscribeHtlcEventsClient(100, 1000)
			util.PanicOnError("LNM019", "Error creating Htlc Events client", err)
		}
	}
}

func (m *HtlcEventMonitor) waitForHtlcEvents(shutdownCtx context.Context, waitGroup *sync.WaitGroup, htlcEventChan chan routerrpc.HtlcEvent) {
	waitGroup.Add(1)
	defer close(htlcEventChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Htlc Events")
			return
		case htlcInterceptRequest := <-htlcEventChan:
			m.handleHtlcEvent(htlcInterceptRequest)
		}
	}
}

func (m *HtlcEventMonitor) waitForSubscribeHtlcEventsClient(initialDelay, retryDelay time.Duration) (routerrpc.Router_SubscribeHtlcEventsClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeHtlcEventsClient, err := m.LightningService.SubscribeHtlcEvents(&routerrpc.SubscribeHtlcEventsRequest{})

		if err == nil {
			return subscribeHtlcEventsClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for SubscribeHtlcEvents client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
