package psbtfund

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/satimoto/go-datastore/pkg/channelrequest"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/psbtfundingstate"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/service"
)

type PsbtFund interface {
	Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup)
	OpenChannel(ctx context.Context, request *lnrpc.OpenChannelRequest, channelRequest db.ChannelRequest) error
}

type PsbtFundService struct {
	LightningService           lightningnetwork.LightningNetwork
	ChannelRequestRepository   channelrequest.ChannelRequestRepository
	PsbtFundingStateRepository psbtfundingstate.PsbtFundingStateRepository
	mutex                      *sync.Mutex
	shutdownCtx                context.Context
	waitGroup                  *sync.WaitGroup
	nodeID                     int64
}

func NewService(repositoryService *db.RepositoryService, services *service.ServiceResolver) PsbtFund {
	return &PsbtFundService{
		LightningService:           services.LightningService,
		ChannelRequestRepository:   channelrequest.NewRepository(repositoryService),
		PsbtFundingStateRepository: psbtfundingstate.NewRepository(repositoryService),
		mutex:                      &sync.Mutex{},
	}
}

func (s *PsbtFundService) Start(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Psbt Fund service")
	s.nodeID = nodeID
	s.shutdownCtx = shutdownCtx
	s.waitGroup = waitGroup

	go s.handleUnfundedPsbtFundingStates()
}

func (s *PsbtFundService) OpenChannel(ctx context.Context, request *lnrpc.OpenChannelRequest, channelRequest db.ChannelRequest) error {
	s.lock()
	defer s.unlock()

	if request.FundingShim == nil {
		log.Printf("Psbt Fund open channel request")

		startPsbtMonitor := true
		psbtFundingState, err := s.PsbtFundingStateRepository.GetUnfundedPsbtFundingState(ctx, s.nodeID)

		request.FundingShim = &lnrpc.FundingShim{
			Shim: &lnrpc.FundingShim_PsbtShim{
				PsbtShim: &lnrpc.PsbtShim{
					PendingChanId: channelRequest.PendingChanID,
					BasePsbt:      psbtFundingState.Psbt,
					NoPublish:     true,
				},
			},
		}

		openChannelClient, err := s.LightningService.OpenChannel(request)

		if err != nil {
			util.LogOnError("LSP085", "Error opening channel", err)
			log.Printf("LSP085: Request=%#v", request)
			return errors.New("Error opening channel")
		}

		openStatusUpdate, err := s.waitForOpenStatusUpdate(ctx, openChannelClient)

		if err != nil {
			util.LogOnError("LSP086", "Error receiving open status update", err)
			return errors.New("Error receiving open status update")
		}

		update, ok := openStatusUpdate.Update.(*lnrpc.OpenStatusUpdate_PsbtFund)

		if !ok {
			util.LogOnError("LSP087", "Error expecting PsbtFund", err)
			log.Printf("LSP087: openStatusUpdate=%#v", openStatusUpdate)
			return errors.New("Error expecting PSBT funding update")
		}

		if len(psbtFundingState.BasePsbt) == 0 {
			psbtBatchTimeout := util.GetEnvInt32("PSBT_BATCH_TIMEOUT", 30)
			expiryDate := time.Now().Add(time.Duration(psbtBatchTimeout) * time.Second)
			createPsbtFundingStateParams := db.CreatePsbtFundingStateParams{
				NodeID:     s.nodeID,
				BasePsbt:   update.PsbtFund.Psbt,
				Psbt:       update.PsbtFund.Psbt,
				ExpiryDate: expiryDate,
			}

			createdPsbtFundingState, err := s.PsbtFundingStateRepository.CreatePsbtFundingState(ctx, createPsbtFundingStateParams)

			if err != nil {
				util.LogOnError("LSP087", "Error creating PSBT funding state", err)
				log.Printf("LSP087: Params=%#v", createPsbtFundingStateParams)
				return errors.New("Error creating PSBT funding state")
			}

			psbtFundingState = createdPsbtFundingState
		} else {
			updatePsbtFundingStateParams := param.NewUpdatePsbtFundingStateParams(psbtFundingState)
			updatePsbtFundingStateParams.Psbt = update.PsbtFund.Psbt

			updatedPsbtFundingState, err := s.PsbtFundingStateRepository.UpdatePsbtFundingState(ctx, updatePsbtFundingStateParams)

			if err != nil {
				util.LogOnError("LSP088", "Error updating PSBT funding state", err)
				log.Printf("LSP088: Params=%#v", updatePsbtFundingStateParams)
				return errors.New("Error updating PSBT funding state")
			}

			psbtFundingState = updatedPsbtFundingState
		}

		s.PsbtFundingStateRepository.SetPsbtFundingStateChannelRequest(ctx, db.SetPsbtFundingStateChannelRequestParams{
			PsbtFundingStateID: psbtFundingState.ID,
			ChannelRequestID:   channelRequest.ID,
		})

		if startPsbtMonitor {
			go s.waitForTimeout(psbtFundingState.ID, psbtFundingState.ExpiryDate)
		}

		return nil
	}

	return errors.New("OpenChannelRequest already contains FundingShim")
}

func (s *PsbtFundService) fundPsbt(psbtFundingStateID int64) {
	s.lock()
	defer s.unlock()

	// Get the PSBT funding state
	ctx := context.Background()
	psbtFundingState, err := s.PsbtFundingStateRepository.GetPsbtFundingState(ctx, psbtFundingStateID)

	if err != nil {
		util.LogOnError("LSP089", "Error retrieving PSBT funding state", err)
		log.Printf("LSP089: PsbtFundingStateID=%v", psbtFundingStateID)
		return
	}

	// Fund the PSBT
	// TODO: calculate sats/vbyte
	fundPsbtRequest := &walletrpc.FundPsbtRequest{
		Template: &walletrpc.FundPsbtRequest_Psbt{
			Psbt: psbtFundingState.Psbt,
		},
		Fees: &walletrpc.FundPsbtRequest_SatPerVbyte{
			SatPerVbyte: 2,
		},
		SpendUnconfirmed: true,
	}

	fundPsbtResponse, err := s.LightningService.FundPsbt(fundPsbtRequest)

	if err != nil {
		util.LogOnError("LSP090", "Error funding PSBT", err)
		log.Printf("LSP090: FundPsbtRequest=%#v", fundPsbtRequest)
		return
	}

	// Verify each pending channel with the funded PSBT
	channelRequests, err := s.PsbtFundingStateRepository.ListPsbtFundingStateChannelRequests(ctx, psbtFundingStateID)

	if err != nil {
		util.LogOnError("LSP091", "Error listing PSBT channel requests", err)
		log.Printf("LSP091: PsbtFundingStateID=%v", psbtFundingStateID)
		return
	}

	for _, channelRequest := range channelRequests {
		fundingTransitionMsg := &lnrpc.FundingTransitionMsg{
			Trigger: &lnrpc.FundingTransitionMsg_PsbtVerify{
				PsbtVerify: &lnrpc.FundingPsbtVerify{
					PendingChanId: channelRequest.PendingChanID,
					FundedPsbt:    fundPsbtResponse.FundedPsbt,
				},
			},
		}

		_, err := s.LightningService.FundingStateStep(fundingTransitionMsg)

		if err != nil {
			util.LogOnError("LSP092", "Error funding PSBT", err)
			log.Printf("LSP092: FundingTransitionMsg=%#v", fundingTransitionMsg)
		}
	}

	// Sign the PSBT
	finalizePsbtRequest := &walletrpc.FinalizePsbtRequest{
		FundedPsbt: fundPsbtResponse.FundedPsbt,
	}

	finalizePsbtResponse, err := s.LightningService.FinalizePsbt(finalizePsbtRequest)

	if err != nil {
		util.LogOnError("LSP093", "Error funding PSBT", err)
		log.Printf("LSP093: FundPsbtRequest=%#v", fundPsbtRequest)
		return
	}

	// Finalize each pending channel with the signed PSBT
	for _, channelRequest := range channelRequests {
		fundingTransitionMsg := &lnrpc.FundingTransitionMsg{
			Trigger: &lnrpc.FundingTransitionMsg_PsbtFinalize{
				PsbtFinalize: &lnrpc.FundingPsbtFinalize{
					PendingChanId: channelRequest.PendingChanID,
					SignedPsbt:    finalizePsbtResponse.SignedPsbt,
				},
			},
		}

		_, err := s.LightningService.FundingStateStep(fundingTransitionMsg)

		if err != nil {
			util.LogOnError("LSP094", "Error funding PSBT", err)
			log.Printf("LSP094: FundingTransitionMsg=%#v", fundingTransitionMsg)
		}
	}

	// Publish transaction
	transaction := &walletrpc.Transaction{
		TxHex: finalizePsbtResponse.RawFinalTx,
	}

	publishTransactionResponse, err := s.LightningService.PublishTransaction(transaction)

	if err != nil {
		util.LogOnError("LSP098", "Error publishing transaction", err)
		log.Printf("LSP098: Transaction=%#v", transaction)
		return
	}

	log.Printf("PublishTransaction: %v", publishTransactionResponse.PublishError)

	// Update PSBT funding state with funded and signed PSBT
	updatePsbtFundingStateParams := param.NewUpdatePsbtFundingStateParams(psbtFundingState)
	updatePsbtFundingStateParams.FundedPsbt = fundPsbtResponse.FundedPsbt
	updatePsbtFundingStateParams.SignedPsbt = finalizePsbtResponse.SignedPsbt
	updatePsbtFundingStateParams.SignedTx = finalizePsbtResponse.RawFinalTx

	_, err = s.PsbtFundingStateRepository.UpdatePsbtFundingState(ctx, updatePsbtFundingStateParams)

	if err != nil {
		util.LogOnError("LSP095", "Error updating PSBT funding state", err)
		log.Printf("LSP095: Params=%#v", updatePsbtFundingStateParams)
	}
}

func (s *PsbtFundService) lock() {
	log.Printf("PSBT thread locked")
	s.waitGroup.Add(1)
	s.mutex.Lock()
}

func (s *PsbtFundService) handleUnfundedPsbtFundingStates() {
	ctx := context.Background()
	psbtFundingStates, err := s.PsbtFundingStateRepository.ListUnfundedPsbtFundingStates(ctx, s.nodeID)

	if err != nil {
		util.LogOnError("LSP100", "Error listing unfunded PSBT funding states", err)
		log.Printf("LSP100: NodeID=%v", s.nodeID)
		return
	}

	for _, psbtFundingState := range psbtFundingStates {
		channelRequests, err := s.PsbtFundingStateRepository.ListPsbtFundingStateChannelRequests(ctx, psbtFundingState.ID)

		if err != nil {
			util.LogOnError("LSP101", "Error listing PSBT funding state channel requests", err)
			log.Printf("LSP101: PsbtFundingStateID=%v", psbtFundingState.ID)
			continue
		}

		for _, channelRequest := range channelRequests {
			channelRequestHtlcs, err := s.ChannelRequestRepository.ListChannelRequestHtlcs(ctx, channelRequest.ID)

			if err != nil {
				util.LogOnError("LSP102", "Error listing channel request HTLCs", err)
				log.Printf("LSP102: ChannelRequestID=%v", channelRequest.ID)
				continue
			}

			for _, channelRequestHtlc := range channelRequestHtlcs {
				s.ChannelRequestRepository.UpdateChannelRequestHtlcByCircuitKey(ctx, db.UpdateChannelRequestHtlcByCircuitKeyParams{
					ChanID:    channelRequestHtlc.ChanID,
					HtlcID:    channelRequestHtlc.HtlcID,
					IsSettled: false,
					IsFailed:  true,
				})
			}
		}

		updatePsbtFundingStateParams := param.NewUpdatePsbtFundingStateParams(psbtFundingState)
		updatePsbtFundingStateParams.IsFailed = true

		_, err = s.PsbtFundingStateRepository.UpdatePsbtFundingState(ctx, updatePsbtFundingStateParams)

		if err != nil {
			util.LogOnError("LSP103", "Error updating psbt funding state", err)
			log.Printf("LSP103: Params=%#v", updatePsbtFundingStateParams)
		}
	}
}

func (s *PsbtFundService) unlock() {
	s.mutex.Unlock()
	s.waitGroup.Done()
	log.Printf("PSBT thread unlocked")
}

func (s *PsbtFundService) waitForOpenStatusUpdate(ctx context.Context, client lnrpc.Lightning_OpenChannelClient) (*lnrpc.OpenStatusUpdate, error) {
	recvChan := make(chan *lnrpc.OpenStatusUpdate)
	errChan := make(chan error)

	go func() {
		openStatusUpdate, err := client.Recv()

		if err != nil {
			errChan <- err
			return
		}

		recvChan <- openStatusUpdate
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("Context cancelled")
	case err := <-errChan:
		return nil, err
	case openStatusUpdate := <-recvChan:
		return openStatusUpdate, nil
	}
}

func (s *PsbtFundService) waitForTimeout(psbtFundingStateID int64, expiry time.Time) {
	until := time.Until(expiry)
	log.Printf("PSBT Fund timeout (%v) set for %v: %v", time.Now(), expiry, until)

	select {
	case <-s.shutdownCtx.Done():
		log.Printf("Cancelling PSBT Fund timeout")
		return
	case <-time.After(until):
		s.fundPsbt(psbtFundingStateID)
	}
}
