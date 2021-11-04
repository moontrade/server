package nosql

import "github.com/moontrade/mdbx-go"

type Index interface {
	ID() uint32

	Owner() CollectionID

	Name() string

	insert(tx *mdbx.Tx, id DocID, document string) error

	update(tx *mdbx.Tx, id DocID, document string) (bool, error)

	delete(tx *mdbx.Tx, id DocID, document string) (bool, error)
}

type indexBase struct {
	meta  IndexMeta
	store *indexStore
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
	Version     int32     `json:"version"`
}

func (im *IndexMeta) Equals(other *IndexMeta) bool {
	if im == nil {
		if other == nil {
			return true
		}
		return false
	}
	if other == nil {
		return false
	}
	return im.indexDescriptor == other.indexDescriptor
}

type indexStore struct {
	store *Store
	count uint64
	bytes uint64
}

func (isb *indexBase) Name() string {
	return isb.meta.Name
}

func (isb *indexBase) ID() uint32 {
	return isb.meta.ID
}

func (isb *indexBase) Owner() CollectionID {
	return isb.meta.Owner
}

type String struct {
	indexBase
	Value func(doc string, into []byte) (result []byte, err error)
}

func (is *String) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *String) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *String) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type StringUnique struct {
	indexBase
	Value func(doc string, into []byte) (result []byte, err error)
}

func (is *StringUnique) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *StringUnique) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *StringUnique) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type StringArray struct {
	indexBase
	Value func(doc string, into []string) (result []string, err error)
}

func (is *StringArray) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *StringArray) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *StringArray) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Int64 struct {
	indexBase
	Value func(data string) (int64, error)
}

func (is *Int64) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Int64) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Int64) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Int64Array struct {
	indexBase
	Value func(data string, into []int64) ([]int64, error)
}

func (is *Int64Array) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Int64Array) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Int64Array) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Int64Unique struct {
	indexBase
	Value func(data string) (int64, error)
}

func (is *Int64Unique) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Int64Unique) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Int64Unique) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Float64 struct {
	indexBase
	Value func(data string) (float64, error)
}

func (is *Float64) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Float64) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Float64) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Float64Array struct {
	indexBase
	Value func(data string, into []float64) ([]float64, error)
}

func (is *Float64Array) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Float64Array) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Float64Array) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

type Float64Unique struct {
	indexBase
	Value func(data string) (float64, error)
}

func (is *Float64Unique) insert(tx *mdbx.Tx, id DocID, document string) error {
	return nil
}

func (is *Float64Unique) update(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}

func (is *Float64Unique) delete(tx *mdbx.Tx, id DocID, document string) (bool, error) {
	return false, nil
}
