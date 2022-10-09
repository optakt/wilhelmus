package position

import (
	"math"
)

type Uniswap struct {
	Liquidity float64
	Fees0     float64
	Cost0     float64
}

func (u Uniswap) Value0(price float64) float64 {
	return 2*math.Sqrt(u.Liquidity*price) - u.Fees0 - u.Cost0
}
