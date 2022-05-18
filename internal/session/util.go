package session

import (
	"math"
	"strings"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/tariff"
)

func CalculateAmountInvoiced(sessionInvoices []db.SessionInvoice) float64 {
	amountFiat := float64(0)

	for _, sessionInvoice := range sessionInvoices {
		amountFiat += sessionInvoice.AmountFiat
	}

	return amountFiat
}

func CalculateCommission(amount float64, commissionPercent float64, taxPercent float64) (total float64, commission float64, tax float64) {
	commission = (amount / 100.0) * commissionPercent
	total = amount + commission
	tax = (total / 100.0) * taxPercent
	total += tax

	return total, commission, tax
}

func calculateCost(priceComponent *tariff.PriceComponentIto, volume float64, factor float64) float64 {
	stepRound := getPriceComponentRounding(priceComponent.StepRound, db.RoundingGranularityUNIT, db.RoundingRuleROUNDUP)
	priceRound := getPriceComponentRounding(priceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
	pricePerStep := priceComponent.Price / factor * float64(priceComponent.StepSize)
	volumeFactor := volume * factor
	steps := volumeFactor / float64(priceComponent.StepSize)
	roundedSteps := calculateRoundedValue(steps, stepRound.Granularity, stepRound.Rule)

	return calculateRoundedValue(pricePerStep*roundedSteps, priceRound.Granularity, priceRound.Rule)
}

func getPriceComponentRounding(priceComponentRounding *tariff.PriceComponentRoundingIto, granularity db.RoundingGranularity, rule db.RoundingRule) tariff.PriceComponentRoundingIto {
	if priceComponentRounding != nil {
		return *priceComponentRounding
	}

	return tariff.PriceComponentRoundingIto{
		Granularity: granularity,
		Rule:        rule,
	}
}

func getPriceComponentByType(priceComponents []*tariff.PriceComponentIto, tariffDimension db.TariffDimension) *tariff.PriceComponentIto {
	for _, priceComponent := range priceComponents {
		if priceComponent.Type == tariffDimension {
			return priceComponent
		}
	}

	return nil
}

func getPriceComponents(elements []*tariff.ElementIto, startDatetime time.Time, endDatetime time.Time, energy float64, minPower float64, maxPower float64) []*tariff.PriceComponentIto {
	list := []*tariff.PriceComponentIto{}
	weekday := strings.ToUpper(startDatetime.Weekday().String())
	duration := int32(endDatetime.Sub(startDatetime).Seconds())

	for _, element := range elements {
		if element.Restrictions == nil {
			list = append(list, element.PriceComponents...)
		} else {
			restrictions := element.Restrictions
			restrictionStartTime := parseTimeOfDay(restrictions.StartTime, startDatetime)
			restrictionStartDate := parseDate(restrictions.StartDate, startDatetime)
			restrictionEndTime := parseTimeOfDay(restrictions.EndTime, startDatetime)
			restrictionEndDate := parseDate(restrictions.EndDate, startDatetime)

			if (restrictionStartTime == nil || startDatetime.After(*restrictionStartTime)) &&
				(restrictionEndTime == nil || startDatetime.Before(*restrictionEndTime)) &&
				(restrictionStartDate == nil || startDatetime.After(*restrictionStartDate)) &&
				(restrictionEndDate == nil || startDatetime.Before(*restrictionEndDate)) &&
				(restrictions.MinKwh == nil || energy >= *restrictions.MinKwh) &&
				(restrictions.MaxKwh == nil || (energy > 0 && energy < *restrictions.MaxKwh)) &&
				(restrictions.MinPower == nil || minPower >= *restrictions.MinPower) &&
				(restrictions.MaxPower == nil || (maxPower > 0 && maxPower < *restrictions.MaxPower)) &&
				(restrictions.MinDuration == nil || duration >= *restrictions.MinDuration) &&
				(restrictions.MaxDuration == nil || duration < *restrictions.MaxDuration) &&
				(len(restrictions.DayOfWeek) == 0 || util.StringsContainString(restrictions.DayOfWeek, weekday)) {
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

func parseTimeOfDay(timeStr *string, datetime time.Time) *time.Time {
	if timeStr != nil {
		splitTime := strings.Split(*timeStr, ":")
		date := time.Date(
			datetime.Year(),
			datetime.Month(),
			datetime.Day(),
			int(util.ParseInt32(splitTime[0], 0)),
			int(util.ParseInt32(splitTime[1], 0)),
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
	} else if granularity == db.RoundingGranularityHUNDRETH {
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

func getVolumeByType(dimensions []*ChargingPeriodDimensionIto, dimensionType db.ChargingPeriodDimensionType) float64 {
	for _, dimension := range dimensions {
		if dimension.Type == dimensionType {
			return dimension.Volume
		}
	}

	return 0
}

func calculateInvoiceInterval(wattage int32) time.Duration {
	if wattage < 25000 {
		return time.Duration(25000/wattage) * time.Minute
	}

	return time.Minute
}
