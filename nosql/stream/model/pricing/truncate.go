package pricing

import "math"

func Truncate(f float64, unit float64) float64 {
	//return math.Trunc(math.Round(f*unit)) / unit
	v := math.Trunc(f * unit)
	if v == 0.0 {
		return 0.0
	}
	return v / unit
}
