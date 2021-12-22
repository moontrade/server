package pricing

import (
	"math"
)

func NewCandle(open, high, low, close float64) *Candle {
	return (&CandleMut{}).SetOpen(open).SetHigh(high).SetLow(low).SetClose(close).Freeze()
}

func (c *CandleMut) Truncate(precision float64) {
	c.SetOpen(Truncate(c.Open(), precision)).
		SetHigh(Truncate(c.High(), precision)).
		SetLow(Truncate(c.Low(), precision)).
		SetClose(Truncate(c.Close(), precision))
}

func (c *CandleMut) AddPrice(price float64) {
	if c.Open() == 0 {
		c.SetOpen(price).SetHigh(price).SetLow(price).SetClose(price)
	} else {
		c.SetClose(price).
			SetLow(math.Min(c.Low(), price)).
			SetHigh(math.Max(c.High(), price))
	}
}

func (c *CandleMut) Append(o *Candle) {
	if o == nil {
		return
	}

	if c.Open() == 0 {
		c.SetOpen(o.Open()).SetHigh(o.High()).SetLow(o.Low()).SetClose(o.Close())
		return
	}
	c.SetClose(o.Close()).
		SetHigh(math.Max(c.High(), o.High())).
		SetLow(math.Min(c.Low(), o.Low()))
}
