package util

import (
	"math/big"

	"github.com/optakt/wilhelmus/b"
)

// GetAmountOut adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L42-L49
//
// // given an input amount of an asset and pair reserves, returns the maximum output amount of the other asset
//
//	function getAmountOut(uint amountIn, uint reserveIn, uint reserveOut) internal pure returns (uint amountOut) {
//	    require(amountIn > 0, 'UniswapV2Library: INSUFFICIENT_INPUT_AMOUNT');
//	    require(reserveIn > 0 && reserveOut > 0, 'UniswapV2Library: INSUFFICIENT_LIQUIDITY');
//	    uint amountInWithFee = amountIn.mul(997);
//	    uint numerator = amountInWithFee.mul(reserveOut);
//	    uint denominator = reserveIn.mul(1000).add(amountInWithFee);
//	    amountOut = numerator / denominator;
//	}
func GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	amountInWithFee := big.NewInt(0).Mul(amountIn, b.D997)
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)
	denominator := big.NewInt(0).Mul(reserveIn, b.D1000)
	denominator.Add(denominator, amountInWithFee)
	amountOut := big.NewInt(0).Div(numerator, denominator)
	return amountOut
}
