package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-lsp/internal/ito"
	metrics "github.com/satimoto/go-lsp/internal/metric"
)

func (r *SessionResolver) ProcessChargingPeriods(sessionIto *ito.SessionIto, tariffIto *ito.TariffIto, connectorWattage int32, timeLocation *time.Location, processDatetime time.Time) float64 {
	totalAmount := float64(0)
	lastDatetime := sessionIto.LastUpdated
	numChargingPeriods := len(sessionIto.ChargingPeriods)
	startDatetime := sessionIto.StartDatetime

	if sessionIto.EndDatetime != nil {
		processDatetime = *sessionIto.EndDatetime
	}

	totalCost := sessionIto.TotalCost
	totalEnergy := sessionIto.TotalEnergy
	totalTime := sessionIto.TotalTime
	totalParkingTime := sessionIto.TotalParkingTime
	totalSessionTime := sessionIto.TotalSessionTime

	// Get price components and FLAT price
	priceComponents := getPriceComponents(tariffIto.Elements, timeLocation, sessionIto.StartDatetime, processDatetime, 0, 0, 0)
	flatPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionFLAT)

	if flatPriceComponent != nil {
		totalAmount += flatPriceComponent.Price
	}

	// Get the time from sessionIto, else calculate it from the session time period
	if totalTime == nil {
		estimatedTime := processDatetime.Sub(startDatetime).Hours()
		totalTime = &estimatedTime
	}

	// Get the parking time from sessionIto, else set to 0
	if totalParkingTime == nil {
		estimatedParkingTime := 0.0
		totalParkingTime = &estimatedParkingTime
	}

	// Get the session time from sessionIto, else calculate it from the session time period
	if totalSessionTime == nil {
		estimatedSessionTime := processDatetime.Sub(startDatetime).Hours()
		totalSessionTime = &estimatedSessionTime
	}

	sessionTimeVolume := calculateRoundedValue(*totalSessionTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)

	if sessionTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionSESSIONTIME); sessionTimePriceComponent != nil {
		// Time charging or not: defined in hours, step_size multiplier: 1 second
		priceRound := getPriceComponentRounding(sessionTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
		cost := calculateCost(sessionTimePriceComponent, sessionTimeVolume, 3600)
		totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
		log.Printf("%v: Estimated session time cost: %v", sessionIto.Uid, cost)
	}

	if numChargingPeriods == 0 {
		lastUpdatedTime := sessionIto.LastUpdated.Sub(startDatetime).Hours()

		if totalCost != nil && *totalCost > 0 {
			// Estimation based on total cost
			totalAmount = *totalCost

			if lastUpdatedTime < *totalTime {
				// Calculate delta
				totalAmount = *totalTime * (*totalCost / lastUpdatedTime)
				log.Printf("%v: Estimated total cost + delta: %v", sessionIto.Uid, totalAmount)
			}
		} else {
			if totalEnergy > 0 {
				if lastUpdatedTime > 0 && lastUpdatedTime < *totalTime {
					// Estimation based on duration and connector wattage
					estimatedEnergy := *totalTime * (totalEnergy / lastUpdatedTime)
					totalEnergy = estimatedEnergy
					log.Printf("%v: Energy based on kWh + delta: %v", sessionIto.Uid, totalEnergy)
				}
			} else {
				// kWh = hours * (watts / 1000)
				connectorKiloWattage := (float64(connectorWattage) / 1000)
				log.Printf("%v: Connector kW: %v", sessionIto.Uid, connectorKiloWattage)
		
				if connectorKiloWattage > 100 {
					connectorKiloWattage = 100
					log.Printf("%v: Connector capped kW: %v", sessionIto.Uid, connectorKiloWattage)
				}
		
				totalEnergy = *totalTime * connectorKiloWattage
				log.Printf("%v: Estimated energy based on kWh: %v", sessionIto.Uid, totalEnergy)
			}

			energyVolume := calculateRoundedValue(totalEnergy, db.RoundingGranularityUNIT, db.RoundingRuleROUNDNEAR)
			timeVolume := calculateRoundedValue(*totalTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
			parkingTimeVolume := calculateRoundedValue(*totalTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)

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
				cost := calculateCost(timePriceComponent, timeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated time cost: %v", sessionIto.Uid, cost)
			}

			if parkingTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionPARKINGTIME); parkingTimePriceComponent != nil {
				// Time not charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(parkingTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(parkingTimePriceComponent, parkingTimeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated parking time cost: %v", sessionIto.Uid, cost)
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
