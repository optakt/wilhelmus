package position

import (
	"math/big"
)

type Uniswap struct {
	Size      uint64
	Liquidity *big.Int
	Profit0   *big.Int
	Profit1   *big.Int
	Fees0     *big.Int
	Cost0     *big.Int
}

func (u Uniswap) Value0(price *big.Int) *big.Int {

	big2 := big.NewInt(2)

	value0 := big.NewInt(0).Set(u.Liquidity)
	value0.Mul(value0, price)
	value0.Sqrt(value0)
	value0.Mul(value0, big2)
	value0.Sub(value0, u.Fees0)
	value0.Sub(value0, u.Cost0)

	return value0
}
