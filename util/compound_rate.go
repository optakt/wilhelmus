package util

import (
	"math/big"
)

var (
	b0    = big.NewInt(0)
	b1    = big.NewInt(1)
	b2    = big.NewInt(2)
	b6    = big.NewInt(6)
	b24   = big.NewInt(24)
	b365  = big.NewInt(365)
	b3600 = big.NewInt(3600)
)

var (
	hpy = big.NewInt(0).Mul(b24, b365)
	spy = big.NewInt(0).Mul(hpy, b3600)
)

// CompoundRate approximation ported from AAVE:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L45
func CompoundRate(rate *big.Int, exp *big.Int) *big.Int {

	if exp.Cmp(b0) == 0 {
		return b0
	}

	em1 := big.NewInt(0)
	em1.Sub(em1, b1)

	em2 := big.NewInt(0)
	em2.Sub(em2, b2)
	if em2.Cmp(b0) < 0 {
		em2 = big.NewInt(0)
	}

	rps := big.NewInt(0)
	rps.Div(rate, spy)

	bp2 := big.NewInt(0)
	bp2.Mul(rps, rps)

	bp3 := big.NewInt(0)
	bp3.Mul(bp2, rps)

	t1 := big.NewInt(0)
	t1.Mul(exp, rps)

	t2 := big.NewInt(0)
	t2.Mul(exp, em1)
	t2.Mul(t2, bp2)
	t2.Div(t2, b2)

	t3 := big.NewInt(0)
	t3.Mul(exp, em1)
	t3.Mul(t3, em2)
	t3.Mul(t3, bp3)
	t3.Div(t3, b6)

	out := big.NewInt(0)
	out.Add(t1, t2)
	out.Add(out, t3)

	return out
}
