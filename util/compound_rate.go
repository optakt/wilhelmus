package util

const (
	spy = 24 * 365 * 3600
)

// CompoundRate approximation ported from AAVE:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L45
func CompoundRate(rate float64, seconds uint) float64 {

	if seconds == 0 {
		return 0
	}

	exp := float64(seconds)
	em1 := exp - 1
	em2 := exp - 2
	if em2 < 0 {
		em2 = 0
	}

	rps := rate / spy
	bp2 := rps * rps
	bp3 := bp2 * rps

	t1 := exp * rps
	t2 := exp * em1 * bp2 / 2
	t3 := exp * em1 * em2 * bp3 / 6

	out := t1 + t2 + t3

	return out
}
