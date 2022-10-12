package position

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

type Uniswap struct {
	Size      uint64
	Liquidity *big.Int
	Fees0     *big.Int
	Cost0     *big.Int
	Profit0   *big.Int
}

func (u Uniswap) Value0(reserve0 *big.Int, reserve1 *big.Int) *big.Int {

	sqrtReserve0 := big.NewInt(0).Sqrt(reserve0)
	sqrtReserve1 := big.NewInt(0).Sqrt(reserve1)

	value0 := big.NewInt(0).Mul(u.Liquidity, sqrtReserve0)
	value0.Div(value0, sqrtReserve1)
	value0.Mul(value0, b.D2)

	value0.Sub(value0, u.Fees0)
	value0.Sub(value0, u.Cost0)

	return value0
}
