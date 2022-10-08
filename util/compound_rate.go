package util

import (
	"fmt"
	"math/big"
)

const (
	d7  = 10_000_000
	d27 = 1_000_000_000_000_000_000_000_000_000
)

var (
	b0    = big.NewInt(0)
	b1    = big.NewInt(1)
	b2    = big.NewInt(2)
	b6    = big.NewInt(6)
	b10   = big.NewInt(10)
	b20   = big.NewInt(20)
	b24   = big.NewInt(24)
	b27   = big.NewInt(27)
	b365  = big.NewInt(365)
	b3600 = big.NewInt(3600)
	e20   = big.NewInt(0).Exp(b10, b20, nil) // 10^20
	e27   = big.NewInt(0).Exp(b10, b27, nil) // 10^27
	hpy   = big.NewInt(0).Mul(b24, b365)     // hours per year
	spy   = big.NewInt(0).Mul(hpy, b3600)    // seconds per year
)

// CompoundRate approximation ported from AAVE:
// => https://github.com/aave/protocol-v2/blob/master/contracts/protocol/libraries/math/MathUtils.sol#L45
func CompoundRate(fRate float64, seconds uint) float64 {

	if seconds == 0 {
		return 1
	}

	rate := big.NewInt(int64(fRate * d7)) // get all the decimals in front of the decimal point
	rate.Mul(rate, e20)                   // Ray are 10^27, we already have 10^7, so 10^20 is left

	exp := big.NewInt(int64(seconds))
	em1 := big.NewInt(0).Sub(exp, b1)
	em2 := big.NewInt(0).Sub(exp, b2)
	if em2.Cmp(b0) == 0 {
		em2 = b0
	}

	rps := big.NewInt(0).Div(rate, spy)

	bp2 := big.NewInt(0).Mul(rps, rps)
	bp3 := big.NewInt(0).Mul(bp2, rps)

	t1 := big.NewInt(0).Mul(exp, rps)

	t2 := big.NewInt(0).Mul(exp, em1)
	t2.Mul(t2, bp2)
	t2.Div(t2, b2)

	t3 := big.NewInt(0).Mul(exp, em1)
	t3.Mul(t3, em2)
	t3.Mul(t3, bp3)
	t3.Div(t3, b6)

	res := big.NewInt(0).Add(e27, t1)
	res.Add(res, t2)
	res.Add(res, t3)

	out := big.NewFloat(0).SetInt(res)

	fmt.Printf("rate: %s, exp: %s, em1: %s, em2: %s, rps: %s, bp2: %s, bp3: %s, t1: %s, t2: %s, t3: %s, res: %s\n",
		rate, exp, em1, em2, rps, bp2, bp3, t1, t2, t3, res,
	)

	fOut, _ := out.Float64()

	fmt.Println(fOut / d27)

	return fOut / d27
}
