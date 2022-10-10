package position

import (
	"math/big"
	"time"
)

type Station interface {
	Gasprice(timestamp time.Time) (*big.Int, error)
}
