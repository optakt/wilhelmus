package util

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// CalculateCompoundedInterest adopted from AAVE v2:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L32-L70
func CalculateCompoundedInterest(rate *big.Int, exp *big.Int) *big.Int {

	if exp.Cmp(b.D0) == 0 {
		return big.NewInt(0)
	}

	em1 := big.NewInt(0).Sub(exp, b.D1)
	em2 := big.NewInt(0).Sub(exp, b.D2)
	if em2.Cmp(b.D0) < 0 {
		em2 = big.NewInt(0)
	}

	rps := big.NewInt(0).Div(rate, b.SPY)

	bp2 := big.NewInt(0).Mul(rps, rps)
	bp2.Add(bp2, b.HALF)
	bp2.Div(bp2, b.E27)

	bp3 := big.NewInt(0).Mul(bp2, rps)
	bp3.Add(bp3, b.HALF)
	bp3.Div(bp3, b.E27)

	t1 := big.NewInt(0).Mul(exp, rps)

	t2 := big.NewInt(0).Mul(exp, em1)
	t2.Mul(t2, bp2)
	t2.Div(t2, b.D2)

	t3 := big.NewInt(0).Mul(exp, em1)
	t3.Mul(t3, em2)
	t3.Mul(t3, bp3)
	t3.Div(t3, b.D6)

	out := big.NewInt(0).Add(t1, t2)
	out.Add(out, t3)

	return out
}
