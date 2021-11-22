package nosql

type Float64ArrayValueOf func(data string, unmarshalled interface{}, into []float64) ([]float64, error)

type Float64Array struct {
	indexBase
	ValueOf Float64ArrayValueOf
}

func NewFloat64Array(
	name, selector, version string,
	valueOf Float64ArrayValueOf,
) *Float64Array {
	if valueOf == nil {
		valueOf = jsonFloat64Array(selector)
	}
	return &Float64Array{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindFloat64, false, true),
	}
}
