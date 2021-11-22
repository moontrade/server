package nosql

import (
	"errors"
	"fmt"
	"github.com/moontrade/mdbx-go"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

var (
	collectionTypeOf    = reflect.TypeOf(Collection{})
	int64TypeOf         = reflect.TypeOf(Int64{})
	uniqueInt64TypeOf   = reflect.TypeOf(Int64Unique{})
	int64ArrayTypeOf    = reflect.TypeOf(Int64Array{})
	float64TypeOf       = reflect.TypeOf(Float64{})
	uniqueFloat64TypeOf = reflect.TypeOf(Float64Unique{})
	float64ArrayTypeOf  = reflect.TypeOf(Float64Array{})
	stringTypeOf        = reflect.TypeOf(String{})
	uniqueStringTypeOf  = reflect.TypeOf(StringUnique{})
	stringArrayTypeOf   = reflect.TypeOf(StringArray{})
	schemaTypeOf        = reflect.TypeOf(Schema{})

	errNotCollectionType = errors.New("not collection type")
	errNotIndexType      = errors.New("not index type")

	ErrAlreadyLoaded          = errors.New("already loaded")
	ErrSchemaFieldNotFound    = errors.New("schema field not found: add '*nosql.Schema' field")
	ErrCollectionStore        = errors.New("collection store not exist")
	ErrCollectionIDExhaustion = errors.New("collection id exhaustion")
	ErrIndexIDExhaustion      = errors.New("index id exhaustion")
	ErrSchemaIDExhaustion     = errors.New("schema id exhaustion")
)

// Schema provides a flat list of named Collections and their indexes.
// It does NOT enforce any layout of the individual document specs. JSON
// formatted documents support indexing natively. JSON indexes utilize gjson
// selector to extract the field(s) required to build index.
//
// It is recommended to use the strongly typed Schema pattern. Hydrating a
// Schema in a Store will produce and apply an evolution to keep the Schema
// consistent. This simplifies a lot of bug prone manual index and collection
// bookkeeping.
type Schema struct {
	Meta          SchemaMeta
	Collections   []Collection
	collectionMap map[string]Collection
	store         *Store
	loaded        bool
	mu            sync.Mutex
}

func (s *Schema) IsLoaded() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store != nil
}

func (s *Schema) buildMeta() *SchemaMeta {
	m := s.Meta
	m.Collections = make([]CollectionMeta, len(s.Collections))
	for i, col := range s.Collections {
		cm := col.CollectionMeta
		if len(col.indexes) > 0 {
			cm.Indexes = make([]IndexMeta, len(col.indexes))
			for ii, index := range col.indexes {
				cm.Indexes[ii] = index.Meta()
			}
			sort.Sort(indexMetasSlice(cm.Indexes))
		}
		m.Collections[i] = cm
	}
	s.Meta = m
	return &m
}

func (s *Schema) Update(fn func(tx *Tx) error) error {
	return s.store.store.Update(func(tx *mdbx.Tx) error {
		txn := s.store.tx
		txn.Reset(tx)
		txn.Tx = tx
		if !s.loaded {
			var err error
			sort.Sort(collectionsByIDSlice(s.Collections))
			for _, collection := range s.Collections {
				if err = collection.collectionStore.ensureLoaded(txn); err != nil {
					return err
				}
			}
			s.loaded = true
		}
		return fn(txn)
	})
}

func (s *Schema) View(fn func(tx *Tx) error) error {
	return s.store.store.View(func(tx *mdbx.Tx) error {
		var txn Tx
		txn.store = s.store
		txn.Reset(tx)
		return fn(&txn)
	})
}

type SchemaMeta struct {
	Id          uint32           `json:"id"`
	UID         string           `json:"uid"`
	Name        string           `json:"name"`
	Pkg         string           `json:"pkg"`
	FQN         string           `json:"fqn"`
	Checksum    uint64           `json:"checksum"`
	Collections []CollectionMeta `json:"collections"`
	Mutations   struct {
	} `json:"mutations"`
}

type SchemaMutation struct {
}

func fqnOf(t reflect.Type) string {
	pkg := t.PkgPath()
	if len(pkg) == 0 {
		return t.Name()
	}
	return fmt.Sprintf("%s.%s", pkg, t.Name())
}

func ParseSchema(prototype interface{}) (*Schema, error) {
	return ParseSchemaWithUID("", prototype)
}

func ParseSchemaWithUID(uid string, prototype interface{}) (*Schema, error) {
	val := reflect.ValueOf(prototype)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()
	if t.Kind() != reflect.Struct {
		return nil, errors.New("not struct")
	}

	fqn := fqnOf(t)
	if len(uid) == 0 {
		uid = fqn
	} else if uid == "@" {
		uid = ""
	}

	var (
		schema = &Schema{
			Meta: SchemaMeta{
				UID:  uid,
				Name: t.Name(),
				Pkg:  t.PkgPath(),
				FQN:  fqn,
			},
			Collections:   make([]Collection, 0, 16),
			collectionMap: make(map[string]Collection),
		}
		numFields        = val.NumField()
		schemaFieldFound bool
		schemaField      reflect.StructField
		schemaValue      reflect.Value
	)

LOOP:
	for i := 0; i < numFields; i++ {
		fieldValue := val.Field(i)
		fieldType := t.Field(i)

		ft := fieldValue.Type()
		if ft.Kind() == reflect.Ptr {
			if ft.Elem().AssignableTo(schemaTypeOf) {
				schemaFieldFound = true
				schemaField = fieldType
				schemaValue = fieldValue
				continue LOOP
			}
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.Struct {
			continue
		}

		col, err := parseCollection(fieldValue.Interface(), fieldValue, ft, &fieldType)
		if err != nil {
			return nil, err
		}
		_, ok := schema.collectionMap[col.Name]
		if ok {
			return nil, fmt.Errorf("duplicate collections named %s", col.Name)
		}
		schema.collectionMap[col.Name] = col

		// Find Collection field
		collectionField, ok := fieldType.Type.FieldByName("Collection")
		if ok {
			fieldValue = fieldValue.FieldByName("Collection")
			_ = collectionField
		}

		*(*Collection)(unsafe.Pointer(fieldValue.UnsafeAddr())) = col
		schema.Collections = append(schema.Collections, col)
	}

	if !schemaFieldFound {
		return nil, ErrSchemaFieldNotFound
	}

	_ = schemaField
	schemaValue.Set(reflect.ValueOf(schema))

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

		var inter Marshaller
		if fieldType.Name == "Marshaller" && fieldType.Type.Implements(reflect.TypeOf(inter)) {
			col.Marshaller = fieldValueType.Interface().(Marshaller)
		}

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
				col.Name = snakeCase(t.Name())
			}

			break collectionLoop
		}
	}

	if !foundCollection {
		return col, errNotCollectionType
	}

indexLoop:
	for i := 0; i < t.NumField(); i++ {
		var (
			fieldValueType = valueType.Field(i)
			fieldType      = t.Field(i)
		)
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
			if !isUpperChar(fieldType.Name[0]) {
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
			indexName := index.Name()
			existing := col.indexMap[indexName]
			if existing != nil {
				return col, fmt.Errorf("%s.%s duplicate index named %s", col.Name, fieldType.Name, indexName)
			}
			col.indexMap[indexName] = index
		}
	}

	if col.Type != nil && col.Marshaller == nil {
		col.Marshaller = MarshallerOfType(col.Type)
	}

	col.marshaller = col.Marshaller

	return col, nil
}

func parseIndex(
	val reflect.Value,
	field reflect.StructField,
) (Index, error) {
	var (
		name     = strings.TrimSpace(field.Tag.Get("name"))
		ft       = field.Type
		selector = field.Tag.Get("@")
		version  = field.Tag.Get("version")
	)
	if len(name) == 0 {
		name = snakeCase(field.Name)
	}
	switch {
	case ft.AssignableTo(int64TypeOf):
		index := NewInt64(name, selector, version, val.Interface().(Int64).ValueOf)
		*(*Int64)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

	case ft.AssignableTo(uniqueInt64TypeOf):
		index := NewInt64Unique(name, selector, version, val.Interface().(Int64Unique).ValueOf)
		*(*Int64Unique)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

	//case ft.AssignableTo(int64ArrayTypeOf):
	//	index := NewInt64Array(name, selector, version, val.Interface().(Int64Array).ValueOf)
	//	*(*Int64Array)(unsafe.Pointer(val.UnsafeAddr())) = *index
	//	return index, nil

	case ft.AssignableTo(float64TypeOf):
		index := NewFloat64(name, selector, version, val.Interface().(Float64).ValueOf)
		*(*Float64)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

	case ft.AssignableTo(uniqueFloat64TypeOf):
		index := NewFloat64Unique(name, selector, version, val.Interface().(Float64Unique).ValueOf)
		*(*Float64Unique)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

	//case ft.AssignableTo(float64ArrayTypeOf):
	//	index := NewFloat64Array(name, selector, version, val.Interface().(Float64Array).ValueOf)
	//	*(*Float64Array)(unsafe.Pointer(val.UnsafeAddr())) = *index
	//	return index, nil

	case ft.AssignableTo(stringTypeOf):
		index := NewString(name, selector, version, val.Interface().(String).ValueOf)
		*(*String)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

	case ft.AssignableTo(uniqueStringTypeOf):
		index := NewStringUnique(name, selector, version, val.Interface().(StringUnique).ValueOf)
		*(*StringUnique)(unsafe.Pointer(val.UnsafeAddr())) = *index
		return index, nil

		//case ft.AssignableTo(stringArrayTypeOf):
		//	index := NewStringArray(name, selector, version, val.Interface().(StringArray).ValueOf)
		//	*(*StringArray)(unsafe.Pointer(val.UnsafeAddr())) = *index
		//	return index, nil
	}
	return nil, errNotIndexType
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
