package kraken

import (
	"encoding/json"
	"io"
)

func UnmarshalTickerResponse(body io.ReadCloser) (*TickerResponse, error) {
	response := TickerResponse{}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
