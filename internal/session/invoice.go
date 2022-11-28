package session

import (
	"context"
	"crypto/sha256"
	"log"
	"time"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/pkg/util"
)

func (r *SessionResolver) IssueSessionInvoice(ctx context.Context, user db.User, session db.Session, tokenAuthorization db.TokenAuthorization, invoiceParams util.InvoiceParams) *db.SessionInvoice {
	currencyRate, err := r.FerpService.GetRate(session.Currency)

	if err != nil {
		metrics.RecordError("LSP054", "Error retrieving exchange rate", err)
		log.Printf("LSP054: Currency=%v", session.Currency)
		return nil
	}

	rateMsat := float64(currencyRate.RateMsat)
	invoiceParams = util.FillInvoiceRequestParams(invoiceParams, rateMsat)

	if !invoiceParams.TotalMsat.Valid {
		metrics.RecordError("LSP116", "Error filling request params", err)
		log.Printf("LSP116: SessionUid=%v, Params=%#v", session.Uid, invoiceParams)
		return nil
	}

	preimage, err := lightningnetwork.RandomPreimage()

	if err != nil {
		metrics.RecordError("LSP030", "Error creating invoice preimage", err)
		log.Printf("LSP030: SessionUid=%v", session.Uid)
		return nil
	}

	invoice, err := r.LightningService.AddInvoice(&lnrpc.Invoice{
		Memo:      session.Uid,
		RPreimage: preimage[:],
		ValueMsat: invoiceParams.TotalMsat.Int64,
	})

	if err != nil {
		metrics.RecordError("LSP031", "Error creating lightning invoice", err)
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
		metrics.RecordError("LSP003", "Could not create session invoice", err)
		log.Printf("LSP003: Params=%#v", sessionInvoiceParams)
		return nil
	}

	// Metrics
	metricSessionInvoicesTotal.Inc()
	metricSessionInvoicesCommissionFiat.WithLabelValues(session.Currency).Add(sessionInvoice.CommissionFiat)
	metricSessionInvoicesCommissionSatoshis.Add(float64(sessionInvoice.CommissionMsat / 1000))
	metricSessionInvoicesPriceFiat.WithLabelValues(session.Currency).Add(sessionInvoice.PriceFiat)
	metricSessionInvoicesPriceSatoshis.Add(float64(sessionInvoice.PriceMsat / 1000))
	metricSessionInvoicesTaxFiat.WithLabelValues(session.Currency).Add(sessionInvoice.TaxFiat)
	metricSessionInvoicesTaxSatoshis.Add(float64(sessionInvoice.TaxMsat / 1000))
	metricSessionInvoicesTotalFiat.WithLabelValues(session.Currency).Add(sessionInvoice.TotalFiat)
	metricSessionInvoicesTotalSatoshis.Add(float64(sessionInvoice.TotalMsat / 1000))

	// TODO: handle notification failure
	r.SendSessionInvoiceNotification(user, session, sessionInvoice)

	go r.waitForInvoiceExpiry(invoice.PaymentRequest)

	return &sessionInvoice
}

func (r *SessionResolver) waitForInvoiceExpiry(paymentRequest string) {
	payReqParams := &lnrpc.PayReqString{PayReq: paymentRequest}
	expiry := int64(3600)

	if payReqResponse, err := r.LightningService.DecodePayReq(payReqParams); err == nil {
		expiry = payReqResponse.Expiry
	}

	ctx := context.Background()
	timeout := (time.Second * time.Duration(expiry)) + time.Minute

	time.Sleep(timeout)

	if sessionInvoice, err := r.Repository.GetSessionInvoiceByPaymentRequest(ctx, paymentRequest); err == nil {
		if !sessionInvoice.IsSettled && !sessionInvoice.IsExpired {
			updateSessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
			updateSessionInvoiceParams.IsExpired = true

			_, err = r.Repository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

			if err != nil {
				metrics.RecordError("LSP036", "Error updating session invoice", err)
				log.Printf("LSP036: Params=%#v", updateSessionInvoiceParams)
			}

			// Metrics: Increment number of expired session invoices
			metricSessionInvoicesExpiredTotal.Inc()
		}
	}
}
