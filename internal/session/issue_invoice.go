package session

import (
	"context"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
)

func (r *SessionResolver) IssueLightningInvoice(ctx context.Context, user db.User, session db.Session, invoiceAmount float64, commissionAmount float64, taxAmount float64) {
	currencyRate, err := r.FerpService.GetRate(session.Currency)

	if err != nil {
		util.LogOnError("LSP054", "Error retrieving exchange rate", err)
		log.Printf("LSP054: Currency=%v", session.Currency)
		return
	}

	rateMsat := float64(currencyRate.RateMsat)
	amountMsat := int64(invoiceAmount * rateMsat)
	commissionMsat := int64(commissionAmount * rateMsat)
	taxMsat := int64(taxAmount * rateMsat)

	preimage, err := lightningnetwork.RandomPreimage()

	if err != nil {
		util.LogOnError("LSP030", "Error creating invoice preimage", err)
		log.Printf("LSP030: SessionUid=%v", session.Uid)
		return
	}

	invoice, err := r.LightningService.AddInvoice(&lnrpc.Invoice{
		RPreimage: preimage[:],
		ValueMsat: amountMsat,
	})

	if err != nil {
		util.LogOnError("LSP031", "Error creating lightning invoice", err)
		log.Printf("LSP031: Preimage=%v, ValueMsat=%v", preimage.String(), amountMsat)
		return
	}

	sessionInvoiceParams := param.NewCreateSessionInvoiceParams(session)
	sessionInvoiceParams.AmountFiat = invoiceAmount
	sessionInvoiceParams.AmountMsat = amountMsat
	sessionInvoiceParams.CommissionFiat = commissionAmount
	sessionInvoiceParams.CommissionMsat = commissionMsat
	sessionInvoiceParams.TaxFiat = taxAmount
	sessionInvoiceParams.TaxMsat = taxMsat
	sessionInvoiceParams.PaymentRequest = invoice.PaymentRequest

	sessionInvoice, err := r.Repository.CreateSessionInvoice(ctx, sessionInvoiceParams)

	if err != nil {
		util.LogOnError("LSP003", "Could not create session invoice", err)
		log.Printf("LSP003: Params=%#v", sessionInvoiceParams)
		return
	}

	// TODO: Send user device notification
	r.SendNotification(user, session, sessionInvoice)
}
