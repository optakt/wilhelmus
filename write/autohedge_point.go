package write

import (
	"math/big"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/optakt/wilhelmus/b"
	"github.com/optakt/wilhelmus/position"
	"github.com/optakt/wilhelmus/util"
)

func AutohedgePoint(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, autohedge position.Autohedge, outbound api.WriteAPI) {

	number, suffix := humanize.ComputeSI(float64(autohedge.Size))
	size := humanize.Ftoa(number) + suffix

	rehedgeFloat, _ := big.NewFloat(0).SetInt(autohedge.Rehedge).Float64()
	rehedge := humanize.Ftoa(rehedgeFloat/10) + "%"

	interest0 := util.Quote(autohedge.Interest1, reserve1, reserve0)
	debt0 := util.Quote(autohedge.Debt1, reserve1, reserve0)
	debt0.Add(debt0, interest0)
	debt0.Sub(debt0, autohedge.Yield0)

	loss0 := big.NewInt(0).Add(autohedge.Fees0, autohedge.Cost0)
	loss0.Add(loss0, interest0)

	change0 := big.NewInt(0).Sub(autohedge.Profit0, loss0)

	tags := map[string]string{
		"strategy": "autohedge",
		"chain":    "ethereum",
		"size":     size,
		"leverage": "2x",
		"rehedge":  rehedge,
	}
	fields := map[string]interface{}{
		"value":     b.ToFloat(autohedge.Value0(reserve0, reserve1), 6),
		"principal": b.ToFloat(autohedge.Principal0, 6),
		"yield":     b.ToFloat(autohedge.Yield0, 6),
		"debt":      b.ToFloat(debt0, 6),
		"interest":  b.ToFloat(interest0, 6),
		"fees":      b.ToFloat(autohedge.Fees0, 6),
		"cost":      b.ToFloat(autohedge.Cost0, 6),
		"profit":    b.ToFloat(autohedge.Profit0, 6),
		"loss":      b.ToFloat(loss0, 6),
		"change":    b.ToFloat(change0, 6),
	}

	point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	outbound.WritePoint(point)
}
