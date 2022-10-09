package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/optakt/dewalt/position"
	"github.com/optakt/dewalt/station"
	"github.com/optakt/dewalt/util"
)

const (
	d6  = float64(1_000_000)
	d18 = float64(1_000_000_000_000_000_000)
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
		logLevel  string
		gasPrices string

		inputValue   float64
		startTime    string
		endTime      string
		rehedgeRatio float64

		influxAPI             string
		influxToken           string
		influxOrg             string
		influxTimeout         time.Duration
		influxBucketUniswap   string
		influxBucketPositions string

		swapRate   float64
		flashRate  float64
		loanRate   float64
		borrowRate float64

		approveGas float64
		swapGas    float64
		flashGas   float64

		createGas float64
		addGas    float64
		removeGas float64
		closeGas  float64

		lendGas  float64
		claimGas float64

		borrowGas   float64
		reborrowGas float64
		unborrowGas float64
		repayGas    float64
	)

	pflag.StringVarP(&logLevel, "log-level", "l", "info", "Zerolog logger logging message severity")
	pflag.StringVarP(&gasPrices, "gas-prices", "g", "gas-prices.csv", "CSV file for average gas price per day")

	pflag.Float64VarP(&inputValue, "input-value", "i", 100_000, "stable coin input amount")
	pflag.StringVarP(&startTime, "start-time", "s", "2021-10-07T00:00:00Z", "start timestamp for the backtest")
	pflag.StringVarP(&endTime, "end-time", "e", "2022-10-07T23:59:59Z", "end timestamp for the backtest")
	pflag.Float64Var(&rehedgeRatio, "rehedge-ratio", 0.01, "ratio between debt and collateral at which we rehedge")

	pflag.StringVarP(&influxAPI, "influx-api", "a", "https://eu-central-1-1.aws.cloud2.influxdata.com", "InfluxDB API URL")
	pflag.StringVarP(&influxToken, "influx-token", "t", "3Lq2o0e6-NmfpXK_UQbPqknKgQUbALMdNz86Ojhpm6dXGqGnCuEYGZijTMGhP82uxLfoWiWZRS2Vls0n4dZAjQ==", "InfluxDB authentication token")
	pflag.StringVarP(&influxOrg, "influx-org", "o", "optakt", "InfluxDB organization name")
	pflag.DurationVarP(&influxTimeout, "influx-timeout", "u", 15*time.Minute, "InfluxDB query HTTP request timeout")
	pflag.StringVar(&influxBucketUniswap, "influx-bucket-uniswap", "uniswap", "InfluxDB bucket name for Uniswap metrics")
	pflag.StringVar(&influxBucketPositions, "influx-bucket-positions", "positions", "InfluxDB bucket for position values")

	pflag.Float64Var(&swapRate, "swap-rate", 0.003, "fee rate for asset swap")
	pflag.Float64Var(&flashRate, "flash-rate", 0.0009, "fee rate for flash loan")
	pflag.Float64Var(&loanRate, "lend-rate", 0.005, "interest rate for lending asset")
	pflag.Float64Var(&borrowRate, "borrow-rate", 0.025, "interest rate for borrowing asset")

	pflag.Float64Var(&approveGas, "approve-gas", 24102, "gas cost for transfer approval")
	pflag.Float64Var(&swapGas, "swap-gas", 181133, "gas cost for asset swap")
	pflag.Float64Var(&flashGas, "flash-gas", 204493, "gas cost for flash loan")

	pflag.Float64Var(&createGas, "provide-gas", 157880, "gas cost for creating liquidity position")
	pflag.Float64Var(&addGas, "add-gas", 130682, "gas cost for adding liquidity")
	pflag.Float64Var(&removeGas, "remove-gas", 161841, "gas cost to remove liquidity")
	pflag.Float64Var(&closeGas, "close-gas", 207111, "gas cost for close liquidity position")

	pflag.Float64Var(&lendGas, "lend-gas", 217479, "gas cost for lending asset")
	pflag.Float64Var(&claimGas, "claim-gas", 333793, "gas cost to claim back loan")

	pflag.Float64Var(&borrowGas, "borrow-gas", 295250, "gas cost for borrowing asset")
	pflag.Float64Var(&unborrowGas, "unborrow-gas", 193729, "gas cost for reducing debt")
	pflag.Float64Var(&reborrowGas, "increase-gas", 271980, "gas cost for increasing debt")
	pflag.Float64Var(&repayGas, "repay-gas", 188929, "gas cost to repay full debt")

	pflag.Parse()

	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
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
	influx := client.QueryAPI(influxOrg)
	query := fmt.Sprintf(statement, startTime, endTime)
	result, err := influx.Query(context.Background(), query)
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

	record := result.Record()
	timestamp := record.Time()
	values := record.Values()

	reserve0 := values["reserve0"].(float64)
	reserve1 := values["reserve1"].(float64)
	price := reserve0 / reserve1

	// volume0 := values["volume0"].(float64) // won't make a profit until...
	// volume1 := values["volume1"].(float64) // ...the positions are created

	gasPrice, err := station.Gasprice(timestamp)
	if err != nil {
		log.Fatal().Err(err).Time("timestamp", timestamp).Msg("could not get gas price for timestamp")
	}

	input0 := inputValue * d6

	hold := position.Hold{
		Amount0: input0 / 2 * (1 - swapRate/2),
		Amount1: input0 / 2 / price * (1 - swapRate/2),
		Fees0:   input0 / 2 * swapRate,
		Cost0:   (approveGas + swapGas) * gasPrice * price,
	}

	uniswap := position.Uniswap{
		Liquidity: hold.Amount0 * hold.Amount1,
		Fees0:     hold.Fees0,
		Cost0:     hold.Cost0 + createGas*gasPrice*price,
	}

	remaining1 := input0 / (price * (1 + (flashRate / (1 - swapRate))))
	fee1 := remaining1 * flashRate
	fee0 := flashRate * price / (1 - swapRate)
	remaining0 := input0 - fee0

	autohedge := position.Autohedge{
		Liquidity:  remaining0 * remaining1,
		Principal0: remaining0 + remaining1*price,
		Yield0:     0,
		Debt1:      remaining1,
		Interest1:  0,
		Fees0:      fee0 + fee1*price,
		Cost0:      uniswap.Cost0 + (2*approveGas+flashGas+lendGas+borrowGas)*gasPrice*price,
	}

	last := timestamp
	for result.Next() {

		record := result.Record()
		timestamp := record.Time()
		values := record.Values()

		reserve0 := values["reserve0"].(float64)
		reserve1 := values["reserve1"].(float64)
		price := reserve0 / reserve1

		volume0 := values["volume0"].(float64)
		volume1 := values["volume1"].(float64)

		log.Debug().
			Time("timestamp", timestamp).
			Float64("reserve0", reserve0).
			Float64("reserve1", reserve1).
			Float64("volume0", volume0).
			Float64("volume1", volume1).
			Float64("price", price).
			Msg("extracted datapoint from record")

		elapsed := timestamp.Sub(last).Seconds()

		realLoanRate := util.CompoundRate(loanRate, uint(elapsed))
		loanYield := realLoanRate * (autohedge.Principal0 + autohedge.Yield0)
		autohedge.Yield0 += loanYield

		realBorrowRate := util.CompoundRate(borrowRate, uint(elapsed))
		borrowInterest := realBorrowRate * (autohedge.Debt1 + autohedge.Interest1)
		autohedge.Interest1 += borrowInterest

		last = timestamp

		log.Debug().
			Float64("principal", autohedge.Principal0).
			Float64("yield", autohedge.Yield0).
			Float64("debt", autohedge.Debt1).
			Float64("interest", autohedge.Interest1).
			Msg("compounded principal yield and debt interest")

		liquidity := reserve0 * reserve1

		log.Debug().
			Float64("uniswap", uniswap.Liquidity).
			Float64("autohedge", autohedge.Liquidity).
			Float64("liquidity", liquidity).
			Msg("preparing to calculate uniswap returns")

		shareUni := uniswap.Liquidity / liquidity
		profitUni0 := shareUni * volume0
		profitUni1 := shareUni * volume1
		uni0 := math.Sqrt(uniswap.Liquidity * price)
		uni1 := uni0 / price
		uniswap.Liquidity = (uni0 + profitUni0) * (uni1 + profitUni1)

		log.Debug().
			Float64("share", shareUni).
			Float64("profit0", profitUni0).
			Float64("profit1", profitUni1).
			Float64("liquidity", uniswap.Liquidity).
			Msg("added profit to uniswap position")

		shareAuto := autohedge.Liquidity / liquidity
		profitAuto0 := shareAuto * volume0
		profitAuto1 := shareAuto * volume1
		auto0 := math.Sqrt(autohedge.Liquidity * price)
		auto1 := auto0 / price
		autohedge.Liquidity = (auto0 + profitAuto0) * (auto1 + profitAuto1)

		log.Debug().
			Float64("share", shareAuto).
			Float64("profit0", profitAuto0).
			Float64("profit1", profitAuto1).
			Float64("liquidity", autohedge.Liquidity).
			Msg("added profit to autohedge position")

		amount0 := math.Sqrt(uniswap.Liquidity * price)
		amount1 := uni0 / price

		switch {

		case amount1 < (autohedge.Debt1+autohedge.Interest1)*(1-rehedgeRatio):
			delta1 := autohedge.Debt1 + autohedge.Interest1 - amount1
			move1 := delta1 * (1 + swapRate)
			move0 := move1 * price
			autohedge.Liquidity = (amount0 - move0) * (amount1 - move1)
			autohedge.Debt1 -= (delta1 + move1)
			autohedge.Fees0 += move0 * swapRate
			autohedge.Cost0 += (swapGas + unborrowGas + addGas) * gasPrice / price

		case amount1 > (autohedge.Debt1+autohedge.Interest1)*(1+rehedgeRatio):
			delta1 := amount1 - autohedge.Debt1 - autohedge.Interest1
			move1 := delta1
			move0 := delta1 * price
			autohedge.Liquidity = (amount0 + move0) * (amount1 + move1)
			autohedge.Debt1 += ((2 + swapRate) * delta1)
			autohedge.Fees0 += move0 * swapRate
			autohedge.Cost0 += (swapGas + reborrowGas + removeGas) * gasPrice / price
		}

		log.Info().
			Float64("price", price*d18/d6).
			Float64("hold", hold.Value0(price)/d6).
			Float64("uniswap", uniswap.Value0(price)/d6).
			Float64("autohedge", autohedge.Value0(price)/d6).
			Msg("updated position valuations")
	}

	err = result.Err()
	if err != nil {
		log.Fatal().Err(err).Msg("could not finish streaming records")
	}

	os.Exit(0)
}
