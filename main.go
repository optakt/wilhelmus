package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"github.com/optakt/dewalt/position"
	"github.com/optakt/dewalt/station"
)

func main() {

	var (
		inputValue   float64
		startTime    string
		endTime      string
		rehedgeRatio float64

		logLevel              string
		gasPrices             string
		influxAPI             string
		influxToken           string
		influxOrg             string
		influxBucketUniswap   string
		influxBucketPositions string

		swapFee        float64
		flashFee       float64
		lendInterest   float64
		borrowInterest float64

		approveGas  float64
		swapGas     float64
		provideGas  float64
		flashGas    float64
		lendGas     float64
		claimGas    float64
		borrowGas   float64
		reborrowGas float64
		unborrowGas float64
		repayGas    float64
	)

	pflag.Float64VarP(&inputValue, "input-value", "i", 100_000, "stable coin input amount")
	pflag.StringVarP(&startTime, "start-time", "s", "2021-05-09T00:00:00Z", "start timestamp for the backtest")
	pflag.StringVarP(&endTime, "end-time", "e", "2022-10-08T00:00:00Z", "end timestamp for the backtest")
	pflag.Float64Var(&rehedgeRatio, "rehedge-ratio", 0.01, "ratio between debt and collateral at which we rehedge")

	pflag.StringVarP(&logLevel, "log-level", "l", "info", "Zerolog logger logging message severity")
	pflag.StringVarP(&gasPrices, "gas-prices", "g", "gas-prices.csv", "CSV file for average gas price per day")
	pflag.StringVarP(&influxAPI, "influx-api", "i", "https://eu-central-1-1.aws.cloud2.influxdata.com", "InfluxDB API URL")
	pflag.StringVarP(&influxToken, "influx-token", "t", "3Lq2o0e6-NmfpXK_UQbPqknKgQUbALMdNz86Ojhpm6dXGqGnCuEYGZijTMGhP82uxLfoWiWZRS2Vls0n4dZAjQ==", "InfluxDB authentication token")
	pflag.StringVarP(&influxOrg, "influx-org", "o", "optakt", "InfluxDB organization name")
	pflag.StringVarP(&influxBucketUniswap, "influx-bucket-uniswap", "u", "uniswap", "InfluxDB bucket name for Uniswap metrics")
	pflag.StringVarP(&influxBucketPositions, "influx-bucket-positions", "p", "positions", "InfluxDB bucket for position values")

	pflag.Float64Var(&swapFee, "swap-fee", 0.003, "fee rate for asset swap")
	pflag.Float64Var(&flashFee, "flash-fee", 0.0009, "fee rate for flash loan")
	pflag.Float64Var(&lendInterest, "lend-interest", 0.005, "interest rate for lending asset")
	pflag.Float64Var(&borrowInterest, "borrow-interest", 0.025, "interest rate for borrowing asset")

	pflag.Float64Var(&approveGas, "approve-gas", 24102, "gas cost for transfer approval")
	pflag.Float64Var(&swapGas, "swap-gas", 181133, "gas cost for asset swap")
	pflag.Float64Var(&provideGas, "provide-gas", 0, "gas cost for providing liquidity")
	pflag.Float64Var(&flashGas, "flash-gas", 204493, "gas cost for flash loan")
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

	startTimestamp, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		log.Fatal().Err(err).Str("start_time", startTime).Msg("invalid start time")
	}

	endTimestamp, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		log.Fatal().Err(err).Str("end_time", endTime).Msg("invaldi end time")
	}

	station, err := station.New(gasPrices)
	if err != nil {
		log.Fatal().Err(err).Str("gas_prices", gasPrices).Msg("could not create gas station")
	}

	for timestamp := startTimestamp; timestamp == timestamp; _ = endTimestamp {

		// get price series from influx here
		price := float64(0)

		gasPrice, err := station.Gasprice(timestamp)
		if err != nil {
			log.Fatal().Err(err).Time("timestamp", timestamp).Msg("could not get gas price for timestamp")
		}

		hold := position.Hold{
			Stable:   inputValue / 2 * (1 - swapFee/2),
			Volatile: inputValue / 2 / price * (1 - swapFee/2),
			Fees:     inputValue / 2 * swapFee,
			Cost:     (2*approveGas + swapGas) * gasPrice,
		}

		uniswap := position.Uniswap{
			Liquidity: hold.Stable * hold.Volatile,
			Fees:      hold.Fees,
			Cost:      hold.Cost + (2*approveGas+swapGas+provideGas)*gasPrice,
		}

		amountVol := inputValue / (price * (1 + (flashFee / (1 - swapFee))))
		feeVol := amountVol * flashFee
		feeStable := flashFee * price / (1 - swapFee)
		amountStable := inputValue - feeStable

		autohedge := position.Autohedge{
			Ratio:     rehedgeRatio,
			Liquidity: (amountStable * amountVol),
			Debt:      amountVol,
			Fees:      feeVol*price + feeStable,
			Cost:      uniswap.Cost + (4*approveGas+flashGas+lendGas+borrowGas)*gasPrice,
			Interest:  0,
		}

		_ = autohedge
	}

	os.Exit(0)
}
