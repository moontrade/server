package nosql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var (
	collectionTypeOf    = reflect.TypeOf(Collection{})
	int64TypeOf         = reflect.TypeOf(Int64{})
	uniqueInt64TypeOf   = reflect.TypeOf(UniqueInt64{})
	int64ArrayTypeOf    = reflect.TypeOf(Int64Array{})
	float64TypeOf       = reflect.TypeOf(Float64{})
	uniqueFloat64TypeOf = reflect.TypeOf(UniqueFloat64{})
	float64ArrayTypeOf  = reflect.TypeOf(Float64Array{})
	stringTypeOf        = reflect.TypeOf(String{})
	uniqueStringTypeOf  = reflect.TypeOf(UniqueString{})
	stringArrayTypeOf   = reflect.TypeOf(StringArray{})

	errNotCollectionType = errors.New("not collection type")
	errNotIndexType      = errors.New("not index type")
)

type Schema struct {
	Collections   []Collection
	CollectionMap map[string]Collection
}

func Load(prototype interface{}) (*Schema, error) {
	val := reflect.ValueOf(prototype)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("not struct")
	}

	schema := &Schema{
		Collections:   make([]Collection, 0, 16),
		CollectionMap: make(map[string]Collection),
	}
	numFields := val.NumField()
	for i := 0; i < numFields; i++ {
		fieldValue := val.Field(i)
		fieldType := t.Field(i)

		ft := fieldValue.Type()
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.Struct {
			continue
		}

		col, err := parseCollection(fieldValue.Interface(), fieldValue, ft, &fieldType)
		if err != nil {
			return nil, err
		}
		_, ok := schema.CollectionMap[col.Name]
		if ok {
			return nil, fmt.Errorf("duplicate collections named %s", col.Name)
		}
		schema.CollectionMap[col.Name] = col
		*(*Collection)(unsafe.Pointer(fieldValue.UnsafeAddr())) = col
		schema.Collections = append(schema.Collections, col)
	}

	return schema, nil
}

func parseCollection(
	value interface{},
	valueType reflect.Value,
	t reflect.Type,
	tf *reflect.StructField,
) (Collection, error) {
	col := Collection{
		collectionStore: &collectionStore{},
	}
	var foundCollection = false

	// Init collection
collectionLoop:
	for i := 0; i < t.NumField(); i++ {
		fieldValueType := valueType.Field(i)
		fieldType := t.Field(i)
		switch fieldType.Name {

		case "Collection":
			foundCollection = true

			ct := fieldType.Type
			if ct.Kind() == reflect.Ptr {
				if deref(ct).AssignableTo(collectionTypeOf) {
					return Collection{}, fmt.Errorf("%s.Collection must be of type Collection not pointer type *Collection", t.Name())
				}
				return Collection{}, fmt.Errorf("%s.Collection must be of type Collection", t.Name())
			}
			if !ct.AssignableTo(collectionTypeOf) {
				return Collection{}, fmt.Errorf("%s.Collection must be of type Collection", t.Name())
			}

			colBase, ok := fieldValueType.Interface().(Collection)
			if !ok {
				return Collection{}, fmt.Errorf("%s.Collection must be of type nosql.Collection", t.Name())
			}
			col.Name = colBase.Name
			colBase.collectionStore = col.collectionStore

			if tf != nil {
				name := tf.Tag.Get("name")
				if len(name) > 0 {
					col.Name = name
				} else if len(col.Name) == 0 {
					col.Name = strings.ToLower(tf.Name)
				}
			}
			if len(col.Name) == 0 {
				col.Name = strings.ToLower(t.Name())
			}

			col.name = col.Name
			break collectionLoop
		}
	}

	if !foundCollection {
		return col, errNotCollectionType
	}

indexLoop:
	for i := 0; i < t.NumField(); i++ {
		fieldValueType := valueType.Field(i)
		fieldType := t.Field(i)
		switch fieldType.Name {
		case "Collection":
			continue indexLoop

		case "_":
			ofType := fieldType.Type
			for ofType.Kind() == reflect.Ptr {
				ofType = ofType.Elem()
			}
			col.Type = ofType

		default:
			if len(fieldType.Name) == 0 {
				continue indexLoop
			}
			if !isUpper(fieldType.Name[0]) {
				continue indexLoop
			}
			ct := fieldType.Type
			if ct.Kind() == reflect.Ptr {
				return Collection{}, fmt.Errorf("%s.%s must be of type %s not pointer type *%s", t.Name(), ct.Name(), ct.Name(), ct.Name())
			}

			index, err := parseIndex(fieldValueType, fieldType)
			if err != nil {
				if err == errNotIndexType {
					continue indexLoop
				}
			}

			if col.indexes == nil {
				col.indexes = make([]Index, 0, 4)
			}
			if col.indexMap == nil {
				col.indexMap = make(map[string]Index)
			}
			col.indexes = append(col.indexes, index)
			indexName := index.GetName()
			existing := col.indexMap[indexName]
			if existing != nil {
				return col, fmt.Errorf("%s.%s duplicate index named %s", col.Name, fieldType.Name, indexName)
			}
			col.indexMap[indexName] = index
		}
	}

	return col, nil
}

func parseIndex(
	val reflect.Value,
	field reflect.StructField,
) (Index, error) {
	name := strings.TrimSpace(field.Tag.Get("name"))
	if len(name) == 0 {
		name = strings.ToLower(field.Name)
	}

	fieldType := field.Type
	selector := field.Tag.Get("@")

	if fieldType.AssignableTo(int64TypeOf) {
		indexCast := val.Interface().(Int64)
		index := Int64{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonInt64(selector)
		}
		index.int64Store = &int64Store{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeInt64,
			unique: false,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(uniqueInt64TypeOf) {
		indexCast := val.Interface().(UniqueInt64)
		index := UniqueInt64{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonInt64(selector)
		}
		index.uniqueInt64Store = &uniqueInt64Store{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeInt64,
			unique: true,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(int64ArrayTypeOf) {
		indexCast := val.Interface().(Int64Array)
		index := Int64Array{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonInt64Array(selector)
		}
		index.int64ArrayStore = &int64ArrayStore{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeInt64,
			unique: false,
			array:  true,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(float64TypeOf) {
		indexCast := val.Interface().(Float64)
		index := Float64{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonFloat64(selector)
		}
		index.float64Store = &float64Store{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeFloat64,
			unique: false,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(uniqueFloat64TypeOf) {
		indexCast := val.Interface().(UniqueFloat64)
		index := UniqueFloat64{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonFloat64(selector)
		}
		index.uniqueFloat64Store = &uniqueFloat64Store{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeFloat64,
			unique: true,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(float64ArrayTypeOf) {
		indexCast := val.Interface().(Float64Array)
		index := Float64Array{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonFloat64Array(selector)
		}
		index.float64ArrayStore = &float64ArrayStore{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeFloat64,
			unique: false,
			array:  true,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(stringTypeOf) {
		indexCast := val.Interface().(String)
		index := String{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonString(selector)
		}
		index.stringStore = &stringStore{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeString,
			unique: false,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(uniqueStringTypeOf) {
		indexCast := val.Interface().(UniqueString)
		index := UniqueString{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonString(selector)
		}
		index.uniqueStringStore = &uniqueStringStore{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeString,
			unique: true,
			array:  false,
		}, get: index.Get}
		return index, nil
	}
	if fieldType.AssignableTo(stringArrayTypeOf) {
		indexCast := val.Interface().(StringArray)
		index := StringArray{}
		index.Name = indexCast.Name
		index.Get = indexCast.Get
		if len(index.Name) == 0 {
			index.Name = name
		}
		if index.Get == nil {
			index.Get = jsonStringArray(selector)
		}
		index.stringArrayStore = &stringArrayStore{indexStoreBase: indexStoreBase{
			name:   name,
			kind:   IndexTypeString,
			unique: false,
			array:  true,
		}, get: index.Get}
		return index, nil
	}
	return nil, errNotIndexType
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func isLower(c byte) bool {
	switch c {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		return true
	}
	return false
}

func isUpper(c byte) bool {
	switch c {
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		return true
	}
	return false
}
