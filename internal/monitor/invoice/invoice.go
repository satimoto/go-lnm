package invoice

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/session"
	"github.com/satimoto/go-lsp/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceMonitor struct {
	LightningService lightningnetwork.LightningNetwork
	InvoicesClient   lnrpc.Lightning_SubscribeInvoicesClient
	SessionResolver  *session.SessionResolver
}

func NewInvoiceMonitor(repositoryService *db.RepositoryService, lightningService lightningnetwork.LightningNetwork) *InvoiceMonitor {
	return &InvoiceMonitor{
		LightningService: lightningService,
		SessionResolver:  session.NewResolver(repositoryService),
	}
}

func (m *InvoiceMonitor) StartMonitor(ctx context.Context, waitGroup *sync.WaitGroup) {
	log.Printf("Starting up Invoices")
	invoiceChan := make(chan lnrpc.Invoice)

	go m.waitForInvoices(ctx, waitGroup, invoiceChan)
	go m.subscribeInvoiceInterceptions(invoiceChan)
}

func (m *InvoiceMonitor) handleInvoice(invoice lnrpc.Invoice) {
	log.Print("Invoice")
	log.Printf("PaymentRequest: %v", invoice.PaymentRequest)
	log.Printf("Settled: %v", invoice.Settled)

	/** Invoice received.
	 *  Check that the invoice is settled.
	 *  Find a Session Invoice that has a matching payment request.
	 *  Set the Session Invoice as settled.
	 */
	ctx := context.Background()

	if sessionInvoice, err := m.SessionResolver.Repository.GetSessionInvoiceByPaymentRequest(ctx, invoice.PaymentRequest); err == nil {
		if invoice.Settled {
			// Settle session invoice
			updateSessionInvoiceParams := session.NewUpdateSessionInvoiceParams(sessionInvoice)
			updateSessionInvoiceParams.Settled = invoice.Settled

			_, err = m.SessionResolver.Repository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)
			// TODO: Resume user tokens if all invoices are settled

			if err != nil {
				util.LogOnError("LSP027", "Error updating session invoice", err)
				log.Printf("LSP027: Params=%#v", updateSessionInvoiceParams)
			}
		} else {
			// Monitor expiry of invoice
			go m.waitForInvoiceExpiry(ctx, invoice)
		}
	}
}

func (m *InvoiceMonitor) subscribeInvoiceInterceptions(invoiceChan chan<- lnrpc.Invoice) {
	invoicesClient, err := m.waitForSubscribeInvoicesClient(0, 1000)
	util.PanicOnError("LSP020", "Error creating Invoices client", err)
	m.InvoicesClient = invoicesClient

	for {
		invoice, err := m.InvoicesClient.Recv()

		if err == nil {
			invoiceChan <- *invoice
		} else {
			m.InvoicesClient, err = m.waitForSubscribeInvoicesClient(100, 1000)
			util.PanicOnError("LSP021", "Error creating Invoices client", err)
		}
	}
}

func (m *InvoiceMonitor) waitForInvoices(ctx context.Context, waitGroup *sync.WaitGroup, invoiceChan chan lnrpc.Invoice) {
	waitGroup.Add(1)
	defer close(invoiceChan)
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down Invoices")
			return
		case invoice := <-invoiceChan:
			m.handleInvoice(invoice)
		}
	}
}

func (m *InvoiceMonitor) waitForInvoiceExpiry(ctx context.Context, invoice lnrpc.Invoice) {
	expiry := (time.Second * time.Duration(invoice.Expiry)) + time.Minute

	time.Sleep(expiry)

	if sessionInvoice, err := m.SessionResolver.Repository.GetSessionInvoiceByPaymentRequest(ctx, invoice.PaymentRequest); err == nil {
		if !sessionInvoice.Settled && !sessionInvoice.Expired {
			updateSessionInvoiceParams := session.NewUpdateSessionInvoiceParams(sessionInvoice)
			updateSessionInvoiceParams.Expired = true

			_, err = m.SessionResolver.Repository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

			if err != nil {
				util.LogOnError("LSP036", "Error updating session invoice", err)
				log.Printf("LSP036: Params=%#v", updateSessionInvoiceParams)
			}
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
