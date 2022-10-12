package main

import (
	"math/big"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/position"
	"github.com/optakt/wilhelmus/util"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func writeHold(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, hold position.Hold, outbound api.WriteAPI) {

	number, suffix := humanize.ComputeSI(float64(hold.Size))
	size := humanize.Ftoa(number) + suffix

	tags := map[string]string{
		"strategy": "hold",
		"chain":    "ethereum",
		"size":     size,
	}
	fields := map[string]interface{}{
		"value": b.ToFloat(hold.Value0(reserve0, reserve1), 6),
		"fees":  b.ToFloat(hold.Fees0, 6),
		"cost":  b.ToFloat(hold.Cost0, 6),
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}

func writeUniswap(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, uniswap position.Uniswap, outbound api.WriteAPI) {

	number, suffix := humanize.ComputeSI(float64(uniswap.Size))
	size := humanize.Ftoa(number) + suffix

	tags := map[string]string{
		"strategy": "uniswap",
		"chain":    "ethereum",
		"size":     size,
	}
	fields := map[string]interface{}{
		"value":  b.ToFloat(uniswap.Value0(reserve0, reserve1), 6),
		"profit": b.ToFloat(uniswap.Profit0, 6),
		"fees":   b.ToFloat(uniswap.Fees0, 6),
		"cost":   b.ToFloat(uniswap.Cost0, 6),
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}

func writeAutohedge(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, autohedge position.Autohedge, outbound api.WriteAPI) {

	number, suffix := humanize.ComputeSI(float64(autohedge.Size))
	size := humanize.Ftoa(number) + suffix

	rehedgeFloat, _ := big.NewFloat(0).SetInt(autohedge.Rehedge).Float64()
	rehedge := humanize.Ftoa(rehedgeFloat/10) + "%"

	interest0 := util.Quote(autohedge.Interest1, reserve1, reserve0)
	debt0 := util.Quote(autohedge.Debt1, reserve1, reserve0)
	debt0.Add(debt0, interest0)
	debt0.Sub(debt0, autohedge.Yield0)

	tags := map[string]string{
		"strategy": "autohedge",
		"chain":    "ethereum",
		"size":     size,
		"leverage": "2x",
		"rehedge":  rehedge,
	}
	fields := map[string]interface{}{
		"value":  b.ToFloat(autohedge.Value0(reserve0, reserve1), 6),
		"profit": b.ToFloat(autohedge.Profit0, 6),
		"debt":   b.ToFloat(debt0, 6),
		"fees":   b.ToFloat(autohedge.Fees0, 6),
		"cost":   b.ToFloat(autohedge.Cost0, 6),
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}
