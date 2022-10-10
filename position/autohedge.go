package position

import (
	"math/big"
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

func (a *Autohedge) Value0(price *big.Int) *big.Int {

	big2 := big.NewInt(2)

	debt0 := big.NewInt(0)
	debt0.Mul(a.Debt1, price)

	interest0 := big.NewInt(0)
	interest0.Mul(a.Interest1, price)

	value0 := big.NewInt(0)
	value0.Mul(a.Liquidity, price)
	value0.Mul(value0, big2)
	value0.Add(value0, a.Yield0)
	value0.Sub(value0, debt0)
	value0.Sub(value0, interest0)
	value0.Sub(value0, a.Fees0)
	value0.Sub(value0, a.Cost0)

	return value0
}
