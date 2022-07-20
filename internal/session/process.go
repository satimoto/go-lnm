package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/tariff"
)

func (r *SessionResolver) ProcessChargingPeriods(sessionIto *SessionIto, tariffIto *tariff.TariffIto, processDatetime time.Time) float64 {
	totalAmount := float64(0)
	lastDatetime := sessionIto.LastUpdated
	numChargingPeriods := len(sessionIto.ChargingPeriods)

	if sessionIto.EndDatetime != nil {
		processDatetime = *sessionIto.EndDatetime
	}

	priceComponents := getPriceComponents(tariffIto.Elements, sessionIto.StartDatetime, processDatetime, 0, 0, 0)
	flatPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionFLAT)

	if flatPriceComponent != nil {
		totalAmount += flatPriceComponent.Price
	}

	for i, chargingPeriod := range sessionIto.ChargingPeriods {
		startDatetime := chargingPeriod.StartDateTime
		endDatetime := lastDatetime
		energyVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeENERGY)
		minPowerVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeMINCURRENT)
		maxPowerVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeMAXCURRENT)
		parkingTimeVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypePARKINGTIME)
		timeVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeTIME)

		if i < numChargingPeriods-1 {
			endDatetime = sessionIto.ChargingPeriods[i+1].StartDateTime
		} else {
			// Latest charging period
			ratio := float64(1)
			chargingPeriodDuration := endDatetime.Sub(startDatetime).Hours()
			currentDuration := processDatetime.Sub(startDatetime).Hours()

			if chargingPeriodDuration > 0 && currentDuration > 0 {
				ratio = (1 / chargingPeriodDuration) * currentDuration
			}

			energyVolume = calculateRoundedValue(energyVolume*ratio, db.RoundingGranularityUNIT, db.RoundingRuleROUNDNEAR)
			parkingTimeVolume = calculateRoundedValue(parkingTimeVolume*ratio, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
			timeVolume = calculateRoundedValue(timeVolume*ratio, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		}

		priceComponents = getPriceComponents(tariffIto.Elements, startDatetime, endDatetime, energyVolume, minPowerVolume, maxPowerVolume)

		if energyPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionENERGY); energyPriceComponent != nil {
			cost := calculateCost(energyPriceComponent, energyVolume, 1)
			totalAmount = calculateRoundedValue(totalAmount+cost, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		}

		if timePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionTIME); timePriceComponent != nil {
			cost := calculateCost(timePriceComponent, timeVolume, 3600)
			totalAmount = calculateRoundedValue(totalAmount+cost, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		}

		if parkingTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionPARKINGTIME); parkingTimePriceComponent != nil {
			cost := calculateCost(parkingTimePriceComponent, parkingTimeVolume, 3600)
			totalAmount = calculateRoundedValue(totalAmount+cost, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		}
	}

	return totalAmount
}

func (r *SessionResolver) UpdateSession(ctx context.Context, session db.Session) {
	/** Session status has changed.
	 *  Send a SessionUpdate notification to the user
	 */
	 
	user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID)

	if err != nil {
		util.LogOnError("LSP051", "Error retrieving user from session", err)
		log.Printf("LSP051: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		return
	}

	// TODO: handle notification failure
	r.SendSessionUpdateNotification(user, session)
}
