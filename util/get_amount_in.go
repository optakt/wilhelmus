package util

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// GetAmountIn adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L52-L59
//
// // given an output amount of an asset and pair reserves, returns a required input amount of the other asset
//
//	function getAmountIn(uint amountOut, uint reserveIn, uint reserveOut) internal pure returns (uint amountIn) {
//		require(amountOut > 0, 'UniswapV2Library: INSUFFICIENT_OUTPUT_AMOUNT');
//		require(reserveIn > 0 && reserveOut > 0, 'UniswapV2Library: INSUFFICIENT_LIQUIDITY');
//		uint numerator = reserveIn.mul(amountOut).mul(1000);
//		uint denominator = reserveOut.sub(amountOut).mul(997);
//		amountIn = (numerator / denominator).add(1);
//	}
func GetAmountIn(amountOut *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	numerator := big.NewInt(0).Mul(reserveIn, amountOut)
	numerator.Mul(numerator, b.D1000)
	denominator := big.NewInt(0).Sub(reserveOut, amountOut)
	denominator.Mul(denominator, b.D997)
	amountIn := big.NewInt(0).Div(numerator, denominator)
	amountIn.Add(amountIn, b.D1)
	return amountIn
}
