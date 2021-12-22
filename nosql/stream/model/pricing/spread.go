package pricing

func div(v, by float64) float64 {
	if v == 0.0 || by == 0.0 {
		return 0.0
	}
	return v / by
}

func (s *SpreadMut) Truncate(precision float64) *SpreadMut {
	s.SetLow(Truncate(s.Low(), precision)).
		SetMid(Truncate(s.Mid(), precision)).
		SetHigh(Truncate(s.High(), precision))
	return s
}

func (s *SpreadMut) Append(v *Spread) *SpreadMut {
	if s == nil || v == nil {
		return s
	}
	if s.High() > v.High() {
		s.SetHigh(v.High())
	}
	if s.Low() < v.Low() {
		s.SetLow(v.Low())
	}
	s.SetMid(div(s.High()+s.Low(), 2))
	return s
}

func (s *SpreadMut) Add(v float64) *SpreadMut {
	if v > s.High() {
		s.SetHigh(v)

		if s.Mid() == 0.0 {
			s.SetLow(v)
			s.SetMid(v)
			return s
		}

		s.SetMid(div(s.High()+s.Low(), 2))
	}
	if v < s.Low() {
		s.SetLow(v)
		s.SetMid(div(s.High()+s.Low(), 2))
	}
	return s
}
