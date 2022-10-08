package position

import (
	"time"
)

type Station interface {
	Gasprice(timestamp time.Time) (float64, error)
}
