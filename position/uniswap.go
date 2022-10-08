package position

import (
	"math"
)

type Uniswap struct {
	liquidity float64
	debt      float64
}

func NewUniswap(input float64, price float64) *Uniswap {
	return &Uniswap{
		liquidity: input * input / price,
		debt:      input,
	}
}

func (u Uniswap) Value(price float64) float64 {
	return 2*math.Sqrt(u.liquidity*price) - u.debt
}
