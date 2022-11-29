package util

func CalculateLocalFundingAmount(amount int64) int64 {
	localFundingAmount := int64(float64(amount) * 1.25)

	if localFundingAmount < 20000 {
		localFundingAmount = 20000
	}

	return localFundingAmount
}