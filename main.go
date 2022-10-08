package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

const (
	d18 = 1_000_000_000_000_000_000
)

func main() {

	var (
		depositAmountStable float64
		volAssetPrice       float64
		flashFeeRate        float64
		swapFeeRate         float64
		borrowFeeRate       float64
		txCost              float64
	)

	pflag.Float64VarP(&depositAmountStable, "input", "i", 2000, "stable coin input amount")
	pflag.Float64VarP(&volAssetPrice, "price", "p", 2000, "volatile price in stable")
	pflag.Float64VarP(&flashFeeRate, "flash", "f", 0.0005, "flash loan fee rate")
	pflag.Float64VarP(&swapFeeRate, "swap", "s", 0.003, "token swap fee rate")
	pflag.Float64VarP(&borrowFeeRate, "borrow", "b", 0.02, "normal loan fee rate")
	pflag.Float64VarP(&txCost, "cost", "c", 5, "transactions gas cost")

	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	log := zerolog.New(os.Stderr)

	inputAmountVol := depositAmountStable / (volAssetPrice * (1 + (flashFeeRate / (1 - swapFeeRate))))
	flashFeeVol := inputAmountVol * flashFeeRate
	flashFeeStable := flashFeeVol * volAssetPrice / (1 - swapFeeRate)
	inputAmountStable := depositAmountStable - flashFeeStable
	totalEquivalentStable := inputAmountVol*volAssetPrice + inputAmountStable
	totalEquivalentVol := inputAmountVol + inputAmountStable/volAssetPrice
	collateralRatio := totalEquivalentVol / inputAmountVol

	log.Info().
		Float64("deposit_amount_stable", depositAmountStable).
		Float64("input_amount_volatile", inputAmountVol).
		Float64("flash_fee_volatile", flashFeeVol).
		Float64("flash_fee_stable", flashFeeStable).
		Float64("input_amount_stable", inputAmountStable).
		Float64("total_equivalent_stable", totalEquivalentStable).
		Float64("total_equivalent_volatile", totalEquivalentVol).
		Float64("collateral_ratio", collateralRatio).
		Msg("amounts computed")

	pflag.Parse()

	os.Exit(0)
}
