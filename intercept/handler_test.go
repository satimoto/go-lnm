package intercept_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	dbMocks "github.com/satimoto/go-datastore-mocks/db"
	interceptMocks "github.com/satimoto/go-lsp/intercept/mocks"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	aliceTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNKakNDQWN5Z0F3SUJBZ0lRRFZlZWdZQmJxTnBUVTJRMm5xRWJUREFLQmdncWhrak9QUVFEQWpBeE1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmhiR2xqWlRBZQpGdzB5TWpBeE1UQXhNREF4TXpaYUZ3MHlNekF6TURjeE1EQXhNelphTURFeEh6QWRCZ05WQkFvVEZteHVaQ0JoCmRYUnZaMlZ1WlhKaGRHVmtJR05sY25ReERqQU1CZ05WQkFNVEJXRnNhV05sTUZrd0V3WUhLb1pJemowQ0FRWUkKS29aSXpqMERBUWNEUWdBRWhYU0dHOXFsNEg1LzRSWjJBYjhWTHB1RkZadDhHQ0xoWEd2SHpjbldHOGJJMzVNaAp3ZFFOWkZ2ZCt5MEJCbnVPS1pjaTRWNDFjUThXV1pKRnZSMEFLS09CeFRDQndqQU9CZ05WSFE4QkFmOEVCQU1DCkFxUXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUhBd0V3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEUKRmdRVThqU1lyYkpEK0dMK1BRbTlUYzdqejVwSTg1NHdhd1lEVlIwUkJHUXdZb0lGWVd4cFkyV0NDV3h2WTJGcwphRzl6ZElJRllXeHBZMldDRG5CdmJHRnlMVzR4TFdGc2FXTmxnZ1IxYm1sNGdncDFibWw0Y0dGamEyVjBnZ2RpCmRXWmpiMjV1aHdSL0FBQUJoeEFBQUFBQUFBQUFBQUFBQUFBQUFBQUJod1NzRXdBQ01Bb0dDQ3FHU000OUJBTUMKQTBnQU1FVUNJUUQwSlpmNXloTDVUR3kwZ2hkVUFkRHlFck1BS3l6VGlVWWl1eUN3MEUvcUxRSWdmRm9mdUMrMwpCeXcwSDRiY2FpWkZiS2h3ekI1S0tyamw2WDlYbHNiTW1kZz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	aliceMacaroonBase64 = "AgEDbG5kAvgBAwoQmVHgCAfdlNFOq0xl+N4WURIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYg1Mv+ek8sURmlwfu2FdwCyyHpnjXmdKOPay9ZR56WoA0="
	aliceHost           = "127.0.0.1:10001"
	alicePubkey         = "0365f44b5052188a9c9c46ecfbfcca1ea4176c864757214d50edcb2d5f0e8ac387"
	bobTlsCertBase64    = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNIRENDQWNLZ0F3SUJBZ0lRT3orTnU3bHJTYnBwYk5tMEJja3E3ekFLQmdncWhrak9QUVFEQWpBdk1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1Rd3dDZ1lEVlFRREV3TmliMkl3SGhjTgpNakl3TVRFd01UQXdOakl6V2hjTk1qTXdNekEzTVRBd05qSXpXakF2TVI4d0hRWURWUVFLRXhac2JtUWdZWFYwCmIyZGxibVZ5WVhSbFpDQmpaWEowTVF3d0NnWURWUVFERXdOaWIySXdXVEFUQmdjcWhrak9QUUlCQmdncWhrak8KUFFNQkJ3TkNBQVErSGFYeVlpbXJ0OXJkUmxDM0QxWlBHSzg2U0ZNa2gwNFFJRmo4bmV4alpuY05SNk1VNnNmMAorU1N3NmpSSGI2bHYycjlFM2ROajQwWTJ4aWdLcUlwUG80Ry9NSUc4TUE0R0ExVWREd0VCL3dRRUF3SUNwREFUCkJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREFUQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCUlEKZkZFNVRYdmpBaFA5L2xneEZSalRBZTJkblRCbEJnTlZIUkVFWGpCY2dnTmliMktDQ1d4dlkyRnNhRzl6ZElJRApZbTlpZ2d4d2IyeGhjaTF1TVMxaWIyS0NCSFZ1YVhpQ0NuVnVhWGh3WVdOclpYU0NCMkoxWm1OdmJtNkhCSDhBCkFBR0hFQUFBQUFBQUFBQUFBQUFBQUFBQUFBR0hCS3dUQUFNd0NnWUlLb1pJemowRUF3SURTQUF3UlFJaEFQckEKL3Uxc2pDaWVpaXNlZUhTRlN6TGU1NStJY3dyWithenRwTm5MenNyd0FpQkRHNUNld21tcWJDcmJEN1N0TWZqTwpXMUtiNXRibnIxbG0xODJ5dUJLRG5BPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	bobMacaroonBase64   = "AgEDbG5kAvgBAwoQnzHVP+ovkUHeVc8Emjo/ShIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgseB8u/LJOFg+cTnA6f47/ggxwtHIkrrrsly3+F0/xfk="
	bobHost             = "127.0.0.1:10002"
	carolTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNKekNDQWN5Z0F3SUJBZ0lRWFpYd0VwR09LSVhrWHBodWRUeVJGakFLQmdncWhrak9QUVFEQWpBeE1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmpZWEp2YkRBZQpGdzB5TWpBeE1URXhOVFE0TlROYUZ3MHlNekF6TURneE5UUTROVE5hTURFeEh6QWRCZ05WQkFvVEZteHVaQ0JoCmRYUnZaMlZ1WlhKaGRHVmtJR05sY25ReERqQU1CZ05WQkFNVEJXTmhjbTlzTUZrd0V3WUhLb1pJemowQ0FRWUkKS29aSXpqMERBUWNEUWdBRWdBeXpGaHdqSUpwTXJDaTBXdDFzaEVEeS94UENVeW5NR2NxQW1LNzVaOHpWWTZqZApjRUlJQzVIK2x2WUJGN1d3UTVKMUtZRjZEQ2pOT2o1SzJ6VlBGYU9CeFRDQndqQU9CZ05WSFE4QkFmOEVCQU1DCkFxUXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUhBd0V3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEUKRmdRVXNUaFp5ejBrQXF2UUU5cDlqUTV6ZFRwdko5OHdhd1lEVlIwUkJHUXdZb0lGWTJGeWIyeUNDV3h2WTJGcwphRzl6ZElJRlkyRnliMnlDRG5CdmJHRnlMVzR4TFdOaGNtOXNnZ1IxYm1sNGdncDFibWw0Y0dGamEyVjBnZ2RpCmRXWmpiMjV1aHdSL0FBQUJoeEFBQUFBQUFBQUFBQUFBQUFBQUFBQUJod1NzRXdBRk1Bb0dDQ3FHU000OUJBTUMKQTBrQU1FWUNJUURXbnpoZ1NnRmtzUzQ3RU1jL1ZpUEpvN2ZiN0Q4d1ZCdDBIMm04cXBXS0RnSWhBUGVpMk5mOApZc3dqQ1dtSm5UR2oyR3E5YThjK2gyTGFSTDU2WnNpMlRWVnIKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	carolMacaroonBase64 = "AgEDbG5kAvgBAwoQRiE6kNc50RM+yLZRMNGyrxIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgIVkK5Gx+Z21PM//7pFWpFJDh1hZebIhNegmG4uzhNPU="
	carolHost           = "127.0.0.1:10003"
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
	aliceMacaroon, _ := base64.StdEncoding.DecodeString(aliceMacaroonBase64)
	aliceCredentials, _ := util.NewCredential(string(aliceTlsCert))
	aliceClientConn, _ := grpc.Dial(aliceHost, grpc.WithTransportCredentials(aliceCredentials))
	defer aliceClientConn.Close()

	bobTlsCert, _ := base64.StdEncoding.DecodeString(bobTlsCertBase64)
	bobMacaroon, _ := base64.StdEncoding.DecodeString(bobMacaroonBase64)
	bobCredentials, _ := util.NewCredential(string(bobTlsCert))
	bobClientConn, _ := grpc.Dial(bobHost, grpc.WithTransportCredentials(bobCredentials))
	defer bobClientConn.Close()

	carolTlsCert, _ := base64.StdEncoding.DecodeString(carolTlsCertBase64)
	carolMacaroon, _ := base64.StdEncoding.DecodeString(carolMacaroonBase64)
	carolCredentials, _ := util.NewCredential(string(carolTlsCert))
	carolClientConn, _ := grpc.Dial(carolHost, grpc.WithTransportCredentials(carolCredentials))
	defer carolClientConn.Close()

	t.Run("Success request", func(t *testing.T) {
		mockRepository := dbMocks.NewMockRepositoryService()
		interceptor := interceptMocks.NewInterceptor(mockRepository, aliceClientConn)
		aliceCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(aliceMacaroon))
		htlcInterceptorClient, err := interceptor.GetRouterClient().HtlcInterceptor(aliceCtx)

		go interceptor.InterceptHtlc(htlcInterceptorClient)

		// Carol creates an incoming invoice
		preimage, err := util.RandomPreimage()
		paymentHash := preimage.Hash()
		t.Logf("preImage: %v", preimage.String())
		t.Logf("preImage: %v", paymentHash.String())
		
		carolLightningClient := lnrpc.NewLightningClient(carolClientConn)
		carolCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(carolMacaroon))

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

		invoiceResponse, err := carolLightningClient.AddInvoice(carolCtx, &lnrpc.Invoice{
			RHash:      paymentHash[:],
			Value:      1000,
			Expiry:     600,
			RouteHints: routingHints,
		})
		util.PanicOnError("AddInvoice", err)

		t.Logf("PaymentAddr: %v", hex.EncodeToString(invoiceResponse.PaymentAddr))
		t.Logf("PaymentRequest: %v", invoiceResponse.PaymentRequest)
		t.Logf("RHash: %v", hex.EncodeToString(invoiceResponse.RHash))

		decodePayResponse, err := carolLightningClient.DecodePayReq(carolCtx, &lnrpc.PayReqString{
			PayReq: invoiceResponse.PaymentRequest,
		})
		util.PanicOnError("DecodePayReq", err)

		t.Logf("PaymentHash: %v", decodePayResponse.PaymentHash)
		t.Logf("NumSatoshis: %v", decodePayResponse.NumSatoshis)
		t.Logf("RouteHints: %v", decodePayResponse.RouteHints)

		// Bob pays the invoice
		//bobLightningClient := lnrpc.NewLightningClient(bobClientConn)
		bobRouterClient := routerrpc.NewRouterClient(bobClientConn)
		bobCtx := metadata.AppendToOutgoingContext(context.Background(), "macaroon", hex.EncodeToString(bobMacaroon))

		sendPaymentClient, err := bobRouterClient.SendPaymentV2(bobCtx, &routerrpc.SendPaymentRequest{
			//FeeLimitMsat: 2,
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
	})
}
