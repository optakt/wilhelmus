package position

import (
	"math"
)

type Uniswap struct {
	Liquidity float64
	Fees      float64
	Cost      float64
}

func (u Uniswap) Value(price float64) float64 {
	return 2*math.Sqrt(u.Liquidity*price) - u.Fees - u.Cost
}
