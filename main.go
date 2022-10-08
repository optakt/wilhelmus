package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/optakt/dewalt/position"
)

func main() {

	var (
		depositAmountStable float64
		volCurrentPrice     float64
		volTargetPrice      float64
		swapFeeRate         float64
	)

	pflag.Float64VarP(&depositAmountStable, "input", "i", 10000, "stable coin input amount")
	pflag.Float64VarP(&volCurrentPrice, "price", "p", 1000, "volatile asset starting price")
	pflag.Float64VarP(&volTargetPrice, "target", "t", 1200, "volatile asset price target")
	pflag.Float64VarP(&swapFeeRate, "swap", "s", 0.003, "token swap fee rate")

	hold := position.NewHold(depositAmountStable, volCurrentPrice)
	uniswap := position.NewUniswap(depositAmountStable, volCurrentPrice)
	autohedge8 := position.NewAutohedge(depositAmountStable, volCurrentPrice, 0.08, swapFeeRate)
	autohedge4 := position.NewAutohedge(depositAmountStable, volCurrentPrice, 0.04, swapFeeRate)
	autohedge2 := position.NewAutohedge(depositAmountStable, volCurrentPrice, 0.02, swapFeeRate)
	autohedge1 := position.NewAutohedge(depositAmountStable, volCurrentPrice, 0.01, swapFeeRate)

	for price := volCurrentPrice; price <= volTargetPrice; price += 1 {
		fmt.Printf("%5.f %5.f %5.f %5.f %5.f %5.f\n",
			hold.Value(price),
			uniswap.Value(price),
			autohedge8.Value(price),
			autohedge4.Value(price),
			autohedge2.Value(price),
			autohedge1.Value(price),
		)
	}

	pflag.Parse()

	os.Exit(0)
}
