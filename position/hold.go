package position

type Hold struct {
	stable   float64
	volatile float64
	debt     float64
}

func NewHold(input float64, price float64) *Hold {
	return &Hold{
		stable:   input,
		volatile: input / price,
		debt:     input,
	}
}

func (h Hold) Value(price float64) float64 {
	return h.stable + h.volatile*price - h.debt
}
