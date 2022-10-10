package position

import (
	"math/big"
)

type Hold struct {
	Size    *big.Int
	Amount0 *big.Int
	Amount1 *big.Int
	Fees0   *big.Int
	Cost0   *big.Int
}

func (h Hold) Value0(price *big.Int) *big.Int {

	amount0 := big.NewInt(0)
	amount0.Mul(h.Amount1, price)

	value0 := big.NewInt(0)
	value0.Add(value0, h.Amount0)
	value0.Add(value0, amount0)
	value0.Sub(value0, h.Fees0)
	value0.Sub(value0, h.Cost0)

	return value0
}
