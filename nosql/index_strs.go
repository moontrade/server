package nosql

type StringArrayValueOf func(doc string, unmarshalled interface{}, into []string) (result []string, err error)

type StringArray struct {
	indexBase
	ValueOf StringArrayValueOf
}

func NewStringArray(
	name, selector, version string,
	valueOf StringArrayValueOf,
) *StringArray {
	if valueOf == nil {
		valueOf = jsonStringArray(selector)
	}
	return &StringArray{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindString, false, true),
	}
}
