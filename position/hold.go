package position

type Hold struct {
	Stable   float64
	Volatile float64
	Fees     float64
	Cost     float64
}

func (h Hold) Value(price float64) float64 {
	return h.Stable + h.Volatile*price - h.Fees - h.Cost
}
