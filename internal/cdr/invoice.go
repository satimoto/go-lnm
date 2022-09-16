package cdr

import (
	"context"
	"errors"
	"log"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-datastore/pkg/util"
)

func (r *CdrResolver) IssueInvoiceRequest(ctx context.Context, userID int64, promotionCode string, amountMsat int64) error {
	circuitUser, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, userID)

	if err != nil {
		util.LogOnError("LSP111", "Error retrieving user", err)
		log.Printf("LSP111: UserID=%v", userID)
		return errors.New("error retrieving user")
	}

	promotion, err := r.PromotionRepository.GetPromotionByCode(ctx, promotionCode)

	if err != nil {
		util.LogOnError("LSP112", "Error retrieving promotion", err)
		log.Printf("LSP112: Code=%v", promotionCode)
		return errors.New("error retrieving promotion")
	}

	getUnsettledInvoiceRequestByPromotionCodeParams := db.GetUnsettledInvoiceRequestByPromotionCodeParams{
		UserID: circuitUser.ID,
		Code:   promotionCode,
	}

	invoiceRequest, err := r.InvoiceRequestRepository.GetUnsettledInvoiceRequestByPromotionCode(ctx, getUnsettledInvoiceRequestByPromotionCodeParams)

	if err == nil {
		updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)
		updateInvoiceRequestParams.AmountMsat = updateInvoiceRequestParams.AmountMsat + amountMsat

		_, err = r.InvoiceRequestRepository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

		if err != nil {
			util.LogOnError("LSP113", "Error updating invoice request", err)
			log.Printf("LSP113: Params=%#v", updateInvoiceRequestParams)
			return errors.New("error updating invoice request")
		}
	} else {
		createInvoiceRequestParams := db.CreateInvoiceRequestParams{
			UserID:      circuitUser.ID,
			PromotionID: promotion.ID,
			AmountMsat:  amountMsat,
			IsSettled:   false,
		}

		_, err = r.InvoiceRequestRepository.CreateInvoiceRequest(ctx, createInvoiceRequestParams)

		if err != nil {
			util.LogOnError("LSP114", "Error creating invoice request", err)
			log.Printf("LSP114: Params=%#v", createInvoiceRequestParams)
			return errors.New("error creating invoice request")
		}
	}

	return nil
}
