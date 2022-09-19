package invoice

import (
	"context"
	"errors"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/lsprpc"
)

func (r *RpcInvoiceResolver) UpdateInvoice(ctx context.Context, input *lsprpc.UpdateInvoiceRequest) (*lsprpc.UpdateInvoiceResponse, error) {
	if input != nil {
		invoiceRequest, err := r.InvoiceRequestResolver.Repository.GetInvoiceRequest(ctx, input.Id)

		if err != nil {
			dbUtil.LogOnError("LSP118", "Error retrieving invoice request", err)
			log.Printf("LSP118: Input=%#v", input)
			return nil, errors.New("error retrieving invoice request")
		}

		if invoiceRequest.UserID != input.UserId {
			dbUtil.LogOnError("LSP119", "Error invalid user for invoice request", err)
			log.Printf("LSP119: Input=%#v", input)
			return nil, errors.New("error invalid user for invoice request")
		}

		if invoiceRequest.IsSettled || invoiceRequest.PaymentRequest.Valid {
			dbUtil.LogOnError("LSP120", "Error invoice request in progress or settled", err)
			log.Printf("LSP120: Input=%#v", input)

			return &lsprpc.UpdateInvoiceResponse{
				Id:             invoiceRequest.ID,
				UserId:         invoiceRequest.UserID,
				PaymentRequest: invoiceRequest.PaymentRequest.String,
				IsSettled:      invoiceRequest.IsSettled,
			}, nil
		}

		updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)
		updateInvoiceRequestParams.PaymentRequest = dbUtil.SqlNullString(input.PaymentRequest)

		invoiceRequest, err = r.InvoiceRequestResolver.Repository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

		if err != nil {
			dbUtil.LogOnError("LSP121", "Error updating invoice request", err)
			log.Printf("LSP121: Params=%#v", updateInvoiceRequestParams)
			return nil, errors.New("error updating invoice request")
		}

		payReq, err := r.LightningService.DecodePayReq(&lnrpc.PayReqString{
			PayReq: input.PaymentRequest,
		})

		if err != nil {
			dbUtil.LogOnError("LSP122", "Error decoding payment request", err)
			log.Printf("LSP122: Input=%#v", input)
			return nil, errors.New("error decoding payment request")
		}

		if payReq.NumMsat != invoiceRequest.TotalMsat {
			dbUtil.LogOnError("LSP123", "Error payment request amount mismatch", err)
			log.Printf("LSP123: Input=%#v", input)
			log.Printf("LSP123: PayReq=%#v", payReq)
			return nil, errors.New("error payment request amount mismatch")
		}

		go r.waitForPayment(invoiceRequest)

		return &lsprpc.UpdateInvoiceResponse{}, nil
	}

	return nil, errors.New("missing request")
}

func (r *RpcInvoiceResolver) waitForPayment(invoiceRequest db.InvoiceRequest) {
	client, err := r.LightningService.SendPaymentV2(&routerrpc.SendPaymentRequest{
		PaymentRequest: invoiceRequest.PaymentRequest.String,
	})

	if err != nil {
		dbUtil.LogOnError("LSP124", "Error sending payment", err)
		log.Printf("LSP124: InvoiceRequest=%#v", invoiceRequest)
		return
	}

	ctx := context.Background()
	updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)

waitLoop:
	for {
		payment, err := client.Recv()

		if err != nil {
			dbUtil.LogOnError("LSP125", "Error waiting for payment", err)
			log.Printf("LSP125: PaymentRequest=%v", invoiceRequest.PaymentRequest)
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

	_, err = r.InvoiceRequestResolver.Repository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

	if err != nil {
		dbUtil.LogOnError("LSP126", "Error updating invoice request", err)
		log.Printf("LSP126: Params=%#v", updateInvoiceRequestParams)
	}
}
