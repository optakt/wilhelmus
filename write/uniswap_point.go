package write

import (
	"math/big"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/position"
)

func UniswapPoint(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, uniswap position.Uniswap, outbound api.WriteAPI) {

	number, suffix := humanize.ComputeSI(float64(uniswap.Size))
	size := humanize.Ftoa(number) + suffix

	loss0 := big.NewInt(0).Add(uniswap.Fees0, uniswap.Cost0)

	change0 := big.NewInt(0).Sub(uniswap.Profit0, loss0)

	tags := map[string]string{
		"strategy": "uniswap",
		"chain":    "ethereum",
		"size":     size,
	}
	fields := map[string]interface{}{
		"value":  b.ToFloat(uniswap.Value0(reserve0, reserve1), 6),
		"fees":   b.ToFloat(uniswap.Fees0, 6),
		"cost":   b.ToFloat(uniswap.Cost0, 6),
		"profit": b.ToFloat(uniswap.Profit0, 6),
		"loss":   b.ToFloat(loss0, 6),
		"change": b.ToFloat(change0, 6),
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}
