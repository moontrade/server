package nosql

import "github.com/moontrade/mdbx-go"

type Sort byte

const (
	SortDefault    Sort = 0
	SortAscending  Sort = 1
	SortDescending Sort = 2
)

type Index interface {
	ID() uint32

	Name() string

	Owner() CollectionID

	Meta() IndexMeta

	setMeta(m IndexMeta)

	getStore() *indexStore

	setStore(s *indexStore)

	insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error

	update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error)

	delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error)
}

type indexBase struct {
	meta  IndexMeta
	store *indexStore
}

func newIndexBase(
	name, selector, version string,
	kind IndexKind,
	unique, array bool,
) indexBase {
	return indexBase{
		store: &indexStore{},
		meta: IndexMeta{indexDescriptor: indexDescriptor{
			Name:     name,
			Selector: selector,
			Version:  version,
			Kind:     kind,
			Unique:   unique,
			Array:    array,
		}}}
}

func (isb *indexBase) ID() uint32 {
	return isb.meta.ID
}

func (isb *indexBase) Name() string {
	return isb.meta.Name
}

func (isb *indexBase) Owner() CollectionID {
	return isb.meta.Owner
}

func (ib *indexBase) Meta() IndexMeta {
	return ib.meta
}

func (ib *indexBase) setMeta(m IndexMeta) {
	ib.meta = m
}

func (ib *indexBase) getStore() *indexStore {
	return ib.store
}

func (ib *indexBase) setStore(s *indexStore) {
	ib.store = s
}

type IndexMeta struct {
	indexDescriptor
	Owner CollectionID `json:"owner"`
	DBI   mdbx.DBI     `json:"dbi"`
	State int32        `json:"state"`
}

type indexDescriptor struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Selector    string    `json:"selector"`
	ID          uint32    `json:"id"`
	Kind        IndexKind `json:"kind"`
	Unique      bool      `json:"unique"`
	Array       bool      `json:"array"`
	Version     string    `json:"version"`
}

func (im IndexMeta) didChange(other IndexMeta) bool {
	return im.Name == other.Name &&
		im.Selector == other.Selector &&
		im.Kind == other.Kind &&
		im.Unique == other.Unique &&
		im.Array == other.Array &&
		im.Version == other.Version
}

type indexStore struct {
	store      *Store
	collection *collectionStore
	count      uint64
	bytes      uint64
}

////////////////////////////////////////////////////////////////////////////////////////
// Int64
////////////////////////////////////////////////////////////////////////////////////////

type Int64ValueOf func(data string, unmarshalled interface{}) (int64, error)

type Int64 struct {
	indexBase
	ValueOf Int64ValueOf
}

func NewInt64(
	name, selector, version string,
	valueOf Int64ValueOf,
) *Int64 {
	if valueOf == nil {
		valueOf = jsonInt64(selector)
	}
	return &Int64{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindInt64, false, false),
	}
}

func (is *Int64) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	// Layout
	return nil
}

func (is *Int64) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Int64) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Int64Unique
////////////////////////////////////////////////////////////////////////////////////////

type Int64Unique struct {
	indexBase
	ValueOf Int64ValueOf
}

func NewInt64Unique(
	name, selector, version string,
	valueOf Int64ValueOf,
) *Int64Unique {
	if valueOf == nil {
		valueOf = jsonInt64(selector)
	}
	return &Int64Unique{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindInt64, true, false),
	}
}

func (is *Int64Unique) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *Int64Unique) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Int64Unique) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Int64Array
////////////////////////////////////////////////////////////////////////////////////////

type Int64ArrayValueOf func(data string, unmarshalled interface{}, into []int64) ([]int64, error)

type Int64Array struct {
	indexBase
	ValueOf Int64ArrayValueOf
}

func NewInt64Array(
	name, selector, version string,
	valueOf Int64ArrayValueOf,
) *Int64Array {
	if valueOf == nil {
		valueOf = jsonInt64Array(selector)
	}
	return &Int64Array{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindInt64, false, true),
	}
}

func (is *Int64Array) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *Int64Array) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Int64Array) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Float64
////////////////////////////////////////////////////////////////////////////////////////

type Float64ValueOf func(data string, unmarshalled interface{}) (float64, error)

type Float64 struct {
	indexBase
	ValueOf Float64ValueOf
}

func NewFloat64(
	name, selector, version string,
	valueOf Float64ValueOf,
) *Float64 {
	if valueOf == nil {
		valueOf = jsonFloat64(selector)
	}
	return &Float64{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindFloat64, false, false),
	}
}

func (is *Float64) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *Float64) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Float64) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Float64Unique
////////////////////////////////////////////////////////////////////////////////////////

type Float64Unique struct {
	indexBase
	ValueOf Float64ValueOf
}

func NewFloat64Unique(
	name, selector, version string,
	valueOf Float64ValueOf,
) *Float64Unique {
	if valueOf == nil {
		valueOf = jsonFloat64(selector)
	}
	return &Float64Unique{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindFloat64, true, false),
	}
}

func (is *Float64Unique) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *Float64Unique) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Float64Unique) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Float64Array
////////////////////////////////////////////////////////////////////////////////////////

type Float64ArrayValueOf func(data string, unmarshalled interface{}, into []float64) ([]float64, error)

type Float64Array struct {
	indexBase
	ValueOf Float64ArrayValueOf
}

func (is *Float64Array) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *Float64Array) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *Float64Array) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
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

////////////////////////////////////////////////////////////////////////////////////////
// String
////////////////////////////////////////////////////////////////////////////////////////

type StringValueOf func(doc string, unmarshalled interface{}, into []byte) (result []byte, err error)

type String struct {
	indexBase
	ValueOf StringValueOf
}

func NewString(
	name, selector, version string,
	valueOf StringValueOf,
) *String {
	if valueOf == nil {
		valueOf = jsonString(selector)
	}
	return &String{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindString, false, false),
	}
}

func (is *String) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *String) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *String) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// StringUnique
////////////////////////////////////////////////////////////////////////////////////////

type StringUnique struct {
	indexBase
	ValueOf StringValueOf
}

func NewStringUnique(
	name, selector, version string,
	valueOf StringValueOf,
) *StringUnique {
	if valueOf == nil {
		valueOf = jsonString(selector)
	}
	return &StringUnique{
		ValueOf:   valueOf,
		indexBase: newIndexBase(name, selector, version, IndexKindString, true, false),
	}
}

func (is *StringUnique) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *StringUnique) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *StringUnique) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// StringArray
////////////////////////////////////////////////////////////////////////////////////////

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

func (is *StringArray) insert(tx *Tx, id DocID, document string, unmarshalled interface{}) error {
	return nil
}

func (is *StringArray) update(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

func (is *StringArray) delete(tx *Tx, id DocID, document string, unmarshalled interface{}) (bool, error) {
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// FullText
////////////////////////////////////////////////////////////////////////////////////////

type FullText struct{}
