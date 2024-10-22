package invoice

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/service"
	"github.com/satimoto/go-lnm/internal/session"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceMonitor struct {
	LightningService lightningnetwork.LightningNetwork
	InvoicesClient   lnrpc.Lightning_SubscribeInvoicesClient
	SessionResolver  *session.SessionResolver
	nodeID           int64
}

func NewInvoiceMonitor(repositoryService *db.RepositoryService, services *service.ServiceResolver) *InvoiceMonitor {
	return &InvoiceMonitor{
		LightningService: services.LightningService,
		SessionResolver:  session.NewResolver(repositoryService, services),
	}
}

func (m *InvoiceMonitor) StartMonitor(nodeID int64, shutdownCtx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Invoices")
	invoiceChan := make(chan lnrpc.Invoice)

	m.nodeID = nodeID
	go m.waitForInvoices(shutdownCtx, waitGroup, invoiceChan)
	go m.subscribeInvoiceInterceptions(invoiceChan)
}

func (m *InvoiceMonitor) handleInvoice(invoice lnrpc.Invoice) {
	settled := invoice.State == lnrpc.Invoice_SETTLED

	log.Print("Invoice")
	log.Printf("PaymentRequest: %v", invoice.PaymentRequest)
	log.Printf("Settled: %v", settled)

	/** Invoice received.
	 *  Check that the invoice is settled.
	 *  Find a Session Invoice that has a matching payment request.
	 *  Set the Session Invoice as settled.
	 *  Get users unsettled session invoices, if all are settled then unlock tokens
	 */
	ctx := context.Background()

	if sessionInvoice, err := m.SessionResolver.Repository.GetSessionInvoiceByPaymentRequest(ctx, invoice.PaymentRequest); err == nil {
		if settled {
			// Settle session invoice
			updateSessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
			updateSessionInvoiceParams.IsSettled = true

			_, err = m.SessionResolver.Repository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

			if err != nil {
				metrics.RecordError("LNM027", "Error updating session invoice", err)
				log.Printf("LNM027: Params=%#v", updateSessionInvoiceParams)
				return
			}

			// Metrics: Increment number of settled session invoices
			metricSessionInvoicesSettledTotal.Inc()

			// Get the user from the session ID
			user, err := m.SessionResolver.UserResolver.Repository.GetUserBySessionID(ctx, sessionInvoice.SessionID)

			if err != nil {
				metrics.RecordError("LNM039", "Error retrieving session user", err)
				log.Printf("LNM039: SessionID=%v", sessionInvoice.SessionID)
				return
			}

			// List users unsettled session invoices
			sessionInvoices, err := m.SessionResolver.Repository.ListSessionInvoicesByUserID(ctx, db.ListSessionInvoicesByUserIDParams{
				ID:        user.ID,
				IsSettled: false,
				IsExpired: false,
			})

			if err != nil {
				metrics.RecordError("LNM040", "Error retrieving user unsettled session invoices", err)
				log.Printf("LNM040: SessionID=%v, UserID=%v", sessionInvoice.SessionID, user.ID)
				return
			}

			// If there are no unsettled invoices then unlock user tokens
			if len(sessionInvoices) == 0 {
				err = m.SessionResolver.UserResolver.UnrestrictUser(ctx, user)

				if err != nil {
					metrics.RecordError("LNM041", "Error unrestricting user", err)
					log.Printf("LNM041: SessionID=%v, UserID=%v", sessionInvoice.SessionID, user.ID)
				}
			}
		}
	}
}

func (m *InvoiceMonitor) subscribeInvoiceInterceptions(invoiceChan chan<- lnrpc.Invoice) {
	invoicesClient, err := m.waitForSubscribeInvoicesClient(0, 1000)
	util.PanicOnError("LNM020", "Error creating Invoices client", err)
	m.InvoicesClient = invoicesClient

	for {
		invoice, err := m.InvoicesClient.Recv()

		if err == nil {
			invoiceChan <- *invoice
		} else {
			m.InvoicesClient, err = m.waitForSubscribeInvoicesClient(100, 1000)
			util.PanicOnError("LNM021", "Error creating Invoices client", err)
		}
	}
}

func (m *InvoiceMonitor) waitForInvoices(shutdownCtx context.Context, waitGroup *sync.WaitGroup, invoiceChan chan lnrpc.Invoice) {
	waitGroup.Add(1)
	defer close(invoiceChan)
	defer waitGroup.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Shutting down Invoices")
			return
		case invoice := <-invoiceChan:
			m.handleInvoice(invoice)
		}
	}
}

func (m *InvoiceMonitor) waitForSubscribeInvoicesClient(initialDelay, retryDelay time.Duration) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	for {
		if initialDelay > 0 {
			time.Sleep(retryDelay * time.Millisecond)
		}

		subscribeInvoicesClient, err := m.LightningService.SubscribeInvoices(&lnrpc.InvoiceSubscription{})

		if err == nil {
			return subscribeInvoicesClient, nil
		} else if status.Code(err) != codes.Unavailable {
			return nil, err
		}

		log.Print("Waiting for Invoices client")
		time.Sleep(retryDelay * time.Millisecond)
	}
}
