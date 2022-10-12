package position

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/util"
)

type Autohedge struct {
	Size       uint64
	Rehedge    *big.Int
	Liquidity  *big.Int
	Principal0 *big.Int
	Debt1      *big.Int
	Yield0     *big.Int
	Interest1  *big.Int
	Fees0      *big.Int
	Cost0      *big.Int
	Profit0    *big.Int
	Count      uint
}

func (a *Autohedge) Value0(reserve0 *big.Int, reserve1 *big.Int) *big.Int {

	sqrtReserve0 := big.NewInt(0).Sqrt(reserve0)
	sqrtReserve1 := big.NewInt(0).Sqrt(reserve1)

	value0 := big.NewInt(0).Mul(a.Liquidity, sqrtReserve0)
	value0.Div(value0, sqrtReserve1)
	value0.Mul(value0, b.D2)

	debt0 := util.Quote(a.Debt1, reserve1, reserve0)
	interest0 := util.Quote(a.Interest1, reserve1, reserve0)

	value0.Sub(value0, debt0)
	value0.Sub(value0, interest0)
	value0.Add(value0, a.Yield0)

	value0.Sub(value0, a.Fees0)
	value0.Sub(value0, a.Cost0)

	return value0
}
