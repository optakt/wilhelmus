package position

import (
	"math"
)

type Autohedge struct {
	liquidity float64
	debt      float64
	delta     float64
}

func NewAutohedge(input float64, price float64, delta float64) *Autohedge {
	return &Autohedge{
		liquidity: input * input / price,
		debt:      input / price,
	}
}

func (a *Autohedge) Update(price float64, cost float64, swap float64) {

	volatile := math.Sqrt(a.liquidity / price)
	switch {

	case volatile < a.debt*(1-a.delta):
		delta := a.debt - volatile
		amountVol := delta / (1 - swap)
		amountStable := amountVol * price
		a.liquidity -= (amountStable * amountVol)
		a.debt -= delta

	case volatile > a.debt*(1+a.delta):
		delta := volatile - a.debt
		amountVol := delta / (1 - swap)
		amountStable := amountVol * price
		a.liquidity += (amountStable * amountVol)
		a.debt += delta
	}
}

func (a Autohedge) Value(price float64) float64 {
	return 2*math.Sqrt(price*a.liquidity) - a.debt*price
}
