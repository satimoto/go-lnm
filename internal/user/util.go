package user

import (
	"math"

	"github.com/satimoto/go-datastore/pkg/db"
)

func GetEstimatedChargePower(user db.User, connector db.Connector) float64 {
	connectorKiloWattage := float64(connector.Wattage) / 1000
	
	if connector.PowerType == db.PowerTypeDC {
		if user.BatteryPowerDc.Valid {
			// 80% of user DC battery power
			maxPowerDc := math.Min(connectorKiloWattage, user.BatteryPowerDc.Float64)

			return maxPowerDc * 0.8
		}

		maxPowerDc := math.Min(connectorKiloWattage, 100)

		return maxPowerDc * 0.8
	}

	if user.BatteryPowerAc.Valid {
		// 80% of user AC battery power
		maxPowerAc := math.Min(connectorKiloWattage, user.BatteryPowerAc.Float64)

		return maxPowerAc * 0.8
	}

	maxPowerAc := math.Min(connectorKiloWattage, 7.5)

	return maxPowerAc * 0.8
}
