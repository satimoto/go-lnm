package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ConvertChannelPoint(op string) ([]byte, uint32, error) {
	parts := strings.Split(op, ":")
	if len(parts) != 2 {
		return nil, 0, errors.New("outpoint should be of the form txid:index")
	}
	txid, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid txid: %v", err)
	}

	outputIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid output index: %v", err)
	}
	return ReverseBytes(txid), uint32(outputIndex), nil
}