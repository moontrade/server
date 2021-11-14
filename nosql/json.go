package nosql

import (
	"github.com/tidwall/gjson"
)

func jsonFloat64(selector string) Float64ValueOf {
	return func(doc string, unmarshalled interface{}) (float64, error) {
		return gjson.Get(doc, selector).Float(), nil
	}
}

func jsonFloat64Array(selector string) Float64ArrayValueOf {
	return func(doc string, unmarshalled interface{}, into []float64) ([]float64, error) {
		return jsonFloat64CopyInto(doc, selector, into)
	}
}

func jsonInt64(selector string) Int64ValueOf {
	return func(doc string, unmarshalled interface{}) (int64, error) {
		return gjson.Get(doc, selector).Int(), nil
	}
}

func jsonInt64Array(selector string) Int64ArrayValueOf {
	return func(doc string, unmarshalled interface{}, into []int64) ([]int64, error) {
		return jsonInt64CopyInto(doc, selector, into)
	}
}

func jsonString(selector string) StringValueOf {
	return func(doc string, unmarshalled interface{}, into []byte) ([]byte, error) {
		return copyInto(gjson.Get(doc, selector).String(), into)
	}
}

func jsonStringArray(selector string) StringArrayValueOf {
	return func(doc string, unmarshalled interface{}, into []string) ([]string, error) {
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
