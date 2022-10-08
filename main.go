package main

import (
	"os"

	"github.com/spf13/pflag"
)

func main() {

	var (
		input float64
		start string
		end   string

		rehedgeRatio   float64
		approveGas     uint64
		swapFee        float64
		swapGas        uint64
		flashFee       float64
		flashGas       uint64
		lendInterest   float64
		lendGas        uint64
		borrowInterest float64
		borrowGas      uint64
	)

	pflag.Float64VarP(&input, "input", "i", 10000, "stable coin input amount")
	pflag.StringVarP(&start, "start", "s", "2021-05-09", "start date for the backtest")
	pflag.StringVarP(&end, "end", "e", "2022-10-08", "end date for the backtest")

	pflag.Float64Var(&rehedgeRatio, "rehedge-ratio", 0.01, "ratio between debt and collateral at which we rehedge")
	pflag.Uint64Var(&approveGas, "approve-gas", 24102, "gas cost for transfer approval")
	pflag.Float64Var(&swapFee, "swap-fee", 0.003, "fee rate for asset swap")
	pflag.Uint64Var(&swapGas, "swap-gas", 172924, "gas cost for asset swap")
	pflag.Float64Var(&flashFee, "flash-fee", 0.0009, "fee rate for flash loan")
	pflag.Uint64Var(&flashGas, "flash-gas", 204493, "gas cost for flash loan")
	pflag.Float64Var(&lendInterest, "lend-interest", 0.004, "interest rate for lending asset")
	pflag.Uint64Var(&lendGas, "lend-gas", 217479, "gas cost for lending asset")
	pflag.Float64Var(&borrowInterest, "borrow-interest", 0.022, "interest rate for borrowing asset")
	pflag.Uint64Var(&borrowGas, "borrow-gas", 295250, "interest rate for borrowing asset")

	pflag.Parse()

	os.Exit(0)
}
