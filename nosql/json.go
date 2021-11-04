package nosql

import (
	"github.com/tidwall/gjson"
)

func jsonFloat64(selector string) func(doc string) (float64, error) {
	return func(doc string) (float64, error) {
		return gjson.Get(doc, selector).Float(), nil
	}
}

func jsonFloat64Array(selector string) func(doc string, into []float64) ([]float64, error) {
	return func(doc string, into []float64) ([]float64, error) {
		return jsonFloat64CopyInto(doc, selector, into)
	}
}

func jsonInt64(selector string) func(doc string) (int64, error) {
	return func(doc string) (int64, error) {
		return gjson.Get(doc, selector).Int(), nil
	}
}

func jsonInt64Array(selector string) func(doc string, into []int64) ([]int64, error) {
	return func(doc string, into []int64) ([]int64, error) {
		return jsonInt64CopyInto(doc, selector, into)
	}
}

func jsonString(selector string) func(doc string, into []byte) ([]byte, error) {
	return func(doc string, into []byte) ([]byte, error) {
		return copyInto(gjson.Get(doc, selector).String(), into)
	}
}

func jsonStringArray(selector string) func(doc string, into []string) ([]string, error) {
	return func(doc string, into []string) ([]string, error) {
		return jsonStringCopyInto(doc, selector, into)
	}
}

func jsonInt64CopyInto(doc, selector string, into []int64) ([]int64, error) {
	results := gjson.Get(doc, selector).Array()
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) > cap(into) {
		into = make([]int64, len(results))
	} else {
		into = into[0:len(results)]
	}
	for i, result := range results {
		into[i] = result.Int()
	}
	return into, nil
}

func jsonFloat64CopyInto(doc, selector string, into []float64) ([]float64, error) {
	results := gjson.Get(doc, selector).Array()
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) > cap(into) {
		into = make([]float64, len(results))
	} else {
		into = into[0:len(results)]
	}
	for i, result := range results {
		into[i] = result.Float()
	}
	return into, nil
}

func jsonStringCopyInto(doc, selector string, into []string) ([]string, error) {
	results := gjson.Get(doc, selector).Array()
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) > cap(into) {
		into = make([]string, len(results))
	} else {
		into = into[0:len(results)]
	}
	for i, result := range results {
		into[i] = result.String()
	}
	return into, nil
}

func copyInto(value string, into []byte) ([]byte, error) {
	if len(value) > len(into) {
		value = value[0:len(into)]
	}
	copy(into, value)
	return into[0:len(value)], nil
}
