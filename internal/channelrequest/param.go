package channelrequest

import "github.com/satimoto/go-datastore/pkg/db"

func NewUpdateChannelRequestParams(channelRequest db.ChannelRequest) db.UpdateChannelRequestParams {
	return db.UpdateChannelRequestParams{
		ID:          channelRequest.ID,
		Status:      channelRequest.Status,
		SettledMsat: channelRequest.SettledMsat,
		FundingTxID: channelRequest.FundingTxID,
		OutputIndex: channelRequest.OutputIndex,
	}
}
