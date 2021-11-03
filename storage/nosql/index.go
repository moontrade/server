package nosql

import "github.com/moontrade/mdbx-go"

type Index interface {
	GetName() string

	Insert(tx *mdbx.Tx, id DocID, document string) error

	Update(tx *mdbx.Tx, id DocID, document string) (bool, error)

	Delete(tx *mdbx.Tx, id DocID, document string) (bool, error)
}

type indexBase struct {
	Name string
}

type String struct {
	indexBase
	*stringStore
	Get func(doc string, into []byte) (result []byte, err error)
}

type UniqueString struct {
	indexBase
	*uniqueStringStore
	Get func(doc string, into []byte) (result []byte, err error)
}

type StringArray struct {
	indexBase
	*stringArrayStore
	Get func(doc string, into []string) (result []string, err error)
}

type Int64 struct {
	indexBase
	*int64Store
	Get func(data string) (int64, error)
}

type Int64Array struct {
	indexBase
	*int64ArrayStore
	Get func(data string, into []int64) ([]int64, error)
}

type UniqueInt64 struct {
	indexBase
	*uniqueInt64Store

	Get func(data string) (int64, error)
}

type Float64 struct {
	indexBase
	*float64Store
	Get func(data string) (float64, error)
}

type Float64Array struct {
	indexBase
	*float64ArrayStore

	Get func(data string, into []float64) ([]float64, error)
}

type UniqueFloat64 struct {
	indexBase
	*uniqueFloat64Store
	Get func(data string) (float64, error)
}

type indexStoreBase struct {
	name       string
	kind       IndexKind
	unique     bool
	array      bool
	selector   string
	store      *Store
	collection *Collection
	dbi        mdbx.DBI
	state      int32
	id         CollectionID
}

func (isb *indexStoreBase) GetName() string {
	return isb.name
}

type int64Store struct {
	indexStoreBase
	get func(doc string) (result int64, err error)
}

func (is *int64Store) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *int64Store) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *int64Store) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type uniqueInt64Store struct {
	indexStoreBase
	get func(doc string) (result int64, err error)
}

func (is *uniqueInt64Store) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *uniqueInt64Store) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *uniqueInt64Store) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type int64ArrayStore struct {
	indexStoreBase
	get func(doc string, into []int64) (result []int64, err error)
}

func (is *int64ArrayStore) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *int64ArrayStore) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *int64ArrayStore) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type float64Store struct {
	indexStoreBase
	get func(doc string) (result float64, err error)
}

func (is *float64Store) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *float64Store) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *float64Store) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type uniqueFloat64Store struct {
	indexStoreBase
	get func(doc string) (result float64, err error)
}

func (is *uniqueFloat64Store) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *uniqueFloat64Store) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *uniqueFloat64Store) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type float64ArrayStore struct {
	indexStoreBase
	get func(doc string, into []float64) (result []float64, err error)
}

func (is *float64ArrayStore) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *float64ArrayStore) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *float64ArrayStore) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type stringStore struct {
	indexStoreBase
	get func(doc string, into []byte) (result []byte, err error)
}

func (is *stringStore) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *stringStore) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *stringStore) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type uniqueStringStore struct {
	indexStoreBase
	get func(doc string, into []byte) (result []byte, err error)
}

func (is *uniqueStringStore) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *uniqueStringStore) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *uniqueStringStore) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type stringArrayStore struct {
	indexStoreBase
	get func(doc string, into []string) (result []string, err error)
}

func (is *stringArrayStore) Insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *stringArrayStore) Update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *stringArrayStore) Delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}
