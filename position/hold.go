package position

import (
	"math/big"

	"github.com/optakt/wilhelmus/util"
)

type Hold struct {
	Size    uint64
	Amount0 *big.Int
	Amount1 *big.Int
	Fees0   *big.Int
	Cost0   *big.Int
}

func (h Hold) Value0(reserve0 *big.Int, reserve1 *big.Int) *big.Int {

	amount0 := util.Quote(h.Amount1, reserve1, reserve0)

	value0 := big.NewInt(0).Add(h.Amount0, amount0)
	value0.Sub(value0, h.Fees0)
	value0.Sub(value0, h.Cost0)

	return value0
}
