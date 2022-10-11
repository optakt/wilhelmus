package main

import (
	"math/big"
	"time"

	"github.com/optakt/wilhelmus/position"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

const (
	e6  = 1_000_000
	e18 = 1_000_000_000_000_000_000
)

func writeHold(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, hold position.Hold, outbound api.WriteAPI) {

	// sizeFloat, _ := big.NewFloat(0).SetInt(hold.Size).Float64()
	// number, suffix := humanize.ComputeSI(sizeFloat)
	// size := humanize.Ftoa(number) + suffix

	// value, _ := big.NewFloat(0).SetInt(hold.Value0(price)).Float64()
	// fees, _ := big.NewFloat(0).SetInt(hold.Fees0).Float64()
	// cost, _ := big.NewFloat(0).SetInt(hold.Cost0).Float64()

	// tags := map[string]string{
	// 	"strategy": "hold",
	// 	"chain":    "ethereum",
	// 	"size":     size,
	// }
	// fields := map[string]interface{}{
	// 	"value": value / e6,
	// 	"fees":  fees / e6,
	// 	"cost":  cost / e6,
	// }

	// point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	// outbound.WritePoint(point)
}

func writeUniswap(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, uniswap position.Uniswap, outbound api.WriteAPI) {

	// sizeFloat, _ := big.NewFloat(0).SetInt(uniswap.Size).Float64()
	// number, suffix := humanize.ComputeSI(sizeFloat)
	// size := humanize.Ftoa(number) + suffix

	// priceFloat, _ := big.NewFloat(0).SetInt(price).Float64()
	// priceFloat /= e18

	// profit0, _ := big.NewFloat(0).SetInt(uniswap.Profit0).Float64()
	// profit1, _ := big.NewFloat(0).SetInt(uniswap.Profit1).Float64()
	// profit := profit0 * profit1 * priceFloat

	// value, _ := big.NewFloat(0).SetInt(uniswap.Value0(price)).Float64()
	// fees, _ := big.NewFloat(0).SetInt(uniswap.Fees0).Float64()
	// cost, _ := big.NewFloat(0).SetInt(uniswap.Cost0).Float64()

	// tags := map[string]string{
	// 	"strategy": "uniswap",
	// 	"chain":    "ethereum",
	// 	"size":     size,
	// }
	// fields := map[string]interface{}{
	// 	"value":  value / e6,
	// 	"profit": profit / e6,
	// 	"fees":   fees / e6,
	// 	"cost":   cost / e6,
	// }

	// point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	// outbound.WritePoint(point)
}

func writeAutohedge(timestamp time.Time, reserve0 *big.Int, reserve1 *big.Int, autohedge position.Autohedge, outbound api.WriteAPI) {

	// sizeFloat, _ := big.NewFloat(0).SetInt(autohedge.Size).Float64()
	// number, suffix := humanize.ComputeSI(sizeFloat)
	// size := humanize.Ftoa(number) + suffix

	// rehedgeFloat, _ := big.NewFloat(0).SetInt(autohedge.Rehedge).Float64()
	// rehedge := humanize.Ftoa(rehedgeFloat*100) + "%"

	// priceFloat, _ := big.NewFloat(0).SetInt(price).Float64()
	// priceFloat /= e18

	// profit0, _ := big.NewFloat(0).SetInt(autohedge.Profit0).Float64()
	// profit1, _ := big.NewFloat(0).SetInt(autohedge.Profit1).Float64()
	// profit := profit0 * profit1 * priceFloat

	// value, _ := big.NewFloat(0).SetInt(autohedge.Value0(price)).Float64()
	// fees, _ := big.NewFloat(0).SetInt(autohedge.Fees0).Float64()
	// cost, _ := big.NewFloat(0).SetInt(autohedge.Cost0).Float64()
	// yield, _ := big.NewFloat(0).SetInt(autohedge.Yield0).Float64()

	// debt, _ := big.NewFloat(0).SetInt(autohedge.Debt1).Float64()
	// debt *= priceFloat

	// interest, _ := big.NewFloat(0).SetInt(autohedge.Interest1).Float64()
	// interest *= priceFloat

	// tags := map[string]string{
	// 	"strategy": "autohedge",
	// 	"chain":    "ethereum",
	// 	"size":     size,
	// 	"leverage": "2x",
	// 	"rehedge":  rehedge,
	// }
	// fields := map[string]interface{}{
	// 	"value":    value / e6,
	// 	"profit":   profit / e6,
	// 	"fees":     fees / e6,
	// 	"cost":     cost / e6,
	// 	"yield":    yield / e6,
	// 	"debt":     debt / e6,
	// 	"interest": interest / e6,
	// }

	// point := write.NewPoint("uniswapv2", tags, fields, timestamp)
	// outbound.WritePoint(point)
}
