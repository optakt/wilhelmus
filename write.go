package main

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/optakt/wilhelmus/position"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func writeHold(timestamp time.Time, price float64, hold position.Hold, outbound api.WriteAPI) {

	value, prefix := humanize.ComputeSI(float64(hold.Size))
	size := humanize.Ftoa(value) + prefix

	tags := map[string]string{
		"strategy": "hold",
		"chain":    "ethereum",
		"size":     size,
	}
	fields := map[string]interface{}{
		"value": hold.Value0(price) / d6,
		"fees":  hold.Fees0 / d6,
		"cost":  hold.Cost0 / d6,
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}

func writeUniswap(timestamp time.Time, price float64, uniswap position.Uniswap, outbound api.WriteAPI) {

	value, prefix := humanize.ComputeSI(float64(uniswap.Size))
	size := humanize.Ftoa(value) + prefix

	tags := map[string]string{
		"strategy": "uniswap",
		"chain":    "ethereum",
		"size":     size,
	}
	fields := map[string]interface{}{
		"value":  uniswap.Value0(price) / d6,
		"profit": (uniswap.Profit0 + uniswap.Profit1*price) / d6,
		"fees":   uniswap.Fees0 / d6,
		"cost":   uniswap.Cost0 / d6,
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}

func writeAutohedge(timestamp time.Time, price float64, autohedge position.Autohedge, outbound api.WriteAPI) {

	value, prefix := humanize.ComputeSI(float64(autohedge.Size))
	size := humanize.Ftoa(value) + prefix

	rehedge := humanize.Ftoa(autohedge.Rehedge*100) + "%"

	tags := map[string]string{
		"strategy": "autohedge",
		"chain":    "ethereum",
		"size":     size,
		"leverage": "2x",
		"rehedge":  rehedge,
	}
	fields := map[string]interface{}{
		"value":    autohedge.Value0(price) / d6,
		"profit":   (autohedge.Profit0 + autohedge.Profit1*price) / d6,
		"fees":     autohedge.Fees0 / d6,
		"cost":     autohedge.Cost0 / d6,
		"yield":    autohedge.Yield0 / d6,
		"debt":     autohedge.Debt1 * price / d6,
		"interest": autohedge.Interest1 * price / d6,
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}
