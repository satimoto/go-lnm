package cdr

import (
	"context"
	"errors"
	"log"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-datastore/util"
	"github.com/satimoto/go-lsp/internal/session"
)

func (r *CdrResolver) ProcessCdr(ctx context.Context, cdr db.Cdr) error {
	/** Cdr has been created.
	 *  Calculate final invoiced amount.
	 *  Issue final invoice or rebate if overpaid
	 */

	if !cdr.AuthorizationID.Valid {
		log.Printf("LSP035: Cdr AuthorizationID is nil")
		log.Printf("LSP035: CdrUid=%v", cdr.Uid)
		return errors.New("Cdr AuthorizationID is nil")
	}

	// TODO: How to deal with Sessions and CDRs with no AuthorizationID
	sess, err := r.SessionResolver.Repository.GetSessionByAuthorizationID(ctx, cdr.AuthorizationID.String)

	if err != nil {
		util.LogOnError("LSP043", "Error retrieving cdr session", err)
		log.Printf("LSP043: CdrUid=%v, AuthorizationID=%v", cdr.Uid, cdr.AuthorizationID)
		return errors.New("Error retrieving cdr session")
	}

	sessionInvoices, err := r.SessionResolver.Repository.ListSessionInvoices(ctx, sess.ID)

	if err != nil {
		util.LogOnError("LSP044", "Error retrieving session invoices", err)
		log.Printf("LSP044: SessionUid=%v", sess.Uid)
		return errors.New("Error retrieving session invoices")
	}

	invoicedAmount := session.CalculateAmountInvoiced(sessionInvoices)
	location, err := r.SessionResolver.LocationResolver.Repository.GetLocation(ctx, sess.LocationID)

	if err != nil {
		util.LogOnError("LSP045", "Error retrieving session location", err)
		log.Printf("LSP045: SessionUid=%v, LocationID=%v", sess.Uid, sess.LocationID)
		return errors.New("Error retrieving session location")
	}

	user, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, sess.UserID)

	if err != nil {
		util.LogOnError("LSP046", "Error retrieving session user", err)
		log.Printf("LSP046: SessionUid=%v, UserID=%v", sess.Uid, sess.UserID)
		return errors.New("Error retrieving session user")
	}

	taxPercent := r.SessionResolver.CountryAccountResolver.GetTaxPercentByCountry(ctx, location.Country, util.GetEnvFloat64("DEFAULT_TAX_PERCENT", 20))
	totalAmount, _, _ := session.CalculateCommission(cdr.TotalCost, user.CommissionPercent, taxPercent)

	// TODO: Issue rebate if over paid
	// TODO: Issue referer satsback
	if totalAmount > invoicedAmount {
		invoiceAmount, invoiceCommission, invoiceTax := session.CalculateCommission(totalAmount-invoicedAmount, user.CommissionPercent, taxPercent)

		r.SessionResolver.IssueLightningInvoice(ctx, sess, invoiceAmount, invoiceCommission, invoiceTax)
	}

	return nil
}
