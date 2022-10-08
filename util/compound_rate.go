package util

// CompoundRate approximation ported from AAVE:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L45
func CompoundRate(rate float64, seconds uint) float64 {

	if seconds == 0 {
		return 0
	}

	exp := float64(seconds)

	expMinus1 := exp - 1
	expMinus2 := exp - 2
	if expMinus2 < 0 {
		expMinus2 = 0
	}

	ratePerSecond := rate / 365 * 24 * 3600

	basePow2 := ratePerSecond * ratePerSecond
	basePow3 := basePow2 * ratePerSecond

	term2 := exp * expMinus1 * basePow2 / 2
	term3 := exp * expMinus1 * expMinus2 * basePow3 / 6

	result := exp*ratePerSecond + term2 + term3

	return result
}
