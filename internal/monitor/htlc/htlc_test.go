package htlc_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/satimoto/go-datastore/pkg/db"
	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"

	//interceptMocks "github.com/satimoto/go-lsp/intercept/mocks"
	"github.com/satimoto/go-lsp/internal/lightningnetwork"
	"github.com/satimoto/go-lsp/internal/messages"
	"github.com/satimoto/go-lsp/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	aliceTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNKekNDQWMyZ0F3SUJBZ0lSQUlrYVBQUlFRWitQcWNnNU9OY1BVQ2N3Q2dZSUtvWkl6ajBFQXdJd01URWYKTUIwR0ExVUVDaE1XYkc1a0lHRjFkRzluWlc1bGNtRjBaV1FnWTJWeWRERU9NQXdHQTFVRUF4TUZZV3hwWTJVdwpIaGNOTWpJd05ESTRNVGswT0RRMldoY05Nak13TmpJek1UazBPRFEyV2pBeE1SOHdIUVlEVlFRS0V4WnNibVFnCllYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmhiR2xqWlRCWk1CTUdCeXFHU000OUFnRUcKQ0NxR1NNNDlBd0VIQTBJQUJKY2NlUDFRYUtlLy9VamdIaDNDYnIzSjJHS1ZhcFBBZnBHbHpWWDdNcW5hWTlXMwp0ZDIzR2t5YWREVEJBRGhvY2FaUS9JSlJOcFVLOUJaSlQ0TFpZQStqZ2NVd2djSXdEZ1lEVlIwUEFRSC9CQVFECkFnS2tNQk1HQTFVZEpRUU1NQW9HQ0NzR0FRVUZCd01CTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME8KQkJZRUZBbVZXUlBCN2piQXUzK1RRbndCVE5FanZjd1RNR3NHQTFVZEVRUmtNR0tDQldGc2FXTmxnZ2xzYjJOaApiR2h2YzNTQ0JXRnNhV05sZ2c1d2IyeGhjaTF1TWkxaGJHbGpaWUlFZFc1cGVJSUtkVzVwZUhCaFkydGxkSUlIClluVm1ZMjl1Ym9jRWZ3QUFBWWNRQUFBQUFBQUFBQUFBQUFBQUFBQUFBWWNFckJNQUFqQUtCZ2dxaGtqT1BRUUQKQWdOSUFEQkZBaUVBOVlyWnNJK2w2WXhsUGFqMkVqcGRuZ0JZUDYzZEJHSGhLRjUrWUhDSE03UUNJRVpFT1dlYwpyNk9uMEVzVVhrUFgvQ0I1U1ZFY1RlL1NOOEVSQXlxbHFBU1IKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	aliceMacaroonBase64 = "AgEDbG5kAvgBAwoQlRk+/FPdZVhYK6PpaSf1hhIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgeTsjIGpEi06oDX1DyXkeQM+sta/rjaj2+cyZsu0ZUG8="
	aliceHost           = "127.0.0.1:10001"
	alicePubkey         = "03dc8f6923853e32358bd879543517267e6e0c08c8f37ac064a19564fd6d58678e"
	bobTlsCertBase64    = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNIakNDQWNPZ0F3SUJBZ0lSQUpGTWh4a3RlRDEvQUp4Y1hTK3Zadm93Q2dZSUtvWkl6ajBFQXdJd0x6RWYKTUIwR0ExVUVDaE1XYkc1a0lHRjFkRzluWlc1bGNtRjBaV1FnWTJWeWRERU1NQW9HQTFVRUF4TURZbTlpTUI0WApEVEl5TURReU9ESXdNRE14TmxvWERUSXpNRFl5TXpJd01ETXhObG93THpFZk1CMEdBMVVFQ2hNV2JHNWtJR0YxCmRHOW5aVzVsY21GMFpXUWdZMlZ5ZERFTU1Bb0dBMVVFQXhNRFltOWlNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkkKemowREFRY0RRZ0FFdDVnU0hZUExpZjFvNk50SCsxbHdqWjYwcC92a3VoQUxxWUsvWG5OUGdGUHhZbnV3VjRRWAptU0R0NkVOZzYxbzhJSlFKaDZINXJ0eVA3WGVmazA5T1RhT0J2ekNCdkRBT0JnTlZIUThCQWY4RUJBTUNBcVF3CkV3WURWUjBsQkF3d0NnWUlLd1lCQlFVSEF3RXdEd1lEVlIwVEFRSC9CQVV3QXdFQi96QWRCZ05WSFE0RUZnUVUKOGxqUFpZcVJqUmtweDJMb2ZaaXpWVGlWTVE4d1pRWURWUjBSQkY0d1hJSURZbTlpZ2dsc2IyTmhiR2h2YzNTQwpBMkp2WW9JTWNHOXNZWEl0YmpJdFltOWlnZ1IxYm1sNGdncDFibWw0Y0dGamEyVjBnZ2RpZFdaamIyNXVod1IvCkFBQUJoeEFBQUFBQUFBQUFBQUFBQUFBQUFBQUJod1NzRXdBRE1Bb0dDQ3FHU000OUJBTUNBMGtBTUVZQ0lRQ2cKZVozczVtU2w3L2FtNVBRSHFQWGowWFFBSm5lMXlIajJpSkI5dkRoYVpRSWhBTnREMjlHSG12UklzZVJrVlpLUwo4SUViSnQrR2VqMHYxempidkZ1ZkxPNzIKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	bobMacaroonBase64   = "AgEDbG5kAvgBAwoQUm+DqUUC5TIIxvjdc0NCGhIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgc0m7+oL6jl3G685XmxMQILOLlTwFyf83hA+0WjlOQQ8="
	bobHost             = "127.0.0.1:10002"
	carolTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNKakNDQWN5Z0F3SUJBZ0lRQ1ZkNUlQRkRYaHdQMGhhMERsOTc3REFLQmdncWhrak9QUVFEQWpBeE1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmpZWEp2YkRBZQpGdzB5TWpBME1qZ3lNREF6TXpoYUZ3MHlNekEyTWpNeU1EQXpNemhhTURFeEh6QWRCZ05WQkFvVEZteHVaQ0JoCmRYUnZaMlZ1WlhKaGRHVmtJR05sY25ReERqQU1CZ05WQkFNVEJXTmhjbTlzTUZrd0V3WUhLb1pJemowQ0FRWUkKS29aSXpqMERBUWNEUWdBRXlBaGMrbWlqaTZUZFlBM2dQKzJQeVNkQ0RxdStkZFFKeFdZc2FERXlvaGRDTkxKZwpadUpPa1VqbElyalJWQnA4Qm8rUDZobGdPZzBlYkIrV1E3YS95Nk9CeFRDQndqQU9CZ05WSFE4QkFmOEVCQU1DCkFxUXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUhBd0V3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEUKRmdRVVZSWjhTb2pXdXdDcndoQUk1d1p1ZFFYV2lia3dhd1lEVlIwUkJHUXdZb0lGWTJGeWIyeUNDV3h2WTJGcwphRzl6ZElJRlkyRnliMnlDRG5CdmJHRnlMVzR5TFdOaGNtOXNnZ1IxYm1sNGdncDFibWw0Y0dGamEyVjBnZ2RpCmRXWmpiMjV1aHdSL0FBQUJoeEFBQUFBQUFBQUFBQUFBQUFBQUFBQUJod1NzRXdBRU1Bb0dDQ3FHU000OUJBTUMKQTBnQU1FVUNJUURsaFlQbzZnSjYwSnhhYXJCaCtNamVrKzBtbkZtOWM1UWR3eFRuOUJJa0ZRSWdQY1JiREU2TgpFUGlHa3ZXWU1KV2pIc3k0eGdJYjhnZUczeS95a2R1bTl5WT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	carolMacaroonBase64 = "AgEDbG5kAvgBAwoQUUKDxk7jK5FutZIMMeqnTRIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgwQb10IkH5R4pH0T9LjIR2wbQ+QRqtQ00Rjo6FURBCvM="
	carolHost           = "127.0.0.1:10003"
	carolPubkey         = "03f463ae3f8ae6e1b29a9e44b81d09e0a8ad30d17c6aa653af4d898cc008b67a59"
)

func getRoutingHints(t *testing.T, lnclient lnrpc.LightningClient, ctx context.Context) []*lnrpc.RouteHint {
	var hints []*lnrpc.RouteHint

	listChannelsResponse, err := lnclient.ListChannels(ctx, &lnrpc.ListChannelsRequest{
		PrivateOnly: true,
	})

	if err != nil {
		t.Errorf("ListChannels error: %v", err)
		return hints
	}

	for _, channel := range listChannelsResponse.Channels {
		chanInfoResponse, err := lnclient.GetChanInfo(ctx, &lnrpc.ChanInfoRequest{
			ChanId: channel.ChanId,
		})

		if err != nil {
			continue
		}

		remotePolicy := chanInfoResponse.Node1Policy
		if chanInfoResponse.Node2Policy != nil && chanInfoResponse.Node2Pub == channel.RemotePubkey {
			remotePolicy = chanInfoResponse.Node2Policy
		}

		feeBaseMsat := uint32(1)
		proportionalFee := uint32(1)
		cltvExpiryDelta := uint32(40)

		if remotePolicy != nil {
			feeBaseMsat = uint32(remotePolicy.FeeBaseMsat)
			proportionalFee = uint32(remotePolicy.FeeRateMilliMsat)
			cltvExpiryDelta = remotePolicy.TimeLockDelta
		}

		t.Log("RoutHint")
		t.Logf("RemotePubkey: %v", channel.RemotePubkey)
		t.Logf("ChanId: %v", channel.ChanId)
		t.Logf("feeBaseMsat: %v", feeBaseMsat)
		t.Logf("proportionalFee: %v", proportionalFee)
		t.Logf("cltvExpiryDelta: %v", cltvExpiryDelta)

		hints = append(hints, &lnrpc.RouteHint{
			HopHints: []*lnrpc.HopHint{
				{
					NodeId:                    channel.RemotePubkey,
					ChanId:                    channel.ChanId,
					FeeBaseMsat:               feeBaseMsat,
					FeeProportionalMillionths: proportionalFee,
					CltvExpiryDelta:           cltvExpiryDelta,
				},
			},
		})
	}

	return hints
}

func TestInterceptor(t *testing.T) {
	os.Setenv("LND_MACAROON", aliceMacaroonBase64)
	defer os.Unsetenv("LND_MACAROON")

	aliceTlsCert, _ := base64.StdEncoding.DecodeString(aliceTlsCertBase64)
	//aliceMacaroon, _ := base64.StdEncoding.DecodeString(aliceMacaroonBase64)
	aliceCredentials, _ := util.NewCredential(string(aliceTlsCert))
	aliceClientConn, _ := grpc.Dial(aliceHost, grpc.WithTransportCredentials(aliceCredentials))
	defer aliceClientConn.Close()

	bobTlsCert, _ := base64.StdEncoding.DecodeString(bobTlsCertBase64)
	bobMacaroon, _ := base64.StdEncoding.DecodeString(bobMacaroonBase64)
	bobCredentials, _ := util.NewCredential(string(bobTlsCert))
	bobClientConn, _ := grpc.Dial(bobHost, grpc.WithTransportCredentials(bobCredentials))
	defer bobClientConn.Close()

	carolPubkey := carolPubkey
	carolTlsCert, _ := base64.StdEncoding.DecodeString(carolTlsCertBase64)
	carolMacaroon, _ := base64.StdEncoding.DecodeString(carolMacaroonBase64)
	carolCredentials, _ := util.NewCredential(string(carolTlsCert))
	carolClientConn, _ := grpc.Dial(carolHost, grpc.WithTransportCredentials(carolCredentials))
	defer carolClientConn.Close()

	t.Run("Success request", func(t *testing.T) {
		mockRepository := dbMocks.NewMockRepositoryService()
		//interceptor := interceptMocks.NewInterceptor(mockRepository, aliceClientConn)
		//aliceCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(aliceMacaroon))
		//htlcInterceptorClient, err := interceptor.GetRouterClient().HtlcInterceptor(aliceCtx)

		//go interceptor.InterceptHtlc(htlcInterceptorClient)

		// Carol creates an incoming invoice
		preimage, err := lightningnetwork.RandomPreimage()
		paymentHash := preimage.Hash()
		t.Logf("preImage: %v", preimage.String())
		t.Logf("paymentHash: %v", paymentHash.String())

		carolLightningClient := lnrpc.NewLightningClient(carolClientConn)
		carolCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(carolMacaroon))

		fakeChanID := &lnwire.ShortChannelID{
			BlockHeight: uint32(rand.Intn(math.MaxUint32)),
			TxIndex:     uint32(rand.Intn(math.MaxUint32)),
			TxPosition:  uint16(rand.Intn(math.MaxUint16)),
		}
		t.Logf("fakeChanID: %v", fakeChanID.ToUint64())

		routingHints := []*lnrpc.RouteHint{
			{
				HopHints: []*lnrpc.HopHint{
					{
						NodeId:                    alicePubkey,
						ChanId:                    fakeChanID.ToUint64(),
						FeeBaseMsat:               0,
						FeeProportionalMillionths: 0,
						CltvExpiryDelta:           40,
					},
				},
			}}

		invoiceResponse, err := carolLightningClient.AddInvoice(carolCtx, &lnrpc.Invoice{
			RPreimage:  preimage[:],
			Value:      1000,
			Expiry:     600,
			RouteHints: routingHints,
		})
		dbUtil.PanicOnError("TEST", "AddInvoice", err)

		t.Logf("PaymentAddr: %v", hex.EncodeToString(invoiceResponse.PaymentAddr))
		t.Logf("PaymentRequest: %v", invoiceResponse.PaymentRequest)
		t.Logf("RHash: %v", hex.EncodeToString(invoiceResponse.RHash))

		decodePayResponse, err := carolLightningClient.DecodePayReq(carolCtx, &lnrpc.PayReqString{
			PayReq: invoiceResponse.PaymentRequest,
		})
		dbUtil.PanicOnError("TEST", "DecodePayReq", err)

		t.Logf("PaymentHash: %v", decodePayResponse.PaymentHash)
		t.Logf("NumSatoshis: %v", decodePayResponse.NumSatoshis)
		t.Logf("RouteHints: %v", decodePayResponse.RouteHints)

		// Add Channel Request
		mockRepository.SetGetChannelRequestByPaymentHashMockData(dbMocks.ChannelRequestMockData{
			ChannelRequest: db.ChannelRequest{
				Status:      db.ChannelRequestStatusREQUESTED,
				Pubkey:      carolPubkey,
				PaymentHash: invoiceResponse.RHash,
				PaymentAddr: invoiceResponse.PaymentAddr,
				AmountMsat:  decodePayResponse.NumMsat,
				SettledMsat: 0,
			},
			Error: nil,
		})

		// Bob pays the invoice
		//bobLightningClient := lnrpc.NewLightningClient(bobClientConn)
		bobRouterClient := routerrpc.NewRouterClient(bobClientConn)
		bobCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(bobMacaroon))

		sendPaymentClient, err := bobRouterClient.SendPaymentV2(bobCtx, &routerrpc.SendPaymentRequest{
			FeeLimitMsat:   2,
			PaymentRequest: invoiceResponse.PaymentRequest,
			//RouteHints: getRoutingHints(t, bobLightningClient, bobCtx),
			TimeoutSeconds: 60,
		})

		for {
			paymentResponse, err := sendPaymentClient.Recv()

			if err != nil {
				t.Errorf("SendPaymentV2 Recv error: %v", err)
			}

			t.Logf("Status: %v", paymentResponse.Status)

			if paymentResponse.Status != lnrpc.Payment_IN_FLIGHT {
				if paymentResponse.Status == lnrpc.Payment_FAILED {
					t.Errorf("HTLCs: %v", paymentResponse.Htlcs)
					t.Errorf("FailureReason: %v", paymentResponse.FailureReason)
				}
				break
			}
		}

		/*sendPayment, err := bobLightningClient.SendPaymentSync(bobCtx, &lnrpc.SendRequest{
			PaymentRequest: invoiceResponse.PaymentRequest,
			//RouteHints: getRoutingHints(t, bobLightningClient, bobCtx),
			//TimeoutSeconds: 60,
		})

		t.Logf("PaymentError: %v", sendPayment.PaymentError)
		t.Logf("PaymentPreimage: %v", hex.EncodeToString(sendPayment.PaymentPreimage))
		t.Logf("PaymentRoute: %v", sendPayment.PaymentRoute)
		t.Logf("PaymentHash: %v", hex.EncodeToString(sendPayment.PaymentHash))*/

		t.Errorf("End: %v", invoiceResponse.PaymentRequest)
		time.Sleep(10 * time.Second)
	})
}

func TestHodl(t *testing.T) {
	os.Setenv("LND_MACAROON", aliceMacaroonBase64)
	defer os.Unsetenv("LND_MACAROON")

	aliceTlsCert, _ := base64.StdEncoding.DecodeString(aliceTlsCertBase64)
	aliceMacaroon, _ := base64.StdEncoding.DecodeString(aliceMacaroonBase64)
	aliceCredentials, _ := util.NewCredential(string(aliceTlsCert))
	aliceClientConn, _ := grpc.Dial(aliceHost, grpc.WithTransportCredentials(aliceCredentials))
	defer aliceClientConn.Close()

	bobTlsCert, _ := base64.StdEncoding.DecodeString(bobTlsCertBase64)
	bobMacaroon, _ := base64.StdEncoding.DecodeString(bobMacaroonBase64)
	bobCredentials, _ := util.NewCredential(string(bobTlsCert))
	bobClientConn, _ := grpc.Dial(bobHost, grpc.WithTransportCredentials(bobCredentials))
	defer bobClientConn.Close()

	//carolPubkey := []byte(carolPubkey)
	carolTlsCert, _ := base64.StdEncoding.DecodeString(carolTlsCertBase64)
	carolMacaroon, _ := base64.StdEncoding.DecodeString(carolMacaroonBase64)
	carolCredentials, _ := util.NewCredential(string(carolTlsCert))
	carolClientConn, _ := grpc.Dial(carolHost, grpc.WithTransportCredentials(carolCredentials))
	defer carolClientConn.Close()

	t.Run("Success request", func(t *testing.T) {
		// Client Carol needs a channel
		// 1. Create a preimage and payment hash, only Carol knows the preimage
		preimage, _ := lightningnetwork.RandomPreimage()
		paymentHash := preimage.Hash()
		log.Printf("preImage: %v", preimage.String())
		log.Printf("paymentHash: %v", paymentHash.String())

		// 2. Carol adds a HODL invoice for Alice to settle
		carolInvoicesClient := invoicesrpc.NewInvoicesClient(carolClientConn)
		carolCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(carolMacaroon))

		carolInvoiceResp, _ := carolInvoicesClient.AddHoldInvoice(carolCtx, &invoicesrpc.AddHoldInvoiceRequest{
			Hash:  paymentHash[:],
			Value: 1000,
		})
		log.Printf("Carol Payment Request: %v", carolInvoiceResp.PaymentRequest)

		carolInvoiceChannel := subscribeSingleInvoiceChannel(carolInvoicesClient, carolCtx, paymentHash[:])
		carolInvoice := <-carolInvoiceChannel
		log.Printf("Carol Invoice %v: state %v", hex.EncodeToString(carolInvoice.RHash), carolInvoice.State)

		// 3. Carol passes the invoice to Alice
		// Alice creates an invoice with the payment hash
		aliceInvoicesClient := invoicesrpc.NewInvoicesClient(aliceClientConn)
		aliceCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(aliceMacaroon))

		// This could include channel opening fees
		aliceInvoiceResp, _ := aliceInvoicesClient.AddHoldInvoice(aliceCtx, &invoicesrpc.AddHoldInvoiceRequest{
			Hash:  paymentHash[:],
			Value: 1000,
		})
		log.Printf("Alice Payment Request: %v", aliceInvoiceResp.PaymentRequest)

		aliceInvoiceChannel := subscribeSingleInvoiceChannel(aliceInvoicesClient, aliceCtx, paymentHash[:])
		aliceInvoiceUpdate1 := <-aliceInvoiceChannel
		log.Printf("Alice Invoice %v: state %v", hex.EncodeToString(aliceInvoiceUpdate1.RHash), aliceInvoiceUpdate1.State)

		// 4. Bob pays Alice's invoice
		bobRouterClient := routerrpc.NewRouterClient(bobClientConn)
		bobCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(bobMacaroon))

		sendPaymentClient, _ := bobRouterClient.SendPaymentV2(bobCtx, &routerrpc.SendPaymentRequest{
			//FeeLimitMsat:   200,
			PaymentRequest: aliceInvoiceResp.PaymentRequest,
			//RouteHints: getRoutingHints(t, bobLightningClient, bobCtx),
			TimeoutSeconds: 6000,
		})

		sendPaymentChannel := sendPaymentChannel(sendPaymentClient)
		<-sendPaymentChannel
		<-sendPaymentChannel

		time.Sleep(2 * time.Second)

		aliceInvoiceUpdate2 := <-aliceInvoiceChannel
		log.Printf("Alice Invoice %v: state %v", hex.EncodeToString(aliceInvoiceUpdate2.RHash), aliceInvoiceUpdate2.State)

		log.Printf("Settle")
		_, err := aliceInvoicesClient.SettleInvoice(aliceCtx, &invoicesrpc.SettleInvoiceMsg{
			Preimage: preimage[:],
		})

		if err != nil {
			log.Printf("Error %v", err)
		}

		log.Printf("Done")

		aliceInvoiceUpdate3 := <-aliceInvoiceChannel
		log.Printf("Alice Invoice %v: state %v", hex.EncodeToString(aliceInvoiceUpdate3.RHash), aliceInvoiceUpdate3.State)

		// Temp Cleanup
		carolInvoicesClient.CancelInvoice(carolCtx, &invoicesrpc.CancelInvoiceMsg{
			PaymentHash: paymentHash[:],
		})
		aliceInvoicesClient.CancelInvoice(aliceCtx, &invoicesrpc.CancelInvoiceMsg{
			PaymentHash: paymentHash[:],
		})

		t.Errorf("End: %v", true)
		time.Sleep(2 * time.Second)

	})
}

func TestCustomMessages(t *testing.T) {
	os.Setenv("LND_MACAROON", aliceMacaroonBase64)
	defer os.Unsetenv("LND_MACAROON")

	alicePubkeyBytes, _ := hex.DecodeString(alicePubkey)
	aliceTlsCert, _ := base64.StdEncoding.DecodeString(aliceTlsCertBase64)
	aliceMacaroon, _ := base64.StdEncoding.DecodeString(aliceMacaroonBase64)
	aliceCredentials, _ := util.NewCredential(string(aliceTlsCert))
	aliceClientConn, _ := grpc.Dial(aliceHost, grpc.WithTransportCredentials(aliceCredentials))
	defer aliceClientConn.Close()

	carolPubkeyBytes, _ := hex.DecodeString(carolPubkey)
	carolTlsCert, _ := base64.StdEncoding.DecodeString(carolTlsCertBase64)
	carolMacaroon, _ := base64.StdEncoding.DecodeString(carolMacaroonBase64)
	carolCredentials, _ := util.NewCredential(string(carolTlsCert))
	carolClientConn, _ := grpc.Dial(carolHost, grpc.WithTransportCredentials(carolCredentials))
	defer carolClientConn.Close()

	t.Run("Success request", func(t *testing.T) {
		preimage, _ := lightningnetwork.RandomPreimage()
		paymentHash := preimage.Hash()
		log.Printf("preImage: %v", preimage.String())
		log.Printf("paymentHash: %v", paymentHash.String())

		carolLightningClient := lnrpc.NewLightningClient(carolClientConn)
		carolCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(carolMacaroon))
		carolCustomMessagesChannel := subscribeCustomMessagesChannel(carolLightningClient, carolCtx)

		aliceLightningClient := lnrpc.NewLightningClient(aliceClientConn)
		aliceCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(aliceMacaroon))
		aliceCustomMessagesChannel := subscribeCustomMessagesChannel(aliceLightningClient, aliceCtx)

		aliceLightningClient.SendCustomMessage(aliceCtx, &lnrpc.SendCustomMessageRequest{
			Peer: carolPubkeyBytes,
			Type: messages.CHANNELREQUEST_RECEIVE_PREIMAGE,
			Data: []byte("Ping"),
		})

		<-carolCustomMessagesChannel

		carolLightningClient.SendCustomMessage(carolCtx, &lnrpc.SendCustomMessageRequest{
			Peer: alicePubkeyBytes,
			Type: messages.CHANNELREQUEST_SEND_CHAN_ID,
			Data: []byte("Pong"),
		})

		<-aliceCustomMessagesChannel

		t.Errorf("End: %v", true)
	})
}

func TestHodl2(t *testing.T) {
	os.Setenv("LND_MACAROON", aliceMacaroonBase64)
	defer os.Unsetenv("LND_MACAROON")

	aliceTlsCert, _ := base64.StdEncoding.DecodeString(aliceTlsCertBase64)
	aliceMacaroon, _ := base64.StdEncoding.DecodeString(aliceMacaroonBase64)
	aliceCredentials, _ := util.NewCredential(string(aliceTlsCert))
	aliceClientConn, _ := grpc.Dial(aliceHost, grpc.WithTransportCredentials(aliceCredentials))
	defer aliceClientConn.Close()

	bobTlsCert, _ := base64.StdEncoding.DecodeString(bobTlsCertBase64)
	bobMacaroon, _ := base64.StdEncoding.DecodeString(bobMacaroonBase64)
	bobCredentials, _ := util.NewCredential(string(bobTlsCert))
	bobClientConn, _ := grpc.Dial(bobHost, grpc.WithTransportCredentials(bobCredentials))
	defer bobClientConn.Close()

	carolPubkeyBytes, _ := hex.DecodeString(carolPubkey)
	carolTlsCert, _ := base64.StdEncoding.DecodeString(carolTlsCertBase64)
	carolMacaroon, _ := base64.StdEncoding.DecodeString(carolMacaroonBase64)
	carolCredentials, _ := util.NewCredential(string(carolTlsCert))
	carolClientConn, _ := grpc.Dial(carolHost, grpc.WithTransportCredentials(carolCredentials))
	defer carolClientConn.Close()

	t.Run("Success request", func(t *testing.T) {
		// Client Carol needs a channel
		aliceLightningClient := lnrpc.NewLightningClient(aliceClientConn)
		aliceRouterClient := routerrpc.NewRouterClient(aliceClientConn)
		aliceCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(aliceMacaroon))

		bobRouterClient := routerrpc.NewRouterClient(bobClientConn)
		bobCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(bobMacaroon))

		carolLightningClient := lnrpc.NewLightningClient(carolClientConn)
		carolInvoicesClient := invoicesrpc.NewInvoicesClient(carolClientConn)
		carolCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(carolMacaroon))
		carolCustomMessagesChannel := subscribeCustomMessagesChannel(carolLightningClient, carolCtx)

		// 1. Carol request that alice creates a channel to her.
		//    Alice creates a preimage and paymentHash, giving the paymentHash to Carol
		preimage, _ := lightningnetwork.RandomPreimage()
		paymentHash := preimage.Hash()
		log.Printf("preImage: %v", preimage.String())
		log.Printf("paymentHash: %v", paymentHash.String())

		aliceHtlcEventsChannel := subscribeHtlcEventsChannel(aliceRouterClient, aliceCtx)

		// 2. Carol creates a HODL invoice and will ask bob to settle it
		//    The invoice includes a routing hint to hop via alice
		fakeChanID := &lnwire.ShortChannelID{BlockHeight: 1, TxIndex: 0, TxPosition: 0}
		routingHints := []*lnrpc.RouteHint{
			{
				HopHints: []*lnrpc.HopHint{
					{
						NodeId:                    alicePubkey,
						ChanId:                    fakeChanID.ToUint64(),
						FeeBaseMsat:               0,
						FeeProportionalMillionths: 0,
						CltvExpiryDelta:           40,
					},
				},
			}}

		carolInvoiceResp, _ := carolInvoicesClient.AddHoldInvoice(carolCtx, &invoicesrpc.AddHoldInvoiceRequest{
			Hash:       paymentHash[:],
			Value:      1000,
			RouteHints: routingHints,
		})
		subscribeSingleInvoiceChannel(carolInvoicesClient, carolCtx, paymentHash[:])

		// 3. Bob pays the invoice
		bobRouterClient.SendPaymentV2(bobCtx, &routerrpc.SendPaymentRequest{
			//FeeLimitMsat:   200,
			PaymentRequest: carolInvoiceResp.PaymentRequest,
			//RouteHints: getRoutingHints(t, bobLightningClient, bobCtx),
			TimeoutSeconds: 6000,
		})

		// 4. Alice intercepts the HTLC
		//    Sends carol the preimage
		<-aliceHtlcEventsChannel

		aliceLightningClient.SendCustomMessage(aliceCtx, &lnrpc.SendCustomMessageRequest{
			Peer: carolPubkeyBytes,
			Type: uint32(50001),
			Data: []byte(preimage.String()),
		})

		// 5. Carol receives the preimage
		//    Settles invoice
		customMessage := <-carolCustomMessagesChannel
		customMessagePreimage, _ := lntypes.MakePreimageFromStr(string(customMessage.Data))

		carolInvoicesClient.SettleInvoice(carolCtx, &invoicesrpc.SettleInvoiceMsg{
			Preimage: customMessagePreimage[:],
		})

		time.Sleep(2 * time.Second)

		t.Errorf("End: %v", true)
	})
}

func sendPaymentChannel(sendPaymentClient routerrpc.Router_SendPaymentV2Client) <-chan lnrpc.Payment {
	paymentChan := make(chan lnrpc.Payment)

	go func() {
		defer close(paymentChan)

		for {
			payment, err := sendPaymentClient.Recv()

			if err != nil {
				log.Printf("Payment Error: %v", err)
				break
			}

			log.Printf("Payment %v: state %v", payment.PaymentHash, payment.Status)
			paymentChan <- *payment
		}
	}()

	return paymentChan
}

func subscribeCustomMessagesChannel(lightningClient lnrpc.LightningClient, ctx context.Context) <-chan lnrpc.CustomMessage {
	customMessagesChan := make(chan lnrpc.CustomMessage)

	go func() {
		defer close(customMessagesChan)

		subscribeCustomMessagesClient, _ := lightningClient.SubscribeCustomMessages(ctx, &lnrpc.SubscribeCustomMessagesRequest{})

		for {
			customMessage, err := subscribeCustomMessagesClient.Recv()

			if err != nil {
				log.Printf("Custom Mesage Error: %v", err)
				break
			}

			log.Printf("Custom Message %v: type %v data: %v", hex.EncodeToString(customMessage.Peer), customMessage.Type, string(customMessage.Data))
			customMessagesChan <- *customMessage
		}
	}()

	return customMessagesChan
}

func subscribeHtlcEventsChannel(routerClient routerrpc.RouterClient, ctx context.Context) <-chan routerrpc.HtlcEvent {
	htlcEventsChan := make(chan routerrpc.HtlcEvent)

	go func() {
		defer close(htlcEventsChan)

		subscribeHtlcEventsClient, _ := routerClient.SubscribeHtlcEvents(ctx, &routerrpc.SubscribeHtlcEventsRequest{})

		for {
			htlcEvent, err := subscribeHtlcEventsClient.Recv()

			if err != nil {
				log.Printf("HTLC Event Error: %v", err)
				break
			}

			log.Printf("HTLC Event %v", htlcEvent.EventType)
			htlcEventsChan <- *htlcEvent
		}
	}()

	return htlcEventsChan
}

func subscribeSingleInvoiceChannel(invoicesClient invoicesrpc.InvoicesClient, ctx context.Context, paymentHash []byte) <-chan lnrpc.Invoice {
	invoiceChan := make(chan lnrpc.Invoice)

	go func() {
		defer close(invoiceChan)

		subscribeSingleInvoiceClient, _ := invoicesClient.SubscribeSingleInvoice(ctx, &invoicesrpc.SubscribeSingleInvoiceRequest{
			RHash: paymentHash,
		})

		for {
			invoice, err := subscribeSingleInvoiceClient.Recv()

			if err != nil {
				log.Printf("Single Invoice Error: %v", err)
				break
			}

			log.Printf("Single Invoice %v: state %v", hex.EncodeToString(invoice.RHash), invoice.State)
			invoiceChan <- *invoice
		}
	}()

	return invoiceChan
}

func subscribeInvoicesChannel(lightningClient lnrpc.LightningClient, ctx context.Context) <-chan lnrpc.Invoice {
	invoiceChan := make(chan lnrpc.Invoice)

	go func() {
		defer close(invoiceChan)

		subscribeInvoicesClient, _ := lightningClient.SubscribeInvoices(ctx, &lnrpc.InvoiceSubscription{})

		for {
			invoice, err := subscribeInvoicesClient.Recv()

			if err != nil {
				log.Printf("Invoice Error: %v", err)
				break
			}

			log.Printf("Invoice %v: state %v", hex.EncodeToString(invoice.RHash), invoice.State)
			invoiceChan <- *invoice
		}
	}()

	return invoiceChan
}
