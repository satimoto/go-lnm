package invoice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/lsprpc"
)

func (r *RpcInvoiceResolver) UpdateInvoiceRequest(reqCtx context.Context, input *lsprpc.UpdateInvoiceRequestRequest) (*lsprpc.UpdateInvoiceRequestResponse, error) {
	if input != nil {
		ctx := context.Background()
		invoiceRequest, err := r.InvoiceRequestRepository.GetInvoiceRequest(ctx, input.Id)

		if err != nil {
			metrics.RecordError("LNM118", "Error retrieving invoice request", err)
			log.Printf("LNM118: Input=%#v", input)
			return nil, errors.New("error retrieving invoice request")
		}

		if invoiceRequest.UserID != input.UserId {
			metrics.RecordError("LNM119", "Error invalid user for invoice request", err)
			log.Printf("LNM119: Input=%#v", input)
			return nil, errors.New("error invalid user for invoice request")
		}

		if invoiceRequest.IsSettled || invoiceRequest.PaymentRequest.Valid {
			metrics.RecordError("LNM120", "Error invoice request in progress or settled", err)
			log.Printf("LNM120: Input=%#v", input)

			return &lsprpc.UpdateInvoiceRequestResponse{
				Id:             invoiceRequest.ID,
				UserId:         invoiceRequest.UserID,
				PaymentRequest: invoiceRequest.PaymentRequest.String,
				IsSettled:      invoiceRequest.IsSettled,
			}, nil
		}

		updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)
		updateInvoiceRequestParams.PaymentRequest = dbUtil.SqlNullString(input.PaymentRequest)

		invoiceRequest, err = r.InvoiceRequestRepository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

		if err != nil {
			metrics.RecordError("LNM121", "Error updating invoice request", err)
			log.Printf("LNM121: Params=%#v", updateInvoiceRequestParams)
			return nil, errors.New("error updating invoice request")
		}

		payReq, err := r.LightningService.DecodePayReq(&lnrpc.PayReqString{
			PayReq: input.PaymentRequest,
		})

		if err != nil {
			metrics.RecordError("LNM122", "Error decoding payment request", err)
			log.Printf("LNM122: Input=%#v", input)
			return nil, errors.New("error decoding payment request")
		}

		// TODO go-api#12: Allow invoice request to be split
		if payReq.NumMsat != invoiceRequest.TotalMsat {
			metrics.RecordError("LNM123", "Error payment request amount mismatch", err)
			log.Printf("LNM123: Input=%#v", input)
			log.Printf("LNM123: PayReq=%#v", payReq)
			return nil, errors.New("error payment request amount mismatch")
		}

		go r.waitForPayment(invoiceRequest)

		return &lsprpc.UpdateInvoiceRequestResponse{}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcInvoiceResolver) UpdateSessionInvoice(reqCtx context.Context, input *lsprpc.UpdateSessionInvoiceRequest) (*lsprpc.UpdateSessionInvoiceResponse, error) {
	if input != nil {
		ctx := context.Background()
		sessionInvoice, err := r.SessionRepository.GetSessionInvoice(ctx, input.Id)

		if err != nil {
			metrics.RecordError("LNM145", "Error retrieving session invoice", err)
			log.Printf("LNM145: Input=%#v", input)
			return nil, errors.New("error retrieving session invoice")
		}

		if sessionInvoice.UserID != input.UserId {
			metrics.RecordError("LNM146", "Error invalid user for session invoice", err)
			log.Printf("LNM146: Input=%#v", input)
			return nil, errors.New("error invalid user for session invoice")
		}

		if sessionInvoice.IsSettled || !sessionInvoice.IsExpired {
			metrics.RecordError("LNM147", "Error session invoice in progress or settled", err)
			log.Printf("LNM147: Input=%#v", input)

			return &lsprpc.UpdateSessionInvoiceResponse{
				Id:             sessionInvoice.ID,
				UserId:         sessionInvoice.UserID,
				PaymentRequest: sessionInvoice.PaymentRequest,
				IsSettled:      sessionInvoice.IsSettled,
				IsExpired:      sessionInvoice.IsExpired,
			}, nil
		}

		session, err := r.SessionRepository.GetSession(ctx, sessionInvoice.SessionID)

		if err != nil {
			metrics.RecordError("LNM148", "Error retrieving session", err)
			log.Printf("LNM148: SessionID=%v", session.ID)
			return nil, errors.New("error retrieving session")
		}

		preimage, err := lightningnetwork.RandomPreimage()

		if err != nil {
			metrics.RecordError("LNM150", "Error creating preimage", err)
			log.Printf("LNM150: SessionUid=%v", session.Uid)
			return nil, errors.New("error creating preimage")
		}

		invoice, err := r.LightningService.AddInvoice(&lnrpc.Invoice{
			Memo:      session.Uid,
			Expiry:    3600,
			RPreimage: preimage[:],
			ValueMsat: sessionInvoice.TotalMsat,
		})

		if err != nil {
			metrics.RecordError("LNM151", "Error creating lightning invoice", err)
			log.Printf("LNM151: Preimage=%v, ValueMsat=%v", preimage.String(), sessionInvoice.TotalMsat)
			return nil, errors.New("error creating lightning invoice")
		}

		signMessage, err := r.LightningService.SignMessage(&lnrpc.SignMessageRequest{
			Msg: []byte(invoice.PaymentRequest),
		})

		if err != nil {
			metrics.RecordError("LNM149", "Error signing payment request", err)
			log.Printf("LNM149: PaymentRequest=%v,", invoice.PaymentRequest)
			return nil, errors.New("error signing payment request")
		}

		updateSessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
		updateSessionInvoiceParams.PaymentRequest = invoice.PaymentRequest
		updateSessionInvoiceParams.Signature = signMessage.Signature
		updateSessionInvoiceParams.IsExpired = false

		sessionInvoice, err = r.SessionRepository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

		if err != nil {
			metrics.RecordError("LNM152", "Error updating session invoice", err)
			log.Printf("LNM152: Params=%#v", updateSessionInvoiceParams)
			return nil, errors.New("error updating session invoice")
		}

		go r.waitForInvoiceExpiry(invoice.PaymentRequest)

		return &lsprpc.UpdateSessionInvoiceResponse{
			Id:             sessionInvoice.ID,
			UserId:         sessionInvoice.UserID,
			PaymentRequest: sessionInvoice.PaymentRequest,
			IsSettled:      sessionInvoice.IsSettled,
			IsExpired:      sessionInvoice.IsExpired,
		}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcInvoiceResolver) waitForInvoiceExpiry(paymentRequest string) {
	payReqParams := &lnrpc.PayReqString{PayReq: paymentRequest}
	expiry := int64(3600)

	if payReqResponse, err := r.LightningService.DecodePayReq(payReqParams); err == nil {
		expiry = payReqResponse.Expiry
	}

	ctx := context.Background()
	timeout := (time.Second * time.Duration(expiry)) + time.Minute

	time.Sleep(timeout)

	if sessionInvoice, err := r.SessionRepository.GetSessionInvoiceByPaymentRequest(ctx, paymentRequest); err == nil {
		if !sessionInvoice.IsSettled && !sessionInvoice.IsExpired {
			updateSessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
			updateSessionInvoiceParams.IsExpired = true

			_, err = r.SessionRepository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

			if err != nil {
				metrics.RecordError("LNM161", "Error updating session invoice", err)
				log.Printf("LNM161: Params=%#v", updateSessionInvoiceParams)
			}
		}
	}
}

func (r *RpcInvoiceResolver) waitForPayment(invoiceRequest db.InvoiceRequest) {
	client, err := r.LightningService.SendPaymentV2(&routerrpc.SendPaymentRequest{
		PaymentRequest: invoiceRequest.PaymentRequest.String,
		TimeoutSeconds: 120,
	})

	if err != nil {
		metrics.RecordError("LNM124", "Error sending payment", err)
		log.Printf("LNM124: InvoiceRequest=%#v", invoiceRequest)
		return
	}

	ctx := context.Background()
	updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)

waitLoop:
	for {
		payment, err := client.Recv()

		if err != nil {
			metrics.RecordError("LNM125", "Error waiting for payment", err)
			log.Printf("LNM125: PaymentRequest=%v", invoiceRequest.PaymentRequest)
			updateInvoiceRequestParams.PaymentRequest = dbUtil.SqlNullString(nil)
			break
		}

		switch payment.Status {
		case lnrpc.Payment_FAILED:
			updateInvoiceRequestParams.PaymentRequest = dbUtil.SqlNullString(nil)
			break waitLoop
		case lnrpc.Payment_SUCCEEDED:
			updateInvoiceRequestParams.IsSettled = true
			break waitLoop
		}
	}

	_, err = r.InvoiceRequestRepository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

	if err != nil {
		metrics.RecordError("LNM126", "Error updating invoice request", err)
		log.Printf("LNM126: Params=%#v", updateInvoiceRequestParams)
	}
}
