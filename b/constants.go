package b

import (
	"math/big"
)

var (
	HPY  = big.NewInt(0).Mul(D365, D24)  // hours per year
	SPY  = big.NewInt(0).Mul(HPY, D3600) // seconds per year
	HALF = big.NewInt(0).Div(E27, D2)    // Half Ray
)
