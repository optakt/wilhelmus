package b

import (
	"math"
	"math/big"
)

func ToFloat(b *big.Int, decimals uint) float64 {
	n, _ := big.NewFloat(0).SetInt(b).Float64()
	d := math.Pow(10, float64(decimals))
	f := n / d
	return f
}
