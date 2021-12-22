package pricing

func (v *VolumeMut) Finish() {
	v.Buy().SetPercent(div(v.Buy().Total(), v.Total()))
	v.Sell().SetPercent(div(v.Sell().Total(), v.Total()))
}
