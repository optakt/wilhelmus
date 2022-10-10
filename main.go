package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/optakt/wilhelmus/aave"
	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/position"
	"github.com/optakt/wilhelmus/station"
)

const (
	statement = `from(bucket: "uniswap")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "ethereum")
	|> filter(fn: (r) => r["pair"] == "WETH/USDC")
	|> filter(fn: (r) => r["_field"] == "volume0" or r["_field"] == "reserve1" or r["_field"] == "reserve0" or r["_field"] == "volume1")
	|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")`
)

func main() {

	var (
		logLevel     string
		writeResults bool
		gasPrices    string

		inputValue       uint64
		startTime        string
		endTime          string
		flagRehedgeRatio uint64

		influxAPI             string
		influxToken           string
		influxOrg             string
		influxTimeout         time.Duration
		influxBucketUniswap   string
		influxBucketPositions string

		flagSwapRate   uint64
		flagFlashRate  uint64
		flagLoanRate   uint64
		flagBorrowRate uint64

		flagApproveGas uint64
		flagSwapGas    uint64
		flagFlashGas   uint64

		flagCreateGas uint64
		flagAddGas    uint64
		flagRemoveGas uint64
		flagCloseGas  uint64

		flagLendGas  uint64
		flagClaimGas uint64

		flagBorrowGas   uint64
		flagIncreaseGas uint64
		flagDecreaseGas uint64
		flagRepayGas    uint64
	)

	pflag.StringVarP(&logLevel, "log-level", "l", "info", "Zerolog logger logging message severity")
	pflag.BoolVarP(&writeResults, "write-results", "w", false, "whether to write the results back to InfluxDB")
	pflag.StringVarP(&gasPrices, "gas-prices", "g", "gas-prices.csv", "CSV file for average gas price per day")

	pflag.Uint64VarP(&inputValue, "input-value", "i", 1_000_000, "stable coin input amount")
	pflag.StringVarP(&startTime, "start-time", "s", "2021-10-07T00:00:00Z", "start timestamp for the backtest")
	pflag.StringVarP(&endTime, "end-time", "e", "2022-10-07T23:59:59Z", "end timestamp for the backtest")
	pflag.Uint64VarP(&flagRehedgeRatio, "rehedge-ratio", "r", 100, "ratio between debt and collateral at which we rehedge (in 1/10000)")

	pflag.StringVarP(&influxAPI, "influx-api", "a", "https://eu-central-1-1.aws.cloud2.influxdata.com", "InfluxDB API URL")
	pflag.StringVarP(&influxToken, "influx-token", "t", "", "InfluxDB authentication token")
	pflag.StringVarP(&influxOrg, "influx-org", "o", "optakt", "InfluxDB organization name")
	pflag.DurationVarP(&influxTimeout, "influx-timeout", "u", 15*time.Minute, "InfluxDB query HTTP request timeout")
	pflag.StringVar(&influxBucketUniswap, "influx-bucket-uniswap", "uniswap", "InfluxDB bucket name for Uniswap metrics")
	pflag.StringVar(&influxBucketPositions, "influx-bucket-positions", "positions", "InfluxDB bucket for position values")

	pflag.Uint64Var(&flagSwapRate, "swap-rate", 30, "fee rate for asset swap (in 1/10000)")
	pflag.Uint64Var(&flagFlashRate, "flash-rate", 9, "fee rate for flash loan (in 1/10000)")
	pflag.Uint64Var(&flagLoanRate, "lend-rate", 50, "interest rate for lending asset (in 1/10000)")
	pflag.Uint64Var(&flagBorrowRate, "borrow-rate", 250, "interest rate for borrowing asset (in 1/10000)")

	pflag.Uint64Var(&flagApproveGas, "approve-gas", 24102, "gas cost for transfer approval")
	pflag.Uint64Var(&flagSwapGas, "swap-gas", 181133, "gas cost for asset swap")
	pflag.Uint64Var(&flagFlashGas, "flash-gas", 204493, "gas cost for flash loan")

	pflag.Uint64Var(&flagCreateGas, "provide-gas", 157880, "gas cost for creating liquidity position")
	pflag.Uint64Var(&flagAddGas, "add-gas", 130682, "gas cost for adding liquidity")
	pflag.Uint64Var(&flagRemoveGas, "remove-gas", 161841, "gas cost to remove liquidity")
	pflag.Uint64Var(&flagCloseGas, "close-gas", 207111, "gas cost for close liquidity position")

	pflag.Uint64Var(&flagLendGas, "lend-gas", 217479, "gas cost for lending asset")
	pflag.Uint64Var(&flagClaimGas, "claim-gas", 333793, "gas cost to claim back loan")

	pflag.Uint64Var(&flagBorrowGas, "borrow-gas", 295250, "gas cost for borrowing asset")
	pflag.Uint64Var(&flagDecreaseGas, "unborrow-gas", 193729, "gas cost for reducing debt")
	pflag.Uint64Var(&flagIncreaseGas, "increase-gas", 271980, "gas cost for increasing debt")
	pflag.Uint64Var(&flagRepayGas, "repay-gas", 188929, "gas cost to repay full debt")

	pflag.Parse()

	log := zerolog.New(os.Stdout)
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Fatal().Err(err).Str("log_level", logLevel).Msg("invalid log level")
	}
	log = log.Level(level)

	station, err := station.New(gasPrices)
	if err != nil {
		log.Fatal().Err(err).Str("gas_prices", gasPrices).Msg("could not create gas station")
	}

	client := influxdb2.NewClientWithOptions(influxAPI, influxToken,
		influxdb2.DefaultOptions().SetHTTPRequestTimeout(uint(influxTimeout.Seconds())),
	)

	outbound := client.WriteAPI(influxOrg, influxBucketPositions)
	go func() {
		for err := range outbound.Errors() {
			log.Fatal().Err(err).Msg("encountered InfluxDB error")
		}
	}()

	inbound := client.QueryAPI(influxOrg)
	query := fmt.Sprintf(statement, startTime, endTime)
	result, err := inbound.Query(context.Background(), query)
	if err != nil {
		log.Fatal().Err(err).Msg("could not execute query")
	}
	if !result.Next() {
		log.Fatal().Msg("no records found")
	}
	err = result.Err()
	if err != nil {
		log.Fatal().Err(result.Err()).Msg("could not stream first record")
	}

	// Convert the USD value given as input into a big integer.
	input0 := big.NewInt(0).SetUint64(inputValue)
	input0.Mul(input0, b.D6) // USDC has 6 decimals, we want to operate at the most granular level

	// Convert the hedge ratio to big integer.
	rehedgeRatio := big.NewInt(0).SetUint64(flagRehedgeRatio)

	// Convert the fee/interest rates into big integers.
	// TODO: check the precision on Uniswap v2 when calculating the swap fee
	swapRate := big.NewInt(0).SetUint64(flagSwapRate)
	swapRate.Mul(swapRate, b.E23)

	flashRate := big.NewInt(0).SetUint64(flagFlashRate)
	flashRate.Mul(flashRate, b.E23)

	loanRate := big.NewInt(0).SetUint64(flagLoanRate)
	loanRate.Mul(loanRate, b.E23)

	borrowRate := big.NewInt(0).SetUint64(flagBorrowRate)
	borrowRate.Mul(borrowRate, b.E23)

	// Convert the gas costs into big integers.
	approveGas := big.NewInt(0).SetUint64(flagApproveGas) // approve ERC20 transfer
	swapGas := big.NewInt(0).SetUint64(flagSwapGas)       // swap assets on Uniswap v2 pair
	flashGas := big.NewInt(0).SetUint64(flagFlashGas)     // take out a flash loan on Aave

	createGas := big.NewInt(0).SetUint64(flagCreateGas) // create liquidity position on Uniswap v2
	addGas := big.NewInt(0).SetUint64(flagAddGas)       // add liquidity on Uniswap v2
	removeGas := big.NewInt(0).SetUint64(flagRemoveGas) // remove liquidity on Uniswap v2
	// closeCas := big.NewInt(0).SetUint64(flagCloseGas)   // close liquidity position on Uniswap v2

	lendGas := big.NewInt(0).SetUint64(flagLendGas) // lend asset on Aave
	// claimGas := big.NewInt(0).SetUint64(flagClaimGas) // claim loan plus yield on Aave

	borrowGas := big.NewInt(0).SetUint64(flagBorrowGas)     // borrow asset on Aave
	increaseGas := big.NewInt(0).SetUint64(flagIncreaseGas) // increase debt on Aaave
	decreaseGas := big.NewInt(0).SetUint64(flagDecreaseGas) // decrease debt on Aave
	// repayGas := big.NewInt(0).SetUint64(flagRepayGas)       // repoy loan on Aave

	// Read the first record to initialize the positions.
	record := result.Record()
	timestamp := record.Time()
	values := record.Values()

	// amountInMultplied := amountIn * 997
	// numerator := amountInMultiplied * reserveOut
	// denominator := resorveIn * 1000 + amountInMultiplied
	// amountOut := numerator / denominator

	// The values from InfluxDB come as hex-encoded strings for now, so convert
	// them back to the original big integers read from the contracts.
	// NOTE: this is because InfluxDB doesn't support number above 64 bits, and
	// with `float64` we get too much imprecision. QuestDB supports 256-bit
	// integers and might be the better option.
	reserve0hex := values["reserve0"].(string)
	reserve1hex := values["reserve1"].(string)

	reserve0bytes, err := hex.DecodeString(reserve0hex)
	if err != nil {
		log.Fatal().Err(err).Msg("could not decode reserve0")
	}
	reserve1bytes, err := hex.DecodeString(reserve1hex)
	if err != nil {
		log.Fatal().Err(err).Msg("could not decode reserve1")
	}

	reserve0 := big.NewInt(0).SetBytes(reserve0bytes)
	reserve1 := big.NewInt(0).SetBytes(reserve1bytes)

	// The price is a ratio that can be smaller than zero.
	// TODO: check what precision Uniswap uses here.
	price := big.NewInt(0).Set(reserve0)
	price.Mul(price, b.E18)
	price.Div(price, reserve1)

	gasPrice1, err := station.Gasprice(timestamp)
	if err != nil {
		log.Fatal().Err(err).Time("timestamp", timestamp).Msg("could not get gas price for timestamp")
	}

	feesHold0 := big.NewInt(0).Set(input0)
	feesHold0.Div(feesHold0, b.D2)
	feesHold0.Mul(feesHold0, swapRate)
	feesHold0.Div(feesHold0, b.E27)

	hold0 := big.NewInt(0).Set(input0)
	hold0.Sub(hold0, input0)
	hold0.Div(hold0, b.D2)

	hold1 := big.NewInt(0).Set(hold0)
	hold1.Mul(hold1, b.E18)
	hold1.Div(hold0, price)

	costHold0 := big.NewInt(0).Set(approveGas)
	costHold0.Add(costHold0, swapGas)
	costHold0.Mul(costHold0, gasPrice1)
	costHold0.Mul(costHold0, price)

	hold := position.Hold{
		Size:    input0,
		Amount0: hold0,
		Amount1: hold1,
		Fees0:   feesHold0,
		Cost0:   costHold0,
	}

	liquidityUni := big.NewInt(0).Set(hold0)
	liquidityUni.Mul(liquidityUni, hold1)

	costCreate0 := big.NewInt(0).Set(createGas)
	costCreate0.Mul(costCreate0, gasPrice1)
	costCreate0.Mul(costCreate0, price)

	feesUni0 := big.NewInt(0).Set(hold.Fees0)

	costUni0 := big.NewInt(0).Set(costHold0)
	costUni0.Add(costUni0, costCreate0)

	uniswap := position.Uniswap{
		Size:      input0,
		Liquidity: liquidityUni,
		Fees0:     feesUni0,
		Cost0:     costUni0,
	}

	// TODO: figure out how to correctly apply with 10^27 values
	auto1 := big.NewInt(0).Set(b.E27)
	auto1.Sub(auto1, swapRate)
	auto1.Mul(auto1, b.E27)
	auto1.Div(flashRate, auto1)
	auto1.Add(b.E27, auto1)
	auto1.Mul(auto1, price)
	auto1.Div(input0, auto1)

	autoFee1 := big.NewInt(0).Set(auto1)
	autoFee1.Mul(autoFee1, flashRate)
	autoFee1.Div(autoFee1, b.E27)

	autoFee0 := big.NewInt(0).Set(b.E27)
	autoFee0.Sub(autoFee0, swapRate)
	autoFee0.Mul(autoFee0, b.E27)
	autoFee0.Div(price, autoFee0)
	autoFee0.Mul(autoFee1, autoFee0)

	auto0 := big.NewInt(0).Set(input0)
	auto0.Sub(auto0, autoFee0)

	liquidityAuto := big.NewInt(0).Set(auto0)
	liquidityAuto.Mul(liquidityAuto, auto1)

	principal0 := big.NewInt(0).Set(auto1)
	principal0.Mul(principal0, price)
	principal0.Add(principal0, auto0)

	costHedge0 := big.NewInt(0).Set(approveGas)
	costHedge0.Mul(costHedge0, b.D2)
	costHedge0.Add(costHedge0, flashGas)
	costHedge0.Add(costHedge0, lendGas)
	costHedge0.Add(costHedge0, borrowGas)
	costHedge0.Mul(costHedge0, gasPrice1)
	costHedge0.Mul(costHedge0, price)

	costAuto0 := big.NewInt(0).Set(costUni0)
	costAuto0.Add(costAuto0, costHedge0)

	autohedge := position.Autohedge{
		Size:       input0,
		Rehedge:    rehedgeRatio,
		Liquidity:  liquidityAuto,
		Principal0: principal0,
		Yield0:     b.D0,
		Debt1:      auto1,
		Interest1:  b.D0,
		Fees0:      autoFee0,
		Cost0:      costAuto0,
	}

	log.Info().
		Time("timestamp", timestamp).
		Str("price", price.String()).
		Str("hold", big.NewInt(0).Div(hold.Value0(price), b.D6).String()).
		Str("uniswap", big.NewInt(0).Div(uniswap.Value0(price), b.D6).String()).
		Str("autohedge", big.NewInt(0).Div(autohedge.Value0(price), b.D6).String()).
		Msg("original positions created")

	if writeResults {
		writeHold(timestamp, price, hold, outbound)
		writeUniswap(timestamp, price, uniswap, outbound)
		writeAutohedge(timestamp, price, autohedge, outbound)
	}

	last := timestamp
	for result.Next() {

		record := result.Record()
		timestamp := record.Time()
		values := record.Values()

		rs0 := values["reserve0"].(string)
		rs1 := values["reserve1"].(string)

		reserve0 := b.FromHex(rs0)
		reserve1 := b.FromHex(rs1)

		s0 := values["volume0"].(string)
		s1 := values["volume1"].(string)

		volume0 := b.FromHex(s0)
		volume1 := b.FromHex(s1)

		log := log.With().
			Time("timestamp", timestamp).
			Logger()

		log.Debug().
			Float64("reserve0", b.ToFloat(reserve0, 6)).
			Float64("reserve1", b.ToFloat(reserve1, 18)).
			Float64("volume0", b.ToFloat(volume0, 6)).
			Float64("volume1", b.ToFloat(volume1, 18)).
			Msg("extracted datapoint from record")

		elapsed := big.NewInt(int64(timestamp.Sub(last).Seconds()))

		realLoanRate := aave.CalculateCompoundedInterest(loanRate, elapsed)
		yieldDelta0 := big.NewInt(0).Add(autohedge.Principal0, autohedge.Yield0)
		yieldDelta0.Mul(yieldDelta0, realLoanRate)
		yieldDelta0.Div(yieldDelta0, b.E27)
		autohedge.Yield0.Add(autohedge.Yield0, yieldDelta0)

		realBorrowRate := aave.CalculateCompoundedInterest(borrowRate, elapsed)
		interestDelta1 := big.NewInt(0).Add(autohedge.Debt1, autohedge.Interest1)
		interestDelta1.Mul(interestDelta1, realBorrowRate)
		interestDelta1.Div(interestDelta1, b.E27)
		autohedge.Interest1.Add(autohedge.Interest1, interestDelta1)

		last = timestamp

		log.Debug().
			Float64("principal0", b.ToFloat(autohedge.Principal0, 6)).
			Float64("yield0", b.ToFloat(autohedge.Yield0, 6)).
			Float64("gain0", b.ToFloat(yieldDelta0, 6)).
			Float64("debt1", b.ToFloat(autohedge.Debt1, 18)).
			Float64("interest1", b.ToFloat(autohedge.Interest1, 18)).
			Float64("loss1", b.ToFloat(interestDelta1, 18)).
			Msg("compounded principal yield and debt interest")

		liquidity := big.NewInt(0).Mul(reserve0, reserve1)

		log.Debug().
			Float64("liquidity", b.ToFloat(liquidity, 0)).
			Float64("uniswap", b.ToFloat(uniswap.Liquidity, 0)).
			Float64("autohedge", b.ToFloat(autohedge.Liquidity, 0)).
			Msg("preparing to calculate uniswap returns")

		uni0 := math.Sqrt(uniswap.Liquidity * price)
		uni1 := uni0 / price
		shareUni := uniswap.Liquidity / liquidity
		profitUni0 := shareUni * volume0
		profitUni1 := shareUni * volume1

		uniswap.Profit0 += profitUni0
		uniswap.Profit1 += profitUni1
		uniswap.Liquidity = (uni0 + profitUni0) * (uni1 + profitUni1)

		log.Debug().
			Float64("uni0", uni0).
			Float64("uni1", uni1).
			Float64("share", shareUni).
			Float64("profit0", profitUni0).
			Float64("profit1", profitUni1).
			Float64("liquidity", uniswap.Liquidity).
			Msg("added profit to uniswap position")

		auto0 := math.Sqrt(autohedge.Liquidity * price)
		auto1 := auto0 / price
		shareAuto := autohedge.Liquidity / liquidity
		profitAuto0 := shareAuto * volume0
		profitAuto1 := shareAuto * volume1

		autohedge.Profit0 += profitAuto0
		autohedge.Profit1 += profitAuto1
		autohedge.Liquidity = (auto0 + profitAuto0) * (auto1 + profitAuto1)

		log.Debug().
			Float64("auto0", auto0).
			Float64("auto1", auto1).
			Float64("share", shareAuto).
			Float64("profit0", profitAuto0).
			Float64("profit1", profitAuto1).
			Float64("liquidity", autohedge.Liquidity).
			Msg("added profit to autohedge position")

		position0 := math.Sqrt(autohedge.Liquidity * price)
		position1 := position0 / price

		// Uniswap v2 swap:
		// amountInMultplied := amountIn * 997
		// numerator := amountInMultiplied * reserveOut
		// denominator := resorveIn * 1000 + amountInMultiplied
		// amountOut := numerator / denominator
		// amountIn - amountOut = fee => solve?

		switch {

		case position1 < (autohedge.Debt1+autohedge.Interest1)*(1-rehedgeRatio):

			delta1 := autohedge.Debt1 + autohedge.Interest1 - position1
			out1 := delta1 * (1 + swapRate)
			out0 := out1 * price

			autohedge.Liquidity = (position0 - out0) * (position1 - out1)
			autohedge.Debt1 -= (out0/price + delta1)
			autohedge.Fees0 += out0 * swapRate
			autohedge.Cost0 += (swapGas + decreaseGas + addGas) * gasPrice1 * price

			log.Debug().
				Float64("position0", position0).
				Float64("position1", position1).
				Float64("delta1", delta1).
				Float64("out1", out1).
				Float64("out0", out0).
				Float64("liquidity", autohedge.Liquidity).
				Float64("debt1", autohedge.Debt1).
				Float64("fees0", autohedge.Fees0).
				Float64("cost0", autohedge.Cost0).
				Msg("decreased debt to rehedge autoswap position")

		case position1 > (autohedge.Debt1+autohedge.Interest1)*(1+rehedgeRatio):

			delta1 := position1 - autohedge.Debt1 - autohedge.Interest1
			in1 := delta1
			in0 := delta1 * price

			autohedge.Liquidity = (position0 + in0) * (position1 + in1)
			autohedge.Debt1 += (in0/price + delta1*(1+swapRate))
			autohedge.Fees0 += in0 * swapRate
			autohedge.Cost0 += (swapGas + increaseGas + removeGas) * gasPrice1 * price

			log.Debug().
				Float64("position0", position0).
				Float64("position1", position1).
				Float64("delta1", delta1).
				Float64("in1", in1).
				Float64("in0", in0).
				Float64("liquidity", autohedge.Liquidity).
				Float64("debt1", autohedge.Debt1).
				Float64("fees0", autohedge.Fees0).
				Float64("cost0", autohedge.Cost0).
				Msg("increased debt to rehedge autoswap position")
		}

		if writeResults {
			writeHold(timestamp, price, hold, outbound)
			writeUniswap(timestamp, price, uniswap, outbound)
			writeAutohedge(timestamp, price, autohedge, outbound)
		}

		log.Info().
			Float64("price", price*d18/d6).
			Float64("hold0", hold.Value0(price)/d6).
			Float64("uniswap0", uniswap.Value0(price)/d6).
			Float64("autohedge0", autohedge.Value0(price)/d6).
			Msg("updated position valuations")
	}

	err = result.Err()
	if err != nil {
		log.Fatal().Err(err).Msg("could not finish streaming records")
	}

	os.Exit(0)
}
