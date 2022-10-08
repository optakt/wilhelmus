package position

import (
	"math"
)

type Autohedge struct {
	Ratio     float64
	Liquidity float64
	Debt      float64
	Fees      float64
	Cost      float64
	Interest  float64
}

func (a *Autohedge) Value(price float64) float64 {

	volatile := math.Sqrt(a.Liquidity / price)
	stable := a.Liquidity / volatile
	switch {

	case volatile < a.Debt*(1-a.Ratio):

		delta := a.Debt - volatile
		amountStable := delta * price
		a.Liquidity = (volatile - delta) * (stable - amountStable)
		a.Debt -= (2 * delta)

	case volatile > a.Debt*(1+a.Ratio):
		delta := volatile - a.Debt
		amountStable := delta * price
		a.Liquidity = (volatile + delta) * (stable + amountStable)
		a.Debt += (2 * delta)
	}

	return 2*math.Sqrt(a.Liquidity*price) - a.Debt*price
}
