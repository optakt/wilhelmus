package uniswap

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// GetAmountOut adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L42-L49
func GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	amountInWithFee := big.NewInt(0).Mul(amountIn, b.D997)
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)
	denominator := big.NewInt(0).Mul(reserveIn, b.D1000)
	denominator.Add(denominator, amountInWithFee)
	amountOut := big.NewInt(0).Div(numerator, denominator)
	return amountOut
}
