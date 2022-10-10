package uniswap

import (
	"math/big"
)

// Quote adopted from Uniswap v2:
// => https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol#L35-L40
func Quote(amountA *big.Int, reserveA *big.Int, reserveB *big.Int) *big.Int {
	amountB := big.NewInt(0).Mul(amountA, reserveB)
	amountB.Div(amountB, reserveA)
	return amountB
}
