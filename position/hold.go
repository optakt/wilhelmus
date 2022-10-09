package position

type Hold struct {
	Size    uint64
	Amount0 float64
	Amount1 float64
	Fees0   float64
	Cost0   float64
}

func (h Hold) Value0(price float64) float64 {
	return h.Amount0 + h.Amount1*price - h.Fees0 - h.Cost0
}
