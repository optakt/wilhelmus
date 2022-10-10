package util

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// CompoundRate approximation ported from AAVE:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L45
// The rate needs to be given in 10^27 format (Ray).
func CompoundRate(rate *big.Int, exp *big.Int) *big.Int {

	if exp.Cmp(b.D0) == 0 {
		return b.D0
	}

	em1 := big.NewInt(0)
	em1.Sub(em1, b.D0)

	em2 := big.NewInt(0)
	em2.Sub(em2, b.D2)
	if em2.Cmp(b.D0) < 0 {
		em2 = big.NewInt(0)
	}

	rps := big.NewInt(0)
	rps.Div(rate, b.SPY)

	bp2 := big.NewInt(0)
	bp2.Mul(rps, rps)

	bp3 := big.NewInt(0)
	bp3.Mul(bp2, rps)

	t1 := big.NewInt(0)
	t1.Mul(exp, rps)

	t2 := big.NewInt(0)
	t2.Mul(exp, em1)
	t2.Mul(t2, bp2)
	t2.Div(t2, b.D2)

	t3 := big.NewInt(0)
	t3.Mul(exp, em1)
	t3.Mul(t3, em2)
	t3.Mul(t3, bp3)
	t3.Div(t3, b.D6)

	out := big.NewInt(0)
	out.Add(t1, t2)
	out.Add(out, t3)

	return out
}
