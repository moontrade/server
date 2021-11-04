package nosql

import (
	"github.com/moontrade/mdbx-go"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// CollectionID is an UID for a single collection used in unique DocID
// instead of variable length string names. This provides deterministic
// key operations regardless of length of collection name.
type CollectionID uint16

// DocID
type DocID uint64

// CollectionID is an UID for a single collection used in unique DocID
// instead of variable length string names. This provides deterministic
// key operations regardless of length of collection name.
func (r DocID) CollectionID() CollectionID {
	return *(*CollectionID)(unsafe.Pointer(&r))
}

// Sequence 48bit unsigned integer that represents the unique sequence within the collection.
func (r DocID) Sequence() uint64 {
	result := uint64(r)
	*(*CollectionID)(unsafe.Pointer(&result)) = 0
	return result
}

func NewRecordID(collection CollectionID, id uint64) DocID {
	*(*uint16)(unsafe.Pointer(&id)) = uint16(collection)
	return DocID(id)
}

type CollectionKind int

const (
	CollectionTypeCustom   CollectionKind = 0
	CollectionTypeJson     CollectionKind = 1
	CollectionTypeProto    CollectionKind = 2
	CollectionTypeProtobuf CollectionKind = 3
)

type IndexKind int

const (
	IndexKindUnknown   IndexKind = 0
	IndexKindInt64     IndexKind = 1
	IndexKindFloat64   IndexKind = 2
	IndexKindString    IndexKind = 3
	IndexKindComposite IndexKind = 10
	//IndexTypeSpatial IndexKind = 11
)

type GetValue func(doc string, into []byte) (result []byte, err error)

type Collection struct {
	*collectionStore
	Name string
}

type CollectionMeta struct {
	Name     string         `json:"name"`
	Created  uint64         `json:"created"`
	Updated  uint64         `json:"updated"`
	Schema   int32          `json:"schema"`
	Kind     CollectionKind `json:"kind"`
	Checksum uint64         `json:"x"`
	Id       CollectionID   `json:"id"`
	Indexes  []IndexMeta    `json:"indexes,omitempty"`
}

type collectionStore struct {
	CollectionMeta
	Type       reflect.Type
	store      *Store
	indexes    []Index
	indexMap   map[string]Index
	id         CollectionID
	minID      DocID
	maxID      DocID
	sequence   uint64
	bytes      uint64
	indexBytes uint64
}

type Cursor struct {
	c *collectionStore
}

func docIDVal(key *DocID) mdbx.Val {
	return mdbx.Val{
		Base: (*byte)(unsafe.Pointer(key)),
		Len:  8,
	}
}

func (s *collectionStore) RecordID(sequence uint64) DocID {
	return NewRecordID(s.id, sequence)
}

func (s *collectionStore) NextID() DocID {
	return NewRecordID(s.id, atomic.AddUint64(&s.sequence, 1))
}

func (s *collectionStore) Insert(tx *mdbx.Tx, data []byte) (DocID, error) {
	var (
		k   = NewRecordID(s.id, atomic.AddUint64(&s.sequence, 1))
		key = docIDVal(&k)
		val = mdbx.BytesVal(&data)
		d   = *(*string)(unsafe.Pointer(&data))
		err error
	)
	if err = tx.Put(s.store.documentsDBI, &key, &val, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		return 0, err
	}
	// Insert indexes
	if len(s.indexes) > 0 {
		for _, index := range s.indexes {
			if err = index.insert(tx, k, d); err != nil {
				return 0, err
			}
		}
	}
	return k, nil
}

func (s *collectionStore) Update(tx *mdbx.Tx, id DocID, data []byte) error {
	var (
		key = docIDVal(&id)
		val = mdbx.BytesVal(&data)
		d   = *(*string)(unsafe.Pointer(&data))
		err error
	)
	if err := tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		return err
	}
	// Update indexes
	if len(s.indexes) > 0 {
		for _, index := range s.indexes {
			if _, err = index.update(tx, id, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *collectionStore) Delete(tx *mdbx.Tx, id DocID) (bool, error) {
	var (
		key = docIDVal(&id)
		val = mdbx.Val{}
		err error
	)
	if err := tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	// Delete indexes
	if len(s.indexes) > 0 {
		data := val.Unsafe()
		d := *(*string)(unsafe.Pointer(&data))
		for _, index := range s.indexes {
			if _, err = index.delete(tx, id, d); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (s *collectionStore) DeleteGet(tx *mdbx.Tx, id DocID, onData func(data mdbx.Val)) (bool, error) {
	var (
		key = docIDVal(&id)
		val = mdbx.Val{}
		err error
	)
	if err := tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	// Delete indexes
	if len(s.indexes) > 0 {
		data := val.Unsafe()
		d := *(*string)(unsafe.Pointer(&data))
		for _, index := range s.indexes {
			if _, err = index.delete(tx, id, d); err != nil {
				return false, err
			}
		}
	}
	if onData != nil {
		onData(val)
	}
	return true, nil
}
