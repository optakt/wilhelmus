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

	record := result.Record()
	timestamp := record.Time()
	values := record.Values()

	reserve0 := values["reserve0"].(float64)
	reserve1 := values["reserve1"].(float64)
	price := reserve0 / reserve1

	gasPrice1, err := station.Gasprice(timestamp)
	if err != nil {
		log.Fatal().Err(err).Time("timestamp", timestamp).Msg("could not get gas price for timestamp")
	}

	input0 := inputValue * d6

	swapFee0 := input0 / 2 * swapRate
	hold0 := (input0 - swapFee0) / 2
	hold1 := hold0 / price
	gasHold := approveGas + swapGas

	hold := position.Hold{
		Amount0: hold0,
		Amount1: hold1,
		Fees0:   swapFee0,
		Cost0:   gasHold * gasPrice1 * price,
	}

	liquidity := hold0 * hold1
	gasUni := gasHold + createGas

	uniswap := position.Uniswap{
		Liquidity: liquidity,
		Fees0:     hold.Fees0,
		Cost0:     gasUni * gasPrice1 * price,
	}

	auto1 := input0 / (price * (1 + (flashRate / (1 - swapRate))))
	fee1 := auto1 * flashRate
	fee0 := fee1 * price / (1 - swapRate)
	auto0 := input0 - fee0
	gasAuto := gasUni + (2*approveGas + flashGas + lendGas + borrowGas)

	autohedge := position.Autohedge{
		Liquidity:  auto0 * auto1,
		Principal0: auto0 + auto1*price,
		Yield0:     0,
		Debt1:      auto1,
		Interest1:  0,
		Fees0:      fee0 + fee1*price,
		Cost0:      gasAuto * gasPrice1 * price,
	}

	writeHold(timestamp, price, hold, outbound)
	writeUniswap(timestamp, price, uniswap, outbound)
	writeAutohedge(timestamp, price, autohedge, outbound)

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
		yieldDelta0 := realLoanRate * (autohedge.Principal0 + autohedge.Yield0)
		autohedge.Yield0 += yieldDelta0

		realBorrowRate := util.CompoundRate(borrowRate, uint(elapsed))
		interestDelta1 := realBorrowRate * (autohedge.Debt1 + autohedge.Interest1)
		autohedge.Interest1 += interestDelta1

		last = timestamp

		log.Debug().
			Float64("principal0", autohedge.Principal0).
			Float64("yield0", autohedge.Yield0).
			Float64("gain0", yieldDelta0).
			Float64("debt1", autohedge.Debt1).
			Float64("interest1", autohedge.Interest1).
			Float64("loss1", interestDelta1).
			Msg("compounded principal yield and debt interest")

		liquidity := reserve0 * reserve1

		log.Debug().
			Float64("liquidity", liquidity).
			Float64("uniswap", uniswap.Liquidity).
			Float64("autohedge", autohedge.Liquidity).
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

		switch {

		case position1 < (autohedge.Debt1+autohedge.Interest1)*(1-rehedgeRatio):

			delta1 := autohedge.Debt1 + autohedge.Interest1 - position1
			out1 := delta1 * (1 + swapRate)
			out0 := out1 * price

			autohedge.Liquidity = (position0 - out0) * (position1 - out1)
			autohedge.Debt1 -= (out0/price + delta1)
			autohedge.Fees0 += out0 * swapRate
			autohedge.Cost0 += (swapGas + unborrowGas + addGas) * gasPrice1 * price

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
			autohedge.Cost0 += (swapGas + reborrowGas + removeGas) * gasPrice1 * price

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

		writeHold(timestamp, price, hold, outbound)
		writeUniswap(timestamp, price, uniswap, outbound)
		writeAutohedge(timestamp, price, autohedge, outbound)

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
