package util

import (
	"math/big"
)

// Quote adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L35-L40
//
// // given some amount of an asset and pair reserves, returns an equivalent amount of the other asset
//
//	function quote(uint amountA, uint reserveA, uint reserveB) internal pure returns (uint amountB) {
//	    require(amountA > 0, 'UniswapV2Library: INSUFFICIENT_AMOUNT');
//	    require(reserveA > 0 && reserveB > 0, 'UniswapV2Library: INSUFFICIENT_LIQUIDITY');
//	    amountB = amountA.mul(reserveB) / reserveA;
//	}
func Quote(amountA *big.Int, reserveA *big.Int, reserveB *big.Int) *big.Int {
	amountB := big.NewInt(0).Mul(amountA, reserveB)
	amountB.Div(amountB, reserveA)
	return amountB
}
