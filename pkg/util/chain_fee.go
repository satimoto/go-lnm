package util

import "github.com/lightningnetwork/lnd/lnwallet/chainfee"

func FeePerVByte(satPerKw int64) uint64 {
	satPerKweight := chainfee.SatPerKWeight(satPerKw)
	satPerVbyte := uint64(satPerKweight.FeePerKVByte()) / 1000

	return satPerVbyte
}
