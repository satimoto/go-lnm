package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/tariff"
)

func (r *SessionResolver) ProcessChargingPeriods(sessionIto *SessionIto, tariffIto *tariff.TariffIto, connectorWattage int32, timeLocation *time.Location, processDatetime time.Time) float64 {
	totalAmount := float64(0)
	lastDatetime := sessionIto.LastUpdated
	numChargingPeriods := len(sessionIto.ChargingPeriods)

	if sessionIto.EndDatetime != nil {
		processDatetime = *sessionIto.EndDatetime
	}

	// Get price components and FLAT price
	priceComponents := getPriceComponents(tariffIto.Elements, timeLocation, sessionIto.StartDatetime, processDatetime, 0, 0, 0)
	flatPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionFLAT)

	if flatPriceComponent != nil {
		totalAmount += flatPriceComponent.Price
	}

	// Calculate session duration
	startDatetime := sessionIto.StartDatetime
	sessionDuration := processDatetime.Sub(startDatetime)
	sessionDurationHours := sessionDuration.Hours()
	sessionTimeVolume := calculateRoundedValue(sessionDurationHours, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)

	if sessionTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionSESSIONTIME); sessionTimePriceComponent != nil {
		// Time charging or not: defined in hours, step_size multiplier: 1 second
		priceRound := getPriceComponentRounding(sessionTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		cost := calculateCost(sessionTimePriceComponent, sessionTimeVolume, 3600)
		totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
		log.Printf("%v: Estimated session time cost: %v", sessionIto.Uid, cost)
	}

	if numChargingPeriods == 0 {
		lastUpdatedDurationSeconds := sessionIto.LastUpdated.Sub(startDatetime).Seconds()
		sessionDurationSeconds := sessionDuration.Seconds()

		if sessionIto.TotalCost != nil && *sessionIto.TotalCost > 0 {
			// Estimation based on total cost
			totalAmount = *sessionIto.TotalCost

			if lastUpdatedDurationSeconds < sessionDurationSeconds {
				totalAmount = (*sessionIto.TotalCost / lastUpdatedDurationSeconds) * sessionDurationSeconds
			}

			log.Printf("%v: Estimated based on total cost + delta: %v", sessionIto.Uid, totalAmount)
		} else {
			// Estimation based on duration and connector wattage
			estimatedMaxEnergy := sessionIto.Kwh

			if estimatedMaxEnergy == 0 {
				estimatedMaxEnergy = (float64(connectorWattage) * sessionDurationHours) / 1000
				log.Printf("%v: Estimated energy based on connector: %v", sessionIto.Uid, estimatedMaxEnergy)
			} else if lastUpdatedDurationSeconds < sessionDurationSeconds {
				estimatedMaxEnergy = (estimatedMaxEnergy / lastUpdatedDurationSeconds) * sessionDurationSeconds
				log.Printf("%v: Estimated energy based on kWh + delta: %v", sessionIto.Uid, estimatedMaxEnergy)
			}

			energyVolume := calculateRoundedValue(estimatedMaxEnergy, db.RoundingGranularityUNIT, db.RoundingRuleROUNDNEAR)

			// TODO: Simulate the charging periods through the time of the session
			//       This includes grouping in periods by restrictions
			if energyPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionENERGY); energyPriceComponent != nil {
				// 	Defined in kWh, step_size multiplier: 1 Wh
				priceRound := getPriceComponentRounding(energyPriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(energyPriceComponent, energyVolume, 1000)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated energy cost: %v", sessionIto.Uid, cost)
			}

			if timePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionTIME); timePriceComponent != nil {
				// Time charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(timePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(timePriceComponent, sessionTimeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated time cost: %v", sessionIto.Uid, cost)
			}
		}
	} else {
		// Estimation based on charging periods
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

				if energyVolume > 0 {
					energyVolume = calculateRoundedValue(energyVolume*ratio, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
					log.Printf("%v: Estimated charging period %v energy + delta: %v", sessionIto.Uid, i, energyVolume)
				}

				if parkingTimeVolume > 0 {
					parkingTimeVolume = calculateRoundedValue(parkingTimeVolume*ratio, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
					log.Printf("%v: Estimated charging period %v parking time + delta: %v", sessionIto.Uid, i, parkingTimeVolume)
				}

				if timeVolume > 0 {
					timeVolume = calculateRoundedValue(timeVolume*ratio, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
					log.Printf("%v: Estimated charging period %v time + delta: %v", sessionIto.Uid, i, timeVolume)
				}
			}

			priceComponents = getPriceComponents(tariffIto.Elements, timeLocation, startDatetime, endDatetime, energyVolume, minPowerVolume, maxPowerVolume)

			if energyPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionENERGY); energyPriceComponent != nil {
				// 	Defined in kWh, step_size multiplier: 1 Wh
				priceRound := getPriceComponentRounding(energyPriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(energyPriceComponent, energyVolume, 1000)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v energy cost: %v", sessionIto.Uid, i, cost)
			}

			if timePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionTIME); timePriceComponent != nil {
				// Time charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(timePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(timePriceComponent, timeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v time cost: %v", sessionIto.Uid, i, cost)
			}

			if parkingTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionPARKINGTIME); parkingTimePriceComponent != nil {
				// Time not charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(parkingTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(parkingTimePriceComponent, parkingTimeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v parking time cost: %v", sessionIto.Uid, i, cost)
			}
		}
	}

	log.Printf("%v: Estimated total cost: %v", sessionIto.Uid, totalAmount)

	return totalAmount
}

func (r *SessionResolver) UpdateSession(session db.Session) {
	/** Session status has changed.
	 *  Send a SessionUpdate notification to the user
	 */

	ctx := context.Background()
	user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID)

	if err != nil {
		metrics.RecordError("LSP051", "Error retrieving user from session", err)
		log.Printf("LSP051: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		return
	}

	// TODO: handle notification failure
	r.SendSessionUpdateNotification(user, session)
}
