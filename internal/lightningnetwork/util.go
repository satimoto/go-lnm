package lightningnetwork

import (
	"crypto/rand"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	metrics "github.com/satimoto/go-lsp/internal/metric"
)


func CreateLightningInvoice(lightningService LightningNetwork, memo string, valueMsat int64) (string, string, error) {
	preimage, err := RandomPreimage()

	if err != nil {
		metrics.RecordError("LSP030", "Error creating invoice preimage", err)
		log.Printf("LSP030: SessionUid=%v", memo)
		return "", "", nil
	}

	invoice, err := lightningService.AddInvoice(&lnrpc.Invoice{
		Memo:      memo,
		Expiry:    3600,
		RPreimage: preimage[:],
		ValueMsat: valueMsat,
	})

	if err != nil {
		metrics.RecordError("LSP031", "Error creating lightning invoice", err)
		log.Printf("LSP031: Preimage=%v, ValueMsat=%v", preimage.String(), valueMsat)
		return "", "", nil
	}

	signMessage, err := lightningService.SignMessage(&lnrpc.SignMessageRequest{
		Msg: []byte(invoice.PaymentRequest),
	})

	if err != nil {
		metrics.RecordError("LSP167", "Error signing payment request", err)
		log.Printf("LSP167: PaymentRequest=%v,", invoice.PaymentRequest)
		return "", "", nil
	}

	return invoice.PaymentRequest, signMessage.Signature, nil
}

func RandomPreimage() (*lntypes.Preimage, error) {
	paymentPreimage := &lntypes.Preimage{}

	if _, err := rand.Read(paymentPreimage[:]); err != nil {
		return &lntypes.Preimage{}, err
	}

	return paymentPreimage, nil
}
