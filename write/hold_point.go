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

func HoldPoint(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, hold position.Hold, outbound api.WriteAPI) {

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
