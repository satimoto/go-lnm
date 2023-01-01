package session

import (
	"math"
	"strings"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/ito"
	"github.com/satimoto/go-lsp/pkg/util"
)

func CalculatePriceInvoiced(sessionInvoices []db.SessionInvoice) (priceFiat float64, priceMsat int64) {
	for _, sessionInvoice := range sessionInvoices {
		priceFiat += sessionInvoice.PriceFiat
		priceMsat += sessionInvoice.PriceMsat
	}

	return priceFiat, priceMsat
}

func CalculateTotalInvoiced(sessionInvoices []db.SessionInvoice) (totalFiat float64, totalMsat int64) {
	for _, sessionInvoice := range sessionInvoices {
		totalFiat += sessionInvoice.TotalFiat
		totalMsat += sessionInvoice.TotalMsat
	}

	return totalFiat, totalMsat
}

func CalculateCommission(amount float64, commissionPercent float64, taxPercent float64) (total float64, commission float64, tax float64) {
	commission = (amount / 100.0) * commissionPercent
	total = amount + commission
	tax = (total / 100.0) * taxPercent
	total += tax

	return total, commission, tax
}

func ReverseCommission(total float64, commissionPercent float64, taxPercent float64) (amount float64, commission float64, tax float64) {
	amount = total / (1 + (taxPercent / 100))
	tax = total - amount
	amount = amount / (1 + (commissionPercent / 100))
	commission = total - amount - tax

	return amount, commission, tax
}

func calculateCost(priceComponent *ito.PriceComponentIto, volume float64, factor float64) float64 {
	stepRound := getPriceComponentRounding(priceComponent.StepRound, db.RoundingGranularityUNIT, db.RoundingRuleROUNDUP)
	priceRound := getPriceComponentRounding(priceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
	pricePerStep := priceComponent.Price / factor * float64(priceComponent.StepSize)
	volumeFactor := volume * factor
	steps := volumeFactor / float64(priceComponent.StepSize)
	roundedSteps := calculateRoundedValue(steps, stepRound.Granularity, stepRound.Rule)

	return calculateRoundedValue(pricePerStep*roundedSteps, priceRound.Granularity, priceRound.Rule)
}

func getPriceComponentRounding(priceComponentRounding *ito.PriceComponentRoundingIto, granularity db.RoundingGranularity, rule db.RoundingRule) ito.PriceComponentRoundingIto {
	if priceComponentRounding != nil {
		return *priceComponentRounding
	}

	return ito.PriceComponentRoundingIto{
		Granularity: granularity,
		Rule:        rule,
	}
}

func getPriceComponentByType(priceComponents []*ito.PriceComponentIto, tariffDimension db.TariffDimension) *ito.PriceComponentIto {
	for _, priceComponent := range priceComponents {
		if priceComponent.Type == tariffDimension {
			return priceComponent
		}
	}

	return nil
}

func getPriceComponents(elements []*ito.ElementIto, timeLocation *time.Location, startDatetime time.Time, endDatetime time.Time, energy float64, minPower float64, maxPower float64) []*ito.PriceComponentIto {
	list := []*ito.PriceComponentIto{}
	startDatetimeAtLocation := startDatetime.In(timeLocation)
	weekday := strings.ToUpper(startDatetimeAtLocation.Weekday().String())
	duration := int32(endDatetime.Sub(startDatetime).Seconds())

	for _, element := range elements {
		if element.Restrictions == nil {
			list = append(list, element.PriceComponents...)
		} else {
			restrictions := element.Restrictions
			restrictionStartTime := parseTimeOfDay(restrictions.StartTime, timeLocation, startDatetimeAtLocation)
			restrictionStartDate := parseDate(restrictions.StartDate, startDatetimeAtLocation)
			restrictionEndTime := parseTimeOfDay(restrictions.EndTime, timeLocation, startDatetimeAtLocation)
			restrictionEndDate := parseDate(restrictions.EndDate, startDatetimeAtLocation)

			if (restrictionStartTime == nil || startDatetimeAtLocation.After(*restrictionStartTime)) &&
				(restrictionEndTime == nil || startDatetimeAtLocation.Before(*restrictionEndTime)) &&
				(restrictionStartDate == nil || startDatetimeAtLocation.After(*restrictionStartDate)) &&
				(restrictionEndDate == nil || startDatetimeAtLocation.Before(*restrictionEndDate)) &&
				(restrictions.MinKwh == nil || energy >= *restrictions.MinKwh) &&
				(restrictions.MaxKwh == nil || (energy > 0 && energy < *restrictions.MaxKwh)) &&
				(restrictions.MinPower == nil || minPower >= *restrictions.MinPower) &&
				(restrictions.MaxPower == nil || (maxPower > 0 && maxPower < *restrictions.MaxPower)) &&
				(restrictions.MinDuration == nil || duration >= *restrictions.MinDuration) &&
				(restrictions.MaxDuration == nil || duration < *restrictions.MaxDuration) &&
				(len(restrictions.DayOfWeek) == 0 || dbUtil.StringsContainString(restrictions.DayOfWeek, weekday)) {
				list = append(list, element.PriceComponents...)
			}
		}
	}

	return list
}

func hasUnsettledInvoices(sessionInvoices []db.SessionInvoice) bool {
	for _, sessionInvoice := range sessionInvoices {
		if !sessionInvoice.IsSettled {
			return true
		}
	}

	return false
}

func parseDate(dateStr *string, datetime time.Time) *time.Time {
	if dateStr != nil {
		date, err := time.Parse("2006-01-02", *dateStr)

		if err == nil {
			return &date
		}
	}

	return nil
}

func parseTimeOfDay(timeStr *string, timeLocation *time.Location, datetime time.Time) *time.Time {
	if timeStr != nil {
		splitTime := strings.Split(*timeStr, ":")
		date := time.Date(
			datetime.Year(),
			datetime.Month(),
			datetime.Day(),
			int(dbUtil.ParseInt32(splitTime[0], 0)),
			int(dbUtil.ParseInt32(splitTime[1], 0)),
			0,
			0,
			datetime.Location())

		return &date
	}

	return nil
}

func calculateRoundedValue(value float64, granularity db.RoundingGranularity, rule db.RoundingRule) float64 {
	factor := float64(1)

	if granularity == db.RoundingGranularityTENTH {
		factor = 10
	} else if granularity == db.RoundingGranularityHUNDREDTH {
		factor = 100
	} else if granularity == db.RoundingGranularityTHOUSANDTH {
		factor = 1000
	}

	if rule == db.RoundingRuleROUNDDOWN {
		return math.Floor(value*factor) / factor
	} else if rule == db.RoundingRuleROUNDNEAR {
		return math.Round(value*factor) / factor
	}

	return math.Ceil(value*factor) / factor
}

func getVolumeByType(dimensions []*ito.ChargingPeriodDimensionIto, dimensionType db.ChargingPeriodDimensionType) float64 {
	for _, dimension := range dimensions {
		if dimension.Type == dimensionType {
			return dimension.Volume
		}
	}

	return 0
}

func calculateInvoiceInterval(wattage int32) time.Duration {
	// Calculate an internal that equates to approximately 1 kWh per payment
	duration := time.Duration(100000/wattage) * time.Minute

	return util.MaxDuration(time.Minute, duration)
}
