package nosql

import (
	"errors"
	"github.com/moontrade/mdbx-go"
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var (
	ErrDocumentNil = errors.New("document nil")
)

// CollectionID is an UID for a single collection used in unique DocID
// instead of variable length string names. This provides deterministic
// key operations regardless of length of collection name.
type CollectionID uint16

const (
	MaxDocSequence = uint64(math.MaxUint64) / uint64(math.MaxUint16)
)

// DocID
type DocID uint64

// CollectionID is an UID for a single collection used in unique DocID
// instead of variable length string names. This provides deterministic
// key operations regardless of length of collection name.
func (r DocID) CollectionID() CollectionID {
	return CollectionID(r / DocID(MaxDocSequence))
}

// Sequence 48bit unsigned integer that represents the unique sequence within the collection.
func (r DocID) Sequence() uint64 {
	return uint64(r) - uint64(r.CollectionID())*MaxDocSequence
}

func NewDocID(collection CollectionID, sequence uint64) DocID {
	if sequence > MaxDocSequence {
		sequence = MaxDocSequence
	}
	return DocID((uint64(collection) * MaxDocSequence) + sequence)
}

func (id *DocID) Key() mdbx.Val {
	return mdbx.Val{
		Base: (*byte)(unsafe.Pointer(id)),
		Len:  8,
	}
}

type Document struct {
	ID        DocID       `json:"i"`
	Timestamp uint64      `json:"t"`
	Data      interface{} `json:"d"`
}

type CollectionKind int

const (
	CollectionTypeJson     CollectionKind = 0
	CollectionTypeProto    CollectionKind = 1
	CollectionTypeProtobuf CollectionKind = 2
	CollectionTypeMsgpack  CollectionKind = 3
	CollectionTypeCustom   CollectionKind = 10
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

//type GetValue func(doc string, into []byte) (result []byte, err error)

type Collection struct {
	*collectionStore
	Name       string
	Marshaller Marshaller
}

type CollectionMeta struct {
	collectionDescriptor
	Id      CollectionID `json:"id"`
	Owner   uint32       `json:"owner"`
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

type collectionStore struct {
	CollectionMeta
	Type       reflect.Type
	marshaller Marshaller
	store      *Store
	indexes    []Index
	indexMap   map[string]Index
	minID      DocID
	sequence   uint64
	estimated  int64
	err        error
	loaded     bool
}

type Cursor struct {
	c *indexStore
}

func (cs *collectionStore) DocID(sequence uint64) DocID {
	return NewDocID(cs.Id, sequence)
}

func (cs *collectionStore) NextID() DocID {
	return NewDocID(cs.Id, atomic.AddUint64(&cs.sequence, 1))
}

func (cs *collectionStore) MinID() DocID {
	return cs.minID
}

func (cs *collectionStore) MaxID() DocID {
	return NewDocID(cs.Id, cs.sequence)
}

func (cs *collectionStore) load(begin, end *mdbx.Cursor) (count int64, min, max DocID, err error) {
	var (
		collectionID = cs.Id
		k            = NewDocID(collectionID, 0)
		key          = k.Key()
		data         = mdbx.Val{}
		id           DocID
		lastID       DocID
		colID        = k.CollectionID()
		seq          = k.Sequence()
	)

	if err = begin.Get(&key, &data, mdbx.CursorSetRange); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
			goto DONE
		} else {
			cs.err = err
			return
		}
	}
	id = DocID(key.U64())
	colID = id.CollectionID()
	seq = id.Sequence()
	if id.CollectionID() != collectionID {
		goto DONE
	}

	k = NewDocID(collectionID, MaxDocSequence)
	key = k.Key()
	colID = k.CollectionID()
	seq = k.Sequence()

	_ = colID
	_ = seq
	if err = end.Get(&key, &data, mdbx.CursorPrevNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
			goto DONE
		} else {
			cs.err = err
			return
		}
	}
	lastID = DocID(key.U64())
	colID = lastID.CollectionID()
	seq = lastID.Sequence()
	if lastID.CollectionID() != collectionID {
		count = 1
		lastID = id
	}

	min = id
	max = lastID

	count, err = mdbx.EstimateDistance(begin, end)
	if err == mdbx.ErrNotFound {
		count = 1
		err = nil
	}
	if err == mdbx.ErrSuccess {
		err = nil
	}
	count++

DONE:
	if err == mdbx.ErrSuccess {
		err = nil
	}
	if err != nil {
		cs.err = err
		return
	}
	cs.minID = min
	atomic.StoreUint64(&cs.sequence, max.Sequence())
	cs.estimated = count
	cs.loaded = true
	cs.err = nil
	return
}

func (cs *collectionStore) ensureLoaded(tx *Tx) error {
	if cs.loaded {
		return nil
	}
	var (
		err error
		end *mdbx.Cursor
	)
	end, err = tx.Tx.OpenCursor(cs.store.documentsDBI)
	if err != mdbx.ErrSuccess {
		return err
	}
	err = nil
	_, _, _, err = cs.load(tx.Docs(), end)
	return err
}

func (cs *collectionStore) Insert(
	tx *Tx,
	id DocID,
	unmarshalled interface{},
	marshalled []byte,
) error {
	var err error
	if err = cs.ensureLoaded(tx); err != nil {
		return err
	}

	if len(marshalled) == 0 {
		if marshalled, err = cs.marshaller.Marshal(unmarshalled, tx.buffer); err != nil {
			return err
		}
	}

	var (
		key    = id.Key()
		val    = mdbx.Bytes(&marshalled)
		d      = *(*string)(unsafe.Pointer(&marshalled))
		cursor = tx.Docs()
	)

	if err = cursor.Put(&key, &val, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		return err
	}
	// Insert indexes
	if len(cs.indexes) > 0 {
		tx.Index()
		tx.doc = d
		tx.docTyped = unmarshalled
		for _, index := range cs.indexes {
			if err = index.doInsert(tx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *collectionStore) Update(
	tx *Tx,
	id DocID,
	unmarshalled interface{},
	marshalled []byte,
	prev func(val mdbx.Val),
) error {
	var err error
	if err = cs.ensureLoaded(tx); err != nil {
		return err
	}
	if len(marshalled) == 0 {
		if marshalled, err = cs.marshaller.Marshal(unmarshalled, tx.buffer); err != nil {
			return err
		}
	}

	var (
		key    = id.Key()
		val    = mdbx.Bytes(&marshalled)
		d      = *(*string)(unsafe.Pointer(&marshalled))
		cursor = tx.Docs()
	)

	if err = cursor.Get(&key, &val, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		return err
	}
	err = nil
	if id != DocID(key.U64()) {
		return mdbx.ErrNotFound
	}

	if err = cursor.Put(&key, &val, 0); err != mdbx.ErrSuccess {
		return err
	}
	// Update indexes
	if len(cs.indexes) > 0 {
		tx.Index()
		tx.doc = d
		tx.docTyped = unmarshalled
		tx.prev = val.UnsafeString()
		tx.prevTyped = nil
		for _, index := range cs.indexes {
			if err = index.doUpdate(tx); err != nil {
				return err
			}
		}
	}

	if prev != nil {
		prev(val)
	}
	return nil
}

func (cs *collectionStore) Delete(
	tx *Tx,
	id DocID,
	unmarshalled interface{},
	prev func(val mdbx.Val),
) (bool, error) {
	var err error
	if err = cs.ensureLoaded(tx); err != nil {
		return false, err
	}
	var (
		key    = id.Key()
		val    = mdbx.Val{}
		cursor = tx.Docs()
	)

	if err = cursor.Get(&key, &val, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		return false, err
	}
	err = nil
	if id != DocID(key.U64()) {
		return false, mdbx.ErrNotFound
	}

	if err := cursor.Delete(0); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	// Delete indexes
	if len(cs.indexes) > 0 {
		tx.Index()
		data := val.UnsafeBytes()
		tx.doc = *(*string)(unsafe.Pointer(&data))
		tx.docTyped = unmarshalled
		for _, index := range cs.indexes {
			if err = index.doDelete(tx); err != nil {
				return false, err
			}
		}
	}

	if prev != nil {
		prev(val)
	}
	return true, nil
}
