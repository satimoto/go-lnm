package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/ito"
	metrics "github.com/satimoto/go-lnm/internal/metric"
)

func (r *SessionResolver) ProcessChargingPeriods(sessionIto *ito.SessionIto, tariffIto *ito.TariffIto, connectorWattage int32, timeLocation *time.Location, processDatetime time.Time) (totalAmount, totalEnergy, totalTime float64) {
	lastDatetime := sessionIto.LastUpdated
	numChargingPeriods := len(sessionIto.ChargingPeriods)
	startDatetime := sessionIto.StartDatetime
	lastUpdatedTime := sessionIto.LastUpdated.Sub(startDatetime).Hours()

	if sessionIto.EndDatetime != nil {
		processDatetime = *sessionIto.EndDatetime
	}

	isCdr := sessionIto.IsCdr
	totalCost := sessionIto.TotalCost
	totalEnergy = sessionIto.TotalEnergy
	totalTime = util.DefaultFloat(sessionIto.TotalTime, 0)
	totalParkingTime := sessionIto.TotalParkingTime
	totalSessionTime := sessionIto.TotalSessionTime

	// Get the time from ITO, else calculate it from the session time period
	if sessionIto.TotalTime == nil {
		totalTime = processDatetime.Sub(startDatetime).Hours()
	}

	if totalCost != nil && *totalCost > 0 {
		// ITO has total cost defined
		totalAmount = *totalCost

		if !isCdr && sessionIto.EndDatetime == nil && lastUpdatedTime < totalTime {
			// Calculate delta
			totalAmount = totalTime * (*totalCost / lastUpdatedTime)
			log.Printf("%v: Estimated total cost + delta: %v", sessionIto.Uid, totalAmount)
		}
	} else {
		// Estimation based on charging periods
		flatCost := 0.0
		chargingPeriodsEnergyCost := 0.0
		chargingPeriodsParkingTimeCost := 0.0
		chargingPeriodsTimeCost := 0.0
		sessionTimeCost := 0.0

		// Get price components and FLAT price
		priceComponents := getPriceComponents(tariffIto.Elements, timeLocation, sessionIto.StartDatetime, processDatetime, totalEnergy, 0, 0)

		// Get the parking time from ITO, else set to 0
		if totalParkingTime == nil {
			estimatedParkingTime := 0.0
			totalParkingTime = &estimatedParkingTime
		}

		// Get the session time from ITO, else calculate it from the session time period
		if totalSessionTime == nil {
			estimatedSessionTime := processDatetime.Sub(startDatetime).Hours()
			totalSessionTime = &estimatedSessionTime
		}

		for i, chargingPeriod := range sessionIto.ChargingPeriods {
			startDatetime := chargingPeriod.StartDateTime
			endDatetime := lastDatetime
			energyVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeENERGY)
			minPowerVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeMINCURRENT)
			maxPowerVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeMAXCURRENT)
			parkingTimeVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypePARKINGTIME)
			timeVolume := getVolumeByType(chargingPeriod.Dimensions, db.ChargingPeriodDimensionTypeTIME)

			if !isCdr {
				if i < numChargingPeriods-1 {
					endDatetime = sessionIto.ChargingPeriods[i+1].StartDateTime
				} else if sessionIto.EndDatetime == nil {
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
			}

			chargingPeriodPriceComponents := getPriceComponents(tariffIto.Elements, timeLocation, startDatetime, endDatetime, energyVolume, minPowerVolume, maxPowerVolume)

			if chargingPeriodEnergyPriceComponent := getPriceComponentByType(chargingPeriodPriceComponents, db.TariffDimensionENERGY); chargingPeriodEnergyPriceComponent != nil {
				// 	Defined in kWh, step_size multiplier: 1 Wh
				priceRound := getPriceComponentRounding(chargingPeriodEnergyPriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(chargingPeriodEnergyPriceComponent, energyVolume, 1000)
				chargingPeriodsEnergyCost = calculateRoundedValue(chargingPeriodsEnergyCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v energy cost: %v", sessionIto.Uid, i, cost)
			}

			if chargingPeriodTimePriceComponent := getPriceComponentByType(chargingPeriodPriceComponents, db.TariffDimensionTIME); chargingPeriodTimePriceComponent != nil {
				// Time charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(chargingPeriodTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(chargingPeriodTimePriceComponent, timeVolume, 3600)
				chargingPeriodsTimeCost = calculateRoundedValue(chargingPeriodsTimeCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v time cost: %v", sessionIto.Uid, i, cost)
			}

			if chargingPeriodParkingTimePriceComponent := getPriceComponentByType(chargingPeriodPriceComponents, db.TariffDimensionPARKINGTIME); chargingPeriodParkingTimePriceComponent != nil {
				// Time not charging: defined in hours, step_size multiplier: 1 second
				priceRound := getPriceComponentRounding(chargingPeriodParkingTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(chargingPeriodParkingTimePriceComponent, parkingTimeVolume, 3600)
				chargingPeriodsParkingTimeCost = calculateRoundedValue(chargingPeriodsParkingTimeCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Charging period %v parking time cost: %v", sessionIto.Uid, i, cost)
			}
		}

		if flatPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionFLAT); flatPriceComponent != nil {
			flatCost = flatPriceComponent.Price
		}

		if chargingPeriodsEnergyCost == 0 {
			// Estimate costs if charging periods energy costs is 0
			if totalEnergy == 0 {
				// kWh = hours * ((watts / 1000) * 20%)
				connectorKiloWattage := (float64(connectorWattage) / 1000) * 0.2
				totalEnergy = totalTime * connectorKiloWattage
				log.Printf("%v: Connector kW: %v", sessionIto.Uid, connectorKiloWattage)
				log.Printf("%v: Estimated energy based on kWh: %v", sessionIto.Uid, totalEnergy)
			} else if !isCdr && sessionIto.EndDatetime == nil {
				if lastUpdatedTime > 0 && lastUpdatedTime < totalTime {
					// Estimate energy based on duration
					estimatedEnergy := totalTime * (totalEnergy / lastUpdatedTime)
					totalEnergy = estimatedEnergy
					log.Printf("%v: Energy based on kWh + delta: %v", sessionIto.Uid, totalEnergy)
				}
			}

			priceComponents = getPriceComponents(tariffIto.Elements, timeLocation, sessionIto.StartDatetime, processDatetime, totalEnergy, 0, 0)

			if energyPriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionENERGY); energyPriceComponent != nil {
				// 	Defined in kWh, step_size multiplier: 1 Wh
				energyVolume := calculateRoundedValue(totalEnergy, db.RoundingGranularityUNIT, db.RoundingRuleROUNDNEAR)
				priceRound := getPriceComponentRounding(energyPriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(energyPriceComponent, energyVolume, 1000)
				chargingPeriodsEnergyCost = calculateRoundedValue(chargingPeriodsEnergyCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated energy cost: %v", sessionIto.Uid, cost)
			}
		}

		if chargingPeriodsTimeCost == 0 {
			if timePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionTIME); timePriceComponent != nil {
				// Time charging: defined in hours, step_size multiplier: 1 second
				timeVolume := calculateRoundedValue(totalTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				priceRound := getPriceComponentRounding(timePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(timePriceComponent, timeVolume, 3600)
				chargingPeriodsTimeCost = calculateRoundedValue(chargingPeriodsTimeCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated time cost: %v", sessionIto.Uid, cost)
			}
		}

		if totalParkingTime != nil && chargingPeriodsParkingTimeCost == 0 {
			if parkingTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionPARKINGTIME); parkingTimePriceComponent != nil {
				// Time not charging: defined in hours, step_size multiplier: 1 second
				parkingTimeVolume := calculateRoundedValue(*totalParkingTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				priceRound := getPriceComponentRounding(parkingTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(parkingTimePriceComponent, parkingTimeVolume, 3600)
				totalAmount = calculateRoundedValue(totalAmount+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated parking time cost: %v", sessionIto.Uid, cost)
			}
		}

		if totalSessionTime != nil {
			if sessionTimePriceComponent := getPriceComponentByType(priceComponents, db.TariffDimensionSESSIONTIME); sessionTimePriceComponent != nil {
				// Time charging or not: defined in hours, step_size multiplier: 1 second
				sessionTimeVolume := calculateRoundedValue(*totalSessionTime, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				priceRound := getPriceComponentRounding(sessionTimePriceComponent.PriceRound, db.RoundingGranularityTHOUSANDTH, db.RoundingRuleROUNDNEAR)
				cost := calculateCost(sessionTimePriceComponent, sessionTimeVolume, 3600)
				sessionTimeCost = calculateRoundedValue(sessionTimeCost+cost, priceRound.Granularity, priceRound.Rule)
				log.Printf("%v: Estimated session time cost: %v", sessionIto.Uid, cost)
			}
		}

		totalAmount = chargingPeriodsEnergyCost + chargingPeriodsParkingTimeCost + chargingPeriodsTimeCost + flatCost + sessionTimeCost
	}

	log.Printf("%v: Total cost: %v", sessionIto.Uid, totalAmount)

	return totalAmount, totalEnergy, totalTime
}

func (r *SessionResolver) UpdateSession(session db.Session) {
	/** Session status has changed.
	 *  Send a SessionUpdate notification to the user
	 */

	ctx := context.Background()
	user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID)

	if err != nil {
		metrics.RecordError("LNM051", "Error retrieving user from session", err)
		log.Printf("LNM051: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		return
	}

	// TODO: handle notification failure
	r.SendSessionUpdateNotification(user, session)
}
