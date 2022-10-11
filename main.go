package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/position"
	"github.com/optakt/wilhelmus/station"
	"github.com/optakt/wilhelmus/util"
)

const (
	statement = `from(bucket: "uniswap")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r["_measurement"] == "ethereum")
	|> filter(fn: (r) => r["pair"] == "USDC/WETH")
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

		flagSwapRate   float64
		flagFlashRate  float64
		flagLoanRate   float64
		flagBorrowRate float64

		flagTransferGas uint64
		flagApproveGas  uint64
		flagSwapGas     uint64
		flagFlashGas    uint64

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

	pflag.Float64Var(&flagSwapRate, "swap-rate", 0.003, "fee rate for asset swap")
	pflag.Float64Var(&flagFlashRate, "flash-rate", 0.0009, "fee rate for flash loan")
	pflag.Float64Var(&flagLoanRate, "lend-rate", 0.005, "interest rate for lending asset")
	pflag.Float64Var(&flagBorrowRate, "borrow-rate", 0.025, "interest rate for borrowing asset")

	pflag.Uint64Var(&flagTransferGas, "transfer-gas", 65601, "gas cost for token transfer")
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
	input0.Mul(input0, b.E6) // USDC has 6 decimals, we want to operate at the most granular level

	// Convert the hedge ratio to big integer.
	rehedgeRatio := big.NewInt(0).SetUint64(flagRehedgeRatio)

	// We keep track of the flash rate as 1/1000 units
	swapRate := big.NewInt(int64(flagSwapRate * 1_000))

	// We keep rate of loan interest rates as 1/10^27 units (Ray)
	flashRate := big.NewInt(0).Mul(big.NewInt(int64(flagFlashRate*10_000)), b.E23)
	loanRate := big.NewInt(0).Mul(big.NewInt(int64(flagLoanRate*10_000)), b.E23)
	borrowRate := big.NewInt(0).Mul(big.NewInt(int64(flagBorrowRate*10_000)), b.E23)

	// Convert the gas costs into big integers.
	approveGas := big.NewInt(0).SetUint64(flagApproveGas) // approve ERC20 transfer
	swapGas := big.NewInt(0).SetUint64(flagSwapGas)       // swap assets on Uniswap v2 pair
	flashGas := big.NewInt(0).SetUint64(flagFlashGas)     // take out a flash loan on Aave

	createGas := big.NewInt(0).SetUint64(flagCreateGas) // create liquidity position on Uniswap v2
	// addGas := big.NewInt(0).SetUint64(flagAddGas)       // add liquidity on Uniswap v2
	// removeGas := big.NewInt(0).SetUint64(flagRemoveGas) // remove liquidity on Uniswap v2
	// closeCas := big.NewInt(0).SetUint64(flagCloseGas)   // close liquidity position on Uniswap v2

	lendGas := big.NewInt(0).SetUint64(flagLendGas) // lend asset on Aave
	// claimGas := big.NewInt(0).SetUint64(flagClaimGas) // claim loan plus yield on Aave

	borrowGas := big.NewInt(0).SetUint64(flagBorrowGas) // borrow asset on Aave
	// increaseGas := big.NewInt(0).SetUint64(flagIncreaseGas) // increase debt on Aaave
	// decreaseGas := big.NewInt(0).SetUint64(flagDecreaseGas) // decrease debt on Aave
	// repayGas := big.NewInt(0).SetUint64(flagRepayGas)       // repoy loan on Aave

	// Read the first record to initialize the positions.
	record := result.Record()
	timestamp := record.Time()
	values := record.Values()

	// The values from InfluxDB come as hex-encoded strings for now, so convert
	// them back to the original big integers read from the contracts.
	// NOTE: this is because InfluxDB doesn't support number above 64 bits, and
	// with `float64` we get too much imprecision. QuestDB supports 256-bit
	// integers and might be the better option.
	reserve0 := b.FromHex(values["reserve0"])
	reserve1 := b.FromHex(values["reserve1"])

	gasPrice1, err := station.Gasprice(timestamp)
	if err != nil {
		log.Fatal().Err(err).Time("timestamp", timestamp).Msg("could not get gas price for timestamp")
	}

	holdDiv := big.NewInt(0).Add(b.D2000, swapRate)

	hold0 := big.NewInt(0).Mul(input0, b.D1000)
	hold0.Div(hold0, holdDiv)

	hold1 := util.Quote(hold0, reserve0, reserve1)

	fee0 := big.NewInt(0).Sub(input0, hold0)
	fee0.Sub(fee0, hold0)

	costHold1 := big.NewInt(0).Add(approveGas, swapGas)
	costHold1.Mul(costHold1, gasPrice1)

	costHold0 := util.Quote(costHold1, reserve1, reserve0)

	hold := position.Hold{
		Size:    input0,
		Amount0: hold0,
		Amount1: hold1,
		Fees0:   fee0,
		Cost0:   costHold0,
	}

	liquidityUni := big.NewInt(0).Mul(hold0, hold1)

	feesUni0 := big.NewInt(0).Set(hold.Fees0)

	costUni1 := big.NewInt(0).Add(approveGas, swapGas)
	costUni1.Add(costUni1, createGas)
	costUni1.Mul(costUni1, gasPrice1)

	costUni0 := util.Quote(costUni1, reserve1, reserve0)

	uniswap := position.Uniswap{
		Size:      input0,
		Liquidity: liquidityUni,
		Profit0:   big.NewInt(0),
		Profit1:   big.NewInt(1),
		Fees0:     feesUni0,
		Cost0:     costUni0,
	}

	autoDivA := big.NewInt(0).Mul(flashRate, swapRate) // 0.003 * 0.0009
	autoDivB := big.NewInt(0).Mul(flashRate, b.E3)     // 0.0009

	autoDiv := big.NewInt(0).Add(autoDivA, autoDivB) // 0.0009 + 0.003 * 0.0009
	autoDiv.Add(autoDiv, b.E30)                      // 1 + 0.0009 + 0.003 * 0.0009

	auto0 := big.NewInt(0).Mul(input0, b.E30)
	auto0.Div(auto0, autoDiv)

	auto1 := util.Quote(auto0, reserve0, reserve1)

	liquidityAuto := big.NewInt(0).Mul(auto0, auto1)

	principal0 := big.NewInt(0).Add(auto0, auto0)

	autoFee0 := big.NewInt(0).Sub(input0, auto0)

	costAuto1 := big.NewInt(0).Add(flashGas, createGas)
	costAuto1.Add(costAuto1, approveGas)
	costAuto1.Add(costAuto1, lendGas)
	costAuto1.Add(costAuto1, borrowGas)
	costAuto1.Add(costAuto1, approveGas)
	costAuto1.Add(costAuto1, swapGas)
	costAuto1.Mul(costAuto1, gasPrice1)

	costAuto0 := util.Quote(costAuto1, reserve1, reserve0)

	autohedge := position.Autohedge{
		Size:       input0,
		Rehedge:    rehedgeRatio,
		Liquidity:  liquidityAuto,
		Profit0:    big.NewInt(0),
		Profit1:    big.NewInt(0),
		Principal0: principal0,
		Yield0:     big.NewInt(0),
		Debt1:      auto1,
		Interest1:  big.NewInt(0),
		Fees0:      autoFee0,
		Cost0:      costAuto0,
	}

	log.Info().
		Time("timestamp", timestamp).
		Float64("value", b.ToFloat(input0, 6)).
		Float64("hold", b.ToFloat(hold.Value0(reserve0, reserve1), 6)).
		Float64("uniswap", b.ToFloat(uniswap.Value0(reserve0, reserve1), 6)).
		Float64("autohedge", b.ToFloat(autohedge.Value0(reserve0, reserve1), 6)).
		Msg("position values initialized")

	if writeResults {
		writeHold(timestamp, reserve0, reserve1, hold, outbound)
		writeUniswap(timestamp, reserve0, reserve1, uniswap, outbound)
		writeAutohedge(timestamp, reserve0, reserve1, autohedge, outbound)
	}

	last := timestamp
	for result.Next() {

		record := result.Record()
		timestamp := record.Time()
		values := record.Values()

		reserve0 := b.FromHex(values["reserve0"])
		reserve1 := b.FromHex(values["reserve1"])

		volume0 := b.FromHex(values["volume0"])
		volume1 := b.FromHex(values["volume1"])

		liquidity := big.NewInt(0).Mul(reserve0, reserve1)

		log := log.With().
			Time("timestamp", timestamp).
			Logger()

		log.Debug().
			Float64("reserve0", b.ToFloat(reserve0, 6)).
			Float64("reserve1", b.ToFloat(reserve1, 18)).
			Float64("volume0", b.ToFloat(volume0, 6)).
			Float64("volume1", b.ToFloat(volume1, 18)).
			Float64("liquidity", b.ToFloat(liquidity, 24)).
			Msg("extracted datapoint from record")

		elapsed := big.NewInt(int64(timestamp.Sub(last).Seconds()))

		realLoanRate := util.CalculateCompoundedInterest(loanRate, elapsed)
		yieldDelta0 := big.NewInt(0).Add(autohedge.Principal0, autohedge.Yield0)
		yieldDelta0.Mul(yieldDelta0, realLoanRate)
		yieldDelta0.Div(yieldDelta0, b.E27)
		autohedge.Yield0.Add(autohedge.Yield0, yieldDelta0)

		realBorrowRate := util.CalculateCompoundedInterest(borrowRate, elapsed)
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

		uni0 := big.NewInt(0).Mul(uniswap.Liquidity, reserve0)
		uni0.Div(uni0, reserve1)
		uni0.Sqrt(uni0)

		uni1 := util.Quote(uni0, reserve0, reserve1)

		profitUni0 := big.NewInt(0).Mul(volume0, uniswap.Liquidity)
		profitUni0.Div(profitUni0, liquidity)
		uniswap.Profit0.Add(uniswap.Profit0, profitUni0)
		uni0.Add(uni0, profitUni0)

		profitUni1 := big.NewInt(0).Mul(volume1, uniswap.Liquidity)
		profitUni1.Div(profitUni1, liquidity)
		uniswap.Profit1.Add(uniswap.Profit1, profitUni1)
		uni1.Add(uni1, profitUni1)

		uniswap.Liquidity = big.NewInt(0).Mul(uni0, uni1)

		log.Debug().
			Float64("uni0", b.ToFloat(uni0, 6)).
			Float64("uni1", b.ToFloat(uni1, 18)).
			Float64("profit0", b.ToFloat(profitUni0, 6)).
			Float64("profit1", b.ToFloat(profitUni1, 18)).
			Float64("liquidity", b.ToFloat(uniswap.Liquidity, 24)).
			Msg("added profit to uniswap position")

		auto0 := big.NewInt(0).Mul(autohedge.Liquidity, reserve0)
		auto0.Div(auto0, reserve1)
		auto0.Sqrt(auto0)

		auto1 := util.Quote(auto0, reserve0, reserve1)

		profitAuto0 := big.NewInt(0).Mul(volume0, autohedge.Liquidity)
		profitAuto0.Div(profitAuto0, liquidity)
		autohedge.Profit0.Add(autohedge.Profit0, profitAuto0)
		auto0.Add(auto0, profitAuto0)

		profitAuto1 := big.NewInt(0).Mul(volume1, autohedge.Liquidity)
		profitAuto1.Div(profitAuto1, liquidity)
		autohedge.Profit1.Add(autohedge.Profit1, profitAuto1)
		auto1.Add(auto1, profitAuto1)

		autohedge.Liquidity = big.NewInt(0).Mul(auto0, auto1)

		log.Debug().
			Float64("auto0", b.ToFloat(auto0, 6)).
			Float64("auto1", b.ToFloat(auto1, 18)).
			Float64("profit0", b.ToFloat(profitAuto0, 6)).
			Float64("profit1", b.ToFloat(profitAuto1, 18)).
			Float64("liquidity", b.ToFloat(autohedge.Liquidity, 24)).
			Msg("added profit to autohedge position")

		// position0 := math.Sqrt(autohedge.Liquidity * price)
		// position1 := position0 / price

		// // Uniswap v2 swap:
		// // amountInMultplied := amountIn * 997
		// // numerator := amountInMultiplied * reserveOut
		// // denominator := resorveIn * 1000 + amountInMultiplied
		// // amountOut := numerator / denominator
		// // amountIn - amountOut = fee => solve?

		// switch {

		// case position1 < (autohedge.Debt1+autohedge.Interest1)*(1-rehedgeRatio):

		// 	delta1 := autohedge.Debt1 + autohedge.Interest1 - position1
		// 	out1 := delta1 * (1 + swapRate)
		// 	out0 := out1 * price

		// 	autohedge.Liquidity = (position0 - out0) * (position1 - out1)
		// 	autohedge.Debt1 -= (out0/price + delta1)
		// 	autohedge.Fees0 += out0 * swapRate
		// 	autohedge.Cost0 += (swapGas + decreaseGas + addGas) * gasPrice1 * price

		// 	log.Debug().
		// 		Float64("position0", position0).
		// 		Float64("position1", position1).
		// 		Float64("delta1", delta1).
		// 		Float64("out1", out1).
		// 		Float64("out0", out0).
		// 		Float64("liquidity", autohedge.Liquidity).
		// 		Float64("debt1", autohedge.Debt1).
		// 		Float64("fees0", autohedge.Fees0).
		// 		Float64("cost0", autohedge.Cost0).
		// 		Msg("decreased debt to rehedge autoswap position")

		// case position1 > (autohedge.Debt1+autohedge.Interest1)*(1+rehedgeRatio):

		// 	delta1 := position1 - autohedge.Debt1 - autohedge.Interest1
		// 	in1 := delta1
		// 	in0 := delta1 * price

		// 	autohedge.Liquidity = (position0 + in0) * (position1 + in1)
		// 	autohedge.Debt1 += (in0/price + delta1*(1+swapRate))
		// 	autohedge.Fees0 += in0 * swapRate
		// 	autohedge.Cost0 += (swapGas + increaseGas + removeGas) * gasPrice1 * price

		// 	log.Debug().
		// 		Float64("position0", position0).
		// 		Float64("position1", position1).
		// 		Float64("delta1", delta1).
		// 		Float64("in1", in1).
		// 		Float64("in0", in0).
		// 		Float64("liquidity", autohedge.Liquidity).
		// 		Float64("debt1", autohedge.Debt1).
		// 		Float64("fees0", autohedge.Fees0).
		// 		Float64("cost0", autohedge.Cost0).
		// 		Msg("increased debt to rehedge autoswap position")
		// }

		if writeResults {
			writeHold(timestamp, reserve0, reserve1, hold, outbound)
			writeUniswap(timestamp, reserve0, reserve1, uniswap, outbound)
			writeAutohedge(timestamp, reserve0, reserve1, autohedge, outbound)
		}

		log.Info().
			Float64("hold", b.ToFloat(hold.Value0(reserve0, reserve1), 6)).
			Float64("uniswap", b.ToFloat(uniswap.Value0(reserve0, reserve1), 6)).
			Float64("autohedge", b.ToFloat(autohedge.Value0(reserve0, reserve1), 6)).
			Msg("position values updated")
	}

	err = result.Err()
	if err != nil {
		log.Fatal().Err(err).Msg("could not finish streaming records")
	}

	os.Exit(0)
}
