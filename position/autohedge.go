package position

import (
	"math/big"

	"github.com/optakt/wilhelmus/util"
)

type Autohedge struct {
	Size       *big.Int
	Rehedge    *big.Int
	Liquidity  *big.Int
	Profit0    *big.Int
	Profit1    *big.Int
	Principal0 *big.Int
	Yield0     *big.Int
	Debt1      *big.Int
	Interest1  *big.Int
	Fees0      *big.Int
	Cost0      *big.Int
}

func (a *Autohedge) Value0(reserve0 *big.Int, reserve1 *big.Int) *big.Int {

	big2 := big.NewInt(2)

	debt0 := util.Quote(a.Debt1, reserve1, reserve0)

	interest0 := util.Quote(a.Interest1, reserve1, reserve0)

	value0 := big.NewInt(0).Mul(a.Liquidity, reserve0)
	value0.Div(value0, reserve1)
	value0.Sqrt(value0)
	value0.Mul(value0, big2)

	value0.Add(value0, a.Yield0)
	value0.Sub(value0, debt0)
	value0.Sub(value0, interest0)
	value0.Sub(value0, a.Fees0)
	value0.Sub(value0, a.Cost0)

	return value0
}
