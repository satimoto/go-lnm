package session

import (
	"context"
	"crypto/sha256"
	"log"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/pkg/util"
)

func (r *SessionResolver) IssueSessionInvoice(ctx context.Context, user db.User, session db.Session, tokenAuthorization db.TokenAuthorization, invoiceParams util.InvoiceParams) *db.SessionInvoice {
	currencyRate, err := r.FerpService.GetRate(session.Currency)

	if err != nil {
		dbUtil.LogOnError("LSP054", "Error retrieving exchange rate", err)
		log.Printf("LSP054: Currency=%v", session.Currency)
		return nil
	}

	rateMsat := float64(currencyRate.RateMsat)
	invoiceParams = util.FillInvoiceRequestParams(invoiceParams, rateMsat)

	if !invoiceParams.TotalMsat.Valid {
		dbUtil.LogOnError("LSP116", "Error filling request params", err)
		log.Printf("LSP116: SessionUid=%v, Params=%#v", session.Uid, invoiceParams)
		return nil
	}

	preimage, err := lightningnetwork.RandomPreimage()

	if err != nil {
		dbUtil.LogOnError("LSP030", "Error creating invoice preimage", err)
		log.Printf("LSP030: SessionUid=%v", session.Uid)
		return nil
	}

	invoice, err := r.LightningService.AddInvoice(&lnrpc.Invoice{
		Memo:      session.Uid,
		RPreimage: preimage[:],
		ValueMsat: invoiceParams.TotalMsat.Int64,
	})

	if err != nil {
		dbUtil.LogOnError("LSP031", "Error creating lightning invoice", err)
		log.Printf("LSP031: Preimage=%v, ValueMsat=%v", preimage.String(), invoiceParams.TotalMsat.Int64)
		return nil
	}

	privateKey := secp.PrivKeyFromBytes(tokenAuthorization.SigningKey)
	hash := sha256.New()
	hash.Write([]byte(invoice.PaymentRequest))
	signature := ecdsa.Sign(privateKey, hash.Sum(nil))

	sessionInvoiceParams := param.NewCreateSessionInvoiceParams(session)
	sessionInvoiceParams.UserID = user.ID
	sessionInvoiceParams.CurrencyRate = currencyRate.Rate
	sessionInvoiceParams.CurrencyRateMsat = currencyRate.RateMsat
	sessionInvoiceParams.PriceFiat = invoiceParams.PriceFiat.Float64
	sessionInvoiceParams.PriceMsat = invoiceParams.PriceMsat.Int64
	sessionInvoiceParams.CommissionFiat = invoiceParams.CommissionFiat.Float64
	sessionInvoiceParams.CommissionMsat = invoiceParams.CommissionMsat.Int64
	sessionInvoiceParams.TaxFiat = invoiceParams.TaxFiat.Float64
	sessionInvoiceParams.TaxMsat = invoiceParams.TaxMsat.Int64
	sessionInvoiceParams.TotalFiat = invoiceParams.TotalFiat.Float64
	sessionInvoiceParams.TotalMsat = invoiceParams.TotalMsat.Int64
	sessionInvoiceParams.PaymentRequest = invoice.PaymentRequest
	sessionInvoiceParams.Signature = signature.Serialize()

	sessionInvoice, err := r.Repository.CreateSessionInvoice(ctx, sessionInvoiceParams)

	if err != nil {
		dbUtil.LogOnError("LSP003", "Could not create session invoice", err)
		log.Printf("LSP003: Params=%#v", sessionInvoiceParams)
		return nil
	}

	// TODO: handle notification failure
	r.SendSessionInvoiceNotification(user, session, sessionInvoice)

	return &sessionInvoice
}
