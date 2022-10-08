package position

import (
	"math"
)

type Autohedge struct {
	Liquidity float64 // total liquidity of the position is stable
	Principal float64 // stable loaned out to gain yield
	Yield     float64 // stable earned on lending out stable
	Debt      float64 // volatile borrowed against interest to hedge
	Interest  float64 // volatile owed on borrowing volatile
	Fees      float64 // on-chain fees for various DeFi applications
	Cost      float64 // transaction fees to pay for gas costs
}

func (a *Autohedge) Value(price float64) float64 {
	position := 2 * math.Sqrt(a.Liquidity*price)
	liability := (a.Debt + a.Interest) * price
	overhead := a.Fees + a.Cost
	return position + a.Yield - liability - overhead
}
