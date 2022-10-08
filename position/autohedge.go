package position

import (
	"math"
)

type Autohedge struct {
	liquidity float64
	debt      float64
	ratio     float64
	swap      float64
}

func NewAutohedge(input float64, price float64, ratio float64, swap float64) *Autohedge {
	return &Autohedge{
		liquidity: input * input / price,
		debt:      input / price,
		ratio:     ratio,
		swap:      swap,
	}
}

func (a *Autohedge) Value(price float64) float64 {

	volatile := math.Sqrt(a.liquidity / price)
	switch {

	case volatile < a.debt*(1-a.ratio):

		delta := a.debt - volatile
		amountStable := delta * price
		a.liquidity -= (amountStable * delta)
		a.debt -= (2 * delta)

	case volatile > a.debt*(1+a.ratio):
		delta := volatile - a.debt
		amountStable := delta * price
		a.liquidity += (amountStable * delta)
		a.debt += (2 * delta)
	}

	return 2*math.Sqrt(a.liquidity*price) - a.debt*price
}
