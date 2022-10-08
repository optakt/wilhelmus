package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"github.com/optakt/dewalt/station"
)

func main() {

	var (
		input float64
		start string
		end   string

		logLevel     string
		gasPrices    string
		influxAPI    string
		influxToken  string
		influxOrg    string
		influxBucket string

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

	pflag.StringVarP(&logLevel, "log-level", "l", "info", "Zerolog logger logging message severity")
	pflag.StringVarP(&gasPrices, "gas-prices", "g", "gas-prices.csv", "CSV file for average gas price per day")
	pflag.StringVarP(&influxAPI, "influx-api", "i", "https://eu-central-1-1.aws.cloud2.influxdata.com", "InfluxDB API URL")
	pflag.StringVarP(&influxToken, "influx-token", "t", "3Lq2o0e6-NmfpXK_UQbPqknKgQUbALMdNz86Ojhpm6dXGqGnCuEYGZijTMGhP82uxLfoWiWZRS2Vls0n4dZAjQ==", "InfluxDB authentication token")
	pflag.StringVarP(&influxOrg, "influx-org", "o", "optakt", "InfluxDB organization name")
	pflag.StringVarP(&influxBucket, "influx-bucket", "u", "uniswap", "InfluxDB bucket name")

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

	_ = station

	os.Exit(0)
}
