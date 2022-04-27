package session

import (
	"context"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/util"
)

func (r *SessionResolver) IssueLightningInvoice(ctx context.Context, session db.Session, amountFiat float64, amountMsat int64) {
	preimage, err := lightningnetwork.RandomPreimage()

	if err != nil {
		util.LogOnError("LSP030", "Error creating invoice preimage", err)
		log.Printf("LSP030: SessionUid=%v", session.Uid)
		return
	}

	invoice, err := r.LightningService.GetLightningClient().AddInvoice(r.LightningService.GetMacaroonCtx(), &lnrpc.Invoice{
		RPreimage: preimage[:],
		ValueMsat: amountMsat,
	})

	if err != nil {
		util.LogOnError("LSP031", "Error creating lightning invoice", err)
		log.Printf("LSP031: Preimage=%v, ValueMsat=%v", preimage.String(), amountMsat)
		return
	}

	sessionInvoiceParams := NewCreateSessionInvoiceParams(session.ID, amountFiat, amountMsat, invoice.PaymentRequest)
	sessionInvoice, err := r.Repository.CreateSessionInvoice(ctx, sessionInvoiceParams)

	if err != nil {
		util.LogOnError("LSP003", "Could not create session invoice", err)
		log.Printf("LSP003: Params=%#v", sessionInvoiceParams)
		return
	}

	// TODO: Send user device notification
	r.SendNotification(session, sessionInvoice) 
}
