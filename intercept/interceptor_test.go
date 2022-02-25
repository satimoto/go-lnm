package intercept_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	dbMocks "github.com/satimoto/go-datastore-mocks/db"
	"github.com/satimoto/go-datastore/db"

	interceptMocks "github.com/satimoto/go-lsp/intercept/mocks"
	"github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	aliceTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNLRENDQWMyZ0F3SUJBZ0lSQUtkdXMxSXJuQkNtQnhxai9tV3cyejh3Q2dZSUtvWkl6ajBFQXdJd01URWYKTUIwR0ExVUVDaE1XYkc1a0lHRjFkRzluWlc1bGNtRjBaV1FnWTJWeWRERU9NQXdHQTFVRUF4TUZZV3hwWTJVdwpIaGNOTWpJd01qRXpNVE13TVRFNFdoY05Nak13TkRFd01UTXdNVEU0V2pBeE1SOHdIUVlEVlFRS0V4WnNibVFnCllYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmhiR2xqWlRCWk1CTUdCeXFHU000OUFnRUcKQ0NxR1NNNDlBd0VIQTBJQUJBWW5mYVA3S05KeHkrSjlmak9wYkJLTEVNRTFtaVFwQUFOUmVPK2QyMWZ3NVRMawoxbTNXdHBwYWdLQVpuMFM0aS9vNmFWampKWnZKWmNXV1Y5Y1NoSWFqZ2NVd2djSXdEZ1lEVlIwUEFRSC9CQVFECkFnS2tNQk1HQTFVZEpRUU1NQW9HQ0NzR0FRVUZCd01CTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME8KQkJZRUZBZ0pyM25HdEYwT3FaUWJuOHdJRmovQkdUdHlNR3NHQTFVZEVRUmtNR0tDQldGc2FXTmxnZ2xzYjJOaApiR2h2YzNTQ0JXRnNhV05sZ2c1d2IyeGhjaTF1TWkxaGJHbGpaWUlFZFc1cGVJSUtkVzVwZUhCaFkydGxkSUlIClluVm1ZMjl1Ym9jRWZ3QUFBWWNRQUFBQUFBQUFBQUFBQUFBQUFBQUFBWWNFckJVQUJqQUtCZ2dxaGtqT1BRUUQKQWdOSkFEQkdBaUVBeklvU2EveVREVGZOMEYrTGt2cmpUZEVpakllNEV3WkRSZmNCRzRBTkRsRUNJUURoT0x5cwoyZWZKdGJ4MEFiVU9od3pFT29uT1BYZFdZWG1FY05OWXpZV0ZQQT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	aliceMacaroonBase64 = "AgEDbG5kAvgBAwoQurO86Pg/X0uSfvXx966lpRIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgY++MCgpVDR87iyTDZ5ep6HVQ3DaW7xsNaALXJ4Da4ZQ="
	aliceHost           = "127.0.0.1:10001"
	alicePubkey         = "03656ab6f8017c86f020d8bd6ef5294cc17177a94495d70efa2af43628889b983a"
	bobTlsCertBase64    = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNHekNDQWNLZ0F3SUJBZ0lRZkV1elYvNnc0WEM5WEJ0V0NkUzd1VEFLQmdncWhrak9QUVFEQWpBdk1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1Rd3dDZ1lEVlFRREV3TmliMkl3SGhjTgpNakl3TWpFek1UTXdNVEU0V2hjTk1qTXdOREV3TVRNd01URTRXakF2TVI4d0hRWURWUVFLRXhac2JtUWdZWFYwCmIyZGxibVZ5WVhSbFpDQmpaWEowTVF3d0NnWURWUVFERXdOaWIySXdXVEFUQmdjcWhrak9QUUlCQmdncWhrak8KUFFNQkJ3TkNBQVJwZFJsdlYrYmE1NzlWRjFYV1BEejlsaTRUbEI5TDI4ak9zLzhmdXZxQml2ZTEvUytKVlA0agprc0dQWi9EVHl2a2c4VXppMUVjK0FZeUNiL3EyTVhNU280Ry9NSUc4TUE0R0ExVWREd0VCL3dRRUF3SUNwREFUCkJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREFUQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCVGEKUFk1UXNXaDhHRkFLNEhsTEYvRzVQL1lUMWpCbEJnTlZIUkVFWGpCY2dnTmliMktDQ1d4dlkyRnNhRzl6ZElJRApZbTlpZ2d4d2IyeGhjaTF1TWkxaWIyS0NCSFZ1YVhpQ0NuVnVhWGh3WVdOclpYU0NCMkoxWm1OdmJtNkhCSDhBCkFBR0hFQUFBQUFBQUFBQUFBQUFBQUFBQUFBR0hCS3dWQUFVd0NnWUlLb1pJemowRUF3SURSd0F3UkFJZ0VpM0IKQmp1eEhmYUdMQ1NlMURYL1I3WWt0dEdlSmZFT0ZyY0ZMOUZYQk5nQ0lBUVpzVU5mMmFlcFNQd1V3L3lhT0ZLZgpGNFVLcTVKZFFqQ2RaN0Y4eURzdwotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	bobMacaroonBase64   = "AgEDbG5kAvgBAwoQdT5uqIvWxNVt8kTIox647hIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgy1UUTUtATTBQh3PbowpAW9SaOOL2X1RP9s6OwQXgmKc="
	bobHost             = "127.0.0.1:10002"
	carolTlsCertBase64  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNKakNDQWN5Z0F3SUJBZ0lRY1B4bklMK0lkZ3JaL1l5eWlyQTJBekFLQmdncWhrak9QUVFEQWpBeE1SOHcKSFFZRFZRUUtFeFpzYm1RZ1lYVjBiMmRsYm1WeVlYUmxaQ0JqWlhKME1RNHdEQVlEVlFRREV3VmpZWEp2YkRBZQpGdzB5TWpBeU1UTXhNekF4TVRoYUZ3MHlNekEwTVRBeE16QXhNVGhhTURFeEh6QWRCZ05WQkFvVEZteHVaQ0JoCmRYUnZaMlZ1WlhKaGRHVmtJR05sY25ReERqQU1CZ05WQkFNVEJXTmhjbTlzTUZrd0V3WUhLb1pJemowQ0FRWUkKS29aSXpqMERBUWNEUWdBRWM3dzhZdDBxTlQvV2NRaWNteGw2QlJ6NUFYa2N4c2N5ZmNnaGJ1b2xGelVERTF3bwpiL01iSmlmbkhZTnYxR2tCTmdXL3VrMWhIdkU3OEw1Q3VtQngvNk9CeFRDQndqQU9CZ05WSFE4QkFmOEVCQU1DCkFxUXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUhBd0V3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEUKRmdRVWlnMmR1NFhhV2w2NXg4SlVEaUlxa1lza0RyRXdhd1lEVlIwUkJHUXdZb0lGWTJGeWIyeUNDV3h2WTJGcwphRzl6ZElJRlkyRnliMnlDRG5CdmJHRnlMVzR5TFdOaGNtOXNnZ1IxYm1sNGdncDFibWw0Y0dGamEyVjBnZ2RpCmRXWmpiMjV1aHdSL0FBQUJoeEFBQUFBQUFBQUFBQUFBQUFBQUFBQUJod1NzRlFBQ01Bb0dDQ3FHU000OUJBTUMKQTBnQU1FVUNJUUNoaUlvV05kMXRhRWxzODVnZzQ1dTMyZGpiSXBkMnpJdTZVRHVhNzBySFdnSWdUbUVXMDU3ZApMTThpMCtlRHhuZVdleVA3WW51SnI3T0JZbnpQK0xNOTdEWT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	carolMacaroonBase64 = "AgEDbG5kAvgBAwoQu1EkS2s/OsizQEI3d1TV0RIBMBoWCgdhZGRyZXNzEgRyZWFkEgV3cml0ZRoTCgRpbmZvEgRyZWFkEgV3cml0ZRoXCghpbnZvaWNlcxIEcmVhZBIFd3JpdGUaIQoIbWFjYXJvb24SCGdlbmVyYXRlEgRyZWFkEgV3cml0ZRoWCgdtZXNzYWdlEgRyZWFkEgV3cml0ZRoXCghvZmZjaGFpbhIEcmVhZBIFd3JpdGUaFgoHb25jaGFpbhIEcmVhZBIFd3JpdGUaFAoFcGVlcnMSBHJlYWQSBXdyaXRlGhgKBnNpZ25lchIIZ2VuZXJhdGUSBHJlYWQAAAYgFTMOWYV4uUaJXRTHbwALf8SNSeraQPu/8g9jGclGtZ0="
	carolHost           = "127.0.0.1:10003"
	carolPubkey         = "029a77b71bf3dfd3530dce0bddf1d070be5ffa71a94e785d7d433c1ddffd550b50"
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

	carolPubkey := carolPubkey
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
		t.Logf("paymentHash: %v", paymentHash.String())

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
			RPreimage:  preimage[:],
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

		// Add Channel Request
		mockRepository.SetGetChannelRequestByPaymentHashMockData(dbMocks.ChannelRequestMockData{
			ChannelRequest: db.ChannelRequest{
				Status:      db.ChannelRequestStatusREQUESTED,
				Pubkey:      carolPubkey,
				Preimage:    preimage[:],
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
		preimage, _ := util.RandomPreimage()
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

func sendPaymentChannel(sendPaymentClient routerrpc.Router_SendPaymentV2Client) <-chan lnrpc.Payment {
	paymentChan := make(chan lnrpc.Payment)

	go func() {
		for {
			payment, err := sendPaymentClient.Recv()

			if err != nil {
				log.Printf("Payment Error: %v", err)
				break
			}

			log.Printf("Payment %v: state %v", payment.PaymentHash, payment.Status)
			paymentChan <- *payment
		}

		close(paymentChan)
	}()

	return paymentChan
}

func subscribeSingleInvoiceChannel(invoicesClient invoicesrpc.InvoicesClient, ctx context.Context, paymentHash []byte) <-chan lnrpc.Invoice {
	invoiceChan := make(chan lnrpc.Invoice)

	go func() {
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

		close(invoiceChan)
	}()

	return invoiceChan
}

func subscribeInvoicesChannel(lightningClient lnrpc.LightningClient, ctx context.Context) <-chan lnrpc.Invoice {
	invoiceChan := make(chan lnrpc.Invoice)

	go func() {
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

		close(invoiceChan)
	}()

	return invoiceChan
}
