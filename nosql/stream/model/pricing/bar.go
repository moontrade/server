package pricing

func NewBar(o, h, l, c float64) *Bar {
	b := &BarMut{}
	b.Price().SetOpen(o).SetHigh(h).SetLow(l).SetClose(c)
	return b.Freeze()
}

func (b *BarMut) Truncate() *BarMut {
	p := b.Precision()
	b.Price().Truncate(p)
	b.Ask().Truncate(p)
	b.Bid().Truncate(p)
	b.Spread().Truncate(p)
	return b
}
