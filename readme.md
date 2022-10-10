# Wilhelmus

## Description

Wilhelmus is a Go command line tool for backtesting DeFi investment strategies.

In particular, the tool currently implements backtesting for hold positions, for Uniswap v2 liquidity positions, and for Autonomy Network powered AutoHedge positions.

## Installation

Simply clone the repository and build the Go binary.

```
git clone https://github.com/optakt/wilhelmus.git
go build
```

## Usage

The binary can output the following help message using `./wilhelmus --help`:

```
Usage of ./wilhelmus:
      --add-gas float                    gas cost for adding liquidity (default 130682)
      --approve-gas float                gas cost for transfer approval (default 24102)
      --borrow-gas float                 gas cost for borrowing asset (default 295250)
      --borrow-rate float                interest rate for borrowing asset (default 0.025)
      --claim-gas float                  gas cost to claim back loan (default 333793)
      --close-gas float                  gas cost for close liquidity position (default 207111)
  -e, --end-time string                  end timestamp for the backtest (default "2022-10-07T23:59:59Z")
      --flash-gas float                  gas cost for flash loan (default 204493)
      --flash-rate float                 fee rate for flash loan (default 0.0009)
  -g, --gas-prices string                CSV file for average gas price per day (default "gas-prices.csv")
      --increase-gas float               gas cost for increasing debt (default 271980)
  -a, --influx-api string                InfluxDB API URL (default "https://eu-central-1-1.aws.cloud2.influxdata.com")
      --influx-bucket-positions string   InfluxDB bucket for position values (default "positions")
      --influx-bucket-uniswap string     InfluxDB bucket name for Uniswap metrics (default "uniswap")
  -o, --influx-org string                InfluxDB organization name (default "optakt")
  -u, --influx-timeout duration          InfluxDB query HTTP request timeout (default 15m0s)
  -t, --influx-token string              InfluxDB authentication token (default "3Lq2o0e6-NmfpXK_UQbPqknKgQUbALMdNz86Ojhpm6dXGqGnCuEYGZijTMGhP82uxLfoWiWZRS2Vls0n4dZAjQ==")
  -i, --input-value uint                 stable coin input amount (default 1000000)
      --lend-gas float                   gas cost for lending asset (default 217479)
      --lend-rate float                  interest rate for lending asset (default 0.005)
  -l, --log-level string                 Zerolog logger logging message severity (default "info")
      --provide-gas float                gas cost for creating liquidity position (default 157880)
  -r, --rehedge-ratio float              ratio between debt and collateral at which we rehedge (default 0.01)
      --remove-gas float                 gas cost to remove liquidity (default 161841)
      --repay-gas float                  gas cost to repay full debt (default 188929)
  -s, --start-time string                start timestamp for the backtest (default "2021-10-07T00:00:00Z")
      --swap-gas float                   gas cost for asset swap (default 181133)
      --swap-rate float                  fee rate for asset swap (default 0.003)
      --unborrow-gas float               gas cost for reducing debt (default 193729)
  -w, --write-results                    whether to write the results back to InfluxDB
```

## Metrics

The tool relies on Uniswap v2 metrics from a InfluxDB bucket.

In order to collect the necessary metrics on a Uniswap v2 tool, you can use Klangbaach, the companion tool:

https://github.com/optakt/klangbaach