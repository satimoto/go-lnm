package util

import (
	"crypto/rand"

	"github.com/lightningnetwork/lnd/lntypes"
)

func RandomPreimage() (*lntypes.Preimage, error) {
	paymentPreimage := &lntypes.Preimage{}

	if _, err := rand.Read(paymentPreimage[:]); err != nil {
		return &lntypes.Preimage{}, err
	}

	return paymentPreimage, nil
}