package uniswap

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// GetAmountIn adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L52-L59
func GetAmountIn(amountOut *big.Int, reserveOut *big.Int, reserveIn *big.Int) *big.Int {
	numerator := big.NewInt(0).Mul(reserveIn, amountOut)
	numerator.Mul(numerator, b.D1000)
	denominator := big.NewInt(0).Sub(reserveOut, amountOut)
	denominator.Mul(denominator, b.D1000)
	amountIn := big.NewInt(0).Mul(numerator, denominator)
	amountIn.Add(amountIn, b.D1)
	return amountIn
}
