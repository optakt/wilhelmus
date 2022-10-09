package position

import (
	"math"
)

type Autohedge struct {
	Liquidity  float64
	Profit0    float64
	Profit1    float64
	Principal0 float64
	Yield0     float64
	Debt1      float64
	Interest1  float64
	Fees0      float64
	Cost0      float64
}

func (a *Autohedge) Value0(price float64) float64 {
	position0 := 2 * math.Sqrt(a.Liquidity*price)
	liability0 := (a.Debt1 + a.Interest1) * price
	overhead0 := a.Fees0 + a.Cost0
	return position0 + a.Yield0 - liability0 - overhead0
}
