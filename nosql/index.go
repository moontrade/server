package nosql

import (
	"errors"
	"github.com/moontrade/mdbx-go"
)

var (
	ErrSkip             = errors.New("skip")
	ErrIndexCorrupted   = errors.New("index corrupted")
	ErrUniqueConstraint = errors.New("unique constraint")
	ErrIndexKeyTooBig   = errors.New("index key too big")
)

const (
	MaxIndexKeySize = 4000
)

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

	doInsert(tx *Tx) error

	doUpdate(tx *Tx) error

	doDelete(tx *Tx) error
}

type indexBase struct {
	IndexMeta
	store *indexStore `json:"-"`
}

func newIndexBase(
	name, selector, version string,
	kind IndexKind,
	unique, array bool,
) indexBase {
	return indexBase{
		store: &indexStore{},
		IndexMeta: IndexMeta{indexDescriptor: indexDescriptor{
			Name:     name,
			Selector: selector,
			Version:  version,
			Kind:     kind,
			Unique:   unique,
			Array:    array,
		}}}
}

func (ib *indexBase) ID() uint32 {
	return ib.IndexMeta.ID
}

func (ib *indexBase) Name() string {
	return ib.IndexMeta.Name
}

func (ib *indexBase) Owner() CollectionID {
	return ib.IndexMeta.Owner
}

func (ib *indexBase) Meta() IndexMeta {
	return ib.IndexMeta
}

func (ib *indexBase) setMeta(m IndexMeta) {
	ib.IndexMeta = m
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

func (im IndexMeta) equals(other IndexMeta) bool {
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
	index      Index
	estimated  uint64
	bytes      uint64
	loaded     bool
}
