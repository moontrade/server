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

func NewDocID(collection CollectionID, id uint64) DocID {
	*(*uint16)(unsafe.Pointer(&id)) = uint16(collection)
	return DocID(id)
}

func (id *DocID) Key() mdbx.Val {
	return mdbx.Val{
		Base: (*byte)(unsafe.Pointer(id)),
		Len:  8,
	}
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
	collectionDescriptor
	Id      CollectionID `json:"id"`
	Owner   int32        `json:"owner"`
	Created uint64       `json:"created"`
	Updated uint64       `json:"updated"`
	Indexes []IndexMeta  `json:"indexes,omitempty"`
	Schema  string       `json:"schema,omitempty"`
}

type collectionDescriptor struct {
	Kind    CollectionKind `json:"kind"`
	Name    string         `json:"name"`
	Version int64          `json:"version"`
}

func (cd *collectionDescriptor) Equals(other *collectionDescriptor) bool {
	if cd == nil {
		return other == nil
	}
	return other != nil &&
		cd.Kind == other.Kind &&
		cd.Name == other.Name &&
		cd.Version == other.Version
}

func (s *Store) EstimateCollectionCount(collectionID CollectionID) (count int64, err error) {
	err = s.store.View(func(tx *mdbx.Tx) error {
		var (
			k     = NewDocID(collectionID, 0)
			key   = k.Key()
			data  = mdbx.Val{}
			first *mdbx.Cursor
			last  *mdbx.Cursor
		)

		first, err = tx.OpenCursor(s.documentsDBI)
		if err != mdbx.ErrSuccess {
			return err
		}
		err = nil
		defer first.Close()

		last, err = tx.OpenCursor(s.documentsDBI)
		if err != mdbx.ErrSuccess {
			return err
		}
		err = nil
		defer last.Close()

		if err = first.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
			if err == mdbx.ErrNotFound {
				err = nil
				return nil
			}
			return err
		}
		id := DocID(key.U64())
		if id.CollectionID() != collectionID {
			return nil
		}
		if err = last.Get(&key, &data, mdbx.CursorPrevNoDup); err != mdbx.ErrSuccess {
			if err == mdbx.ErrNotFound {
				err = nil
				return nil
			}
			return err
		}
		lastID := DocID(key.U64())
		if lastID.CollectionID() != collectionID {
			count = 1
			return nil
		}

		count, err = mdbx.EstimateDistance(first, last)
		if err == mdbx.ErrNotFound {
			count = 1
			return nil
		}
		if err == mdbx.ErrSuccess {
			err = nil
		}
		return err
	})
	if err == mdbx.ErrSuccess {
		err = nil
	}
	return
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
	loaded     bool
}

type Cursor struct {
	c *indexStore
}

func docIDVal(key *DocID) mdbx.Val {
	return mdbx.Val{
		Base: (*byte)(unsafe.Pointer(key)),
		Len:  8,
	}
}

func (s *collectionStore) RecordID(sequence uint64) DocID {
	return NewDocID(s.id, sequence)
}

func (s *collectionStore) NextID() DocID {
	return NewDocID(s.id, atomic.AddUint64(&s.sequence, 1))
}

func (s *collectionStore) Insert(tx *Tx, data []byte, unmarshalled interface{}) (DocID, error) {
	var (
		k   = NewDocID(s.id, atomic.AddUint64(&s.sequence, 1))
		key = docIDVal(&k)
		val = mdbx.Bytes(&data)
		d   = *(*string)(unsafe.Pointer(&data))
		err error
	)
	if err = tx.Tx.Put(s.store.documentsDBI, &key, &val, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		return 0, err
	}
	// Insert indexes
	if len(s.indexes) > 0 {
		for _, index := range s.indexes {
			if err = index.insert(tx, k, d, unmarshalled); err != nil {
				return 0, err
			}
		}
	}
	return k, nil
}

func (s *collectionStore) Update(tx *Tx, id DocID, data []byte, unmarshalled interface{}) error {
	var (
		key = docIDVal(&id)
		val = mdbx.Bytes(&data)
		d   = *(*string)(unsafe.Pointer(&data))
		err error
	)
	if err := tx.Tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		return err
	}
	// Update indexes
	if len(s.indexes) > 0 {
		for _, index := range s.indexes {
			if _, err = index.update(tx, id, d, unmarshalled); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *collectionStore) Delete(tx *Tx, id DocID, unmarshalled interface{}) (bool, error) {
	var (
		key = docIDVal(&id)
		val = mdbx.Val{}
		err error
	)
	if err := tx.Tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	// Delete indexes
	if len(s.indexes) > 0 {
		data := val.UnsafeBytes()
		d := *(*string)(unsafe.Pointer(&data))
		for _, index := range s.indexes {
			if _, err = index.delete(tx, id, d, unmarshalled); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (s *collectionStore) DeleteGet(tx *Tx, id DocID, unmarshalled interface{}, onData func(data mdbx.Val)) (bool, error) {
	var (
		key = docIDVal(&id)
		val = mdbx.Val{}
		err error
	)
	if err := tx.Tx.Put(s.store.documentsDBI, &key, &val, 0); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	// Delete indexes
	if len(s.indexes) > 0 {
		data := val.UnsafeBytes()
		d := *(*string)(unsafe.Pointer(&data))
		for _, index := range s.indexes {
			if _, err = index.delete(tx, id, d, unmarshalled); err != nil {
				return false, err
			}
		}
	}
	if onData != nil {
		onData(val)
	}
	return true, nil
}
