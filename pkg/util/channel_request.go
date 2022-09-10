package util

func CalculateLocalFundingAmount(amount int64) int64 {
	return int64(float64(amount) * 1.25)
}