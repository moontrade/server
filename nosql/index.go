package nosql

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/moontrade/mdbx-go"
	"sort"
	"unsafe"
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

var (
	_ Index = (*Int64)(nil)
	_ Index = (*Int64Unique)(nil)
	_ Index = (*Float64)(nil)
	_ Index = (*Float64Unique)(nil)
	_ Index = (*String)(nil)
	_ Index = (*StringUnique)(nil)
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

func (i64 *Int64) doInsert(tx *Tx) error {
	var (
		value, err = i64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Put(&key, &data, 0); err != mdbx.ErrSuccess {
		return err
	} else {
		err = nil
	}

	return nil
}

func (i64 *Int64) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return i64.doInsert(tx)
	}

	var (
		prevValue, prevErr = i64.ValueOf(tx.prev, tx.prevTyped)
		prevSkip           = prevErr == ErrSkip
		nextValue, nextErr = i64.ValueOf(tx.doc, tx.docTyped)
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	// Previous value?
	if !prevSkip {
		// Did value change?
		if !nextSkip && prevValue == nextValue {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&prevValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
				DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
				bigEndianI64(keyBytes[12:]) == prevValue {
				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Put(&key, &data, 0); nextErr != mdbx.ErrSuccess {
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (i64 *Int64) doDelete(tx *Tx) error {
	var (
		value, err = i64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == 20 &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
		DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
		bigEndianI64(keyBytes[12:]) == value {
		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
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

func (i64 *Int64Unique) doInsert(tx *Tx) error {
	var (
		value, err = i64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
		} else {
			return err
		}
	} else {
		keyBytes := key.UnsafeBytes()
		if len(keyBytes) == 20 &&
			*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
			bigEndianI64(keyBytes[12:]) == value {
			if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
				tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
				return ErrUniqueConstraint
			}
			return nil
		}
	}

	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))
	key = mdbx.Val{
		Base: &tx.buffer[0],
		Len:  20,
	}
	data = mdbx.Val{}

	if err = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		if err == mdbx.ErrKeyExist {
			return nil
		}
		return err
	} else {
		err = nil
	}

	return nil
}

func (i64 *Int64Unique) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return i64.doInsert(tx)
	}

	var (
		prevValue, prevErr = i64.ValueOf(tx.prev, tx.prevTyped)
		prevSkip           = prevErr == ErrSkip
		nextValue, nextErr = i64.ValueOf(tx.doc, tx.docTyped)
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	// Previous value?
	if !prevSkip {
		// Did values change?
		if !nextSkip && prevValue == nextValue {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&prevValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
				bigEndianI64(keyBytes[12:]) == prevValue {
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					return ErrUniqueConstraint
				}

				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrNotFound {
				nextErr = nil
			} else {
				return nextErr
			}
		} else {
			keyBytes := key.UnsafeBytes()
			if len(keyBytes) == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
				bigEndianI64(keyBytes[12:]) == nextValue {
				// UniqueConstraint?
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
					return ErrUniqueConstraint
				}
				// Key already exists
				return nil
			}
		}

		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data = mdbx.Val{}

		if nextErr = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrKeyExist {
				return nil
			}
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (i64 *Int64Unique) doDelete(tx *Tx) error {
	var (
		value, err = i64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  12,
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == 20 &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.meta.ID &&
		bigEndianI64(keyBytes[12:]) == value {
		if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
			return ErrUniqueConstraint
		}
		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
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

func (i64 *Int64Array) doInsert(tx *Tx) error {
	var (
		values, err = i64.ValueOf(tx.doc, tx.docTyped, tx.i64)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	if len(values) == 0 {
		return nil
	}

	sort.Sort(int64Slice(values))

	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID

	for _, value := range values {
		// Set key to next value
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if err = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
			if err == mdbx.ErrKeyExist {
				return ErrUniqueConstraint
			}
			return err
		} else {
			err = nil
		}
	}

	return nil
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

func (f64 *Float64) doInsert(tx *Tx) error {
	var (
		value, err = f64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Put(&key, &data, 0); err != mdbx.ErrSuccess {
		return err
	} else {
		err = nil
	}

	return nil
}

func (f64 *Float64) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return f64.doInsert(tx)
	}

	var (
		prevValue, prevErr = f64.ValueOf(tx.prev, tx.prevTyped)
		prevSkip           = prevErr == ErrSkip
		nextValue, nextErr = f64.ValueOf(tx.doc, tx.docTyped)
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	// Previous value?
	if !prevSkip {
		// Did values change?
		if !nextSkip && prevValue == nextValue {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&prevValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
				DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
				bigEndianF64(keyBytes[12:]) == prevValue {
				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Put(&key, &data, 0); nextErr != mdbx.ErrSuccess {
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (f64 *Float64) doDelete(tx *Tx) error {
	var (
		value, err = f64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == 20 &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
		DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
		bigEndianF64(keyBytes[12:]) == value {
		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
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

func (f64 *Float64Unique) doInsert(tx *Tx) error {
	var (
		value, err = f64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
		} else {
			return err
		}
	} else {
		keyBytes := key.UnsafeBytes()
		if len(keyBytes) == 20 &&
			*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
			bigEndianF64(keyBytes[12:]) == value {
			if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
				tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
				return ErrUniqueConstraint
			}
			return nil
		}
	}

	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))
	key = mdbx.Val{
		Base: &tx.buffer[0],
		Len:  20,
	}
	data = mdbx.Val{}

	if err = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		if err == mdbx.ErrKeyExist {
			return nil
		}
		return err
	} else {
		err = nil
	}

	return nil
}

func (f64 *Float64Unique) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return f64.doInsert(tx)
	}

	var (
		prevValue, prevErr = f64.ValueOf(tx.prev, tx.prevTyped)
		prevSkip           = prevErr == ErrSkip
		nextValue, nextErr = f64.ValueOf(tx.doc, tx.docTyped)
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	// Previous value?
	if !prevSkip {
		// Did value change?
		if !nextSkip && prevValue == nextValue {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&prevValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
				bigEndianF64(keyBytes[12:]) == prevValue {
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					return ErrUniqueConstraint
				}

				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrNotFound {
				nextErr = nil
			} else {
				return nextErr
			}
		} else {
			keyBytes := key.UnsafeBytes()
			if len(keyBytes) == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
				bigEndianF64(keyBytes[12:]) == nextValue {
				// UniqueConstraint?
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
					return ErrUniqueConstraint
				}
				// Key already exists
				return nil
			}
		}

		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data = mdbx.Val{}

		if nextErr = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrKeyExist {
				return nil
			}
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (f64 *Float64Unique) doDelete(tx *Tx) error {
	var (
		value, err = f64.ValueOf(tx.doc, tx.docTyped)
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  12,
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == 20 &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.meta.ID &&
		bigEndianF64(keyBytes[12:]) == value {
		if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
			return ErrUniqueConstraint
		}
		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Float64Array
////////////////////////////////////////////////////////////////////////////////////////

type Float64ArrayValueOf func(data string, unmarshalled interface{}, into []float64) ([]float64, error)

type Float64Array struct {
	indexBase
	ValueOf Float64ArrayValueOf
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

func (str *String) doInsert(tx *Tx) error {
	var (
		value, err = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[12:])
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  uint64(12 + len(value)),
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Put(&key, &data, 0); err != mdbx.ErrSuccess {
		return err
	} else {
		err = nil
	}

	return nil
}

func (str *String) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return str.doInsert(tx)
	}

	var (
		prevValue, prevErr = str.ValueOf(tx.prev, tx.prevTyped, tx.buffer[12:])
		prevSkip           = prevErr == ErrSkip
		nextOffset         = 12 + len(prevValue)
	)
	var (
		nextValue, nextErr = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[nextOffset+12:])
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	if len(prevValue) > MaxIndexKeySize || len(nextValue) > MaxIndexKeySize {
		return ErrIndexKeyTooBig
	}
	if len(tx.buffer) < 48+len(prevValue)+len(nextValue) {
		return ErrIndexKeyTooBig
	}

	// Previous value?
	if !prevSkip {
		// Did values change?
		if !nextSkip && bytes.Equal(prevValue, nextValue) {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))

		var (
			keyLen = uint64(12 + len(prevValue))
			key    = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  keyLen,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == keyLen &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
				DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
				bytes.Equal(keyBytes[12:], prevValue) {
				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[nextOffset])) = str.meta.ID
		binary.BigEndian.PutUint64(tx.buffer[nextOffset+4:], uint64(tx.docID))

		var (
			keyLen = uint64(12 + len(nextValue))
			key    = mdbx.Val{
				Base: &tx.buffer[nextOffset],
				Len:  keyLen,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Put(&key, &data, 0); nextErr != mdbx.ErrSuccess {
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (str *String) doDelete(tx *Tx) error {
	var (
		value, err = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[12:])
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))

	var (
		keyLen = uint64(12 + len(value))
		key    = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  keyLen,
		}
		data = tx.docID.Key()
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == keyLen &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
		DocID(binary.BigEndian.Uint64(keyBytes[4:])) == tx.docID &&
		bytes.Equal(keyBytes[12:], tx.buffer[12:]) {
		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
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

func (str *StringUnique) doInsert(tx *Tx) error {
	var (
		value, err = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[12:])
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}
	if len(value) > MaxIndexKeySize {
		return ErrIndexKeyTooBig
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0

	var (
		keyLen = uint64(12 + len(value))
		key    = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  keyLen,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
		} else {
			return err
		}
	} else {
		keyBytes := key.UnsafeBytes()
		if key.Len == keyLen &&
			*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
			bytes.Equal(keyBytes[12:], value) {
			if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
				tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
				return ErrUniqueConstraint
			}
			return nil
		}
	}

	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = tx.docID
	key = mdbx.Val{
		Base: &tx.buffer[0],
		Len:  keyLen,
	}
	data = mdbx.Val{}

	if err = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); err != mdbx.ErrSuccess {
		if err == mdbx.ErrKeyExist {
			return nil
		}
		return err
	} else {
		err = nil
	}

	return nil
}

func (str *StringUnique) doUpdate(tx *Tx) error {
	if len(tx.prev) == 0 {
		return str.doInsert(tx)
	}

	var (
		prevValue, prevErr = str.ValueOf(tx.prev, tx.prevTyped, tx.buffer[12:])
		prevSkip           = prevErr == ErrSkip
		nextOffset         = 12 + len(prevValue)
	)
	var (
		nextValue, nextErr = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[nextOffset+12:])
		nextSkip           = nextErr == ErrSkip
	)

	if prevSkip {
		prevErr = nil
	}
	if nextSkip {
		if prevSkip {
			return nil
		}
		nextErr = nil
	}

	if nextErr != nil {
		return nextErr
	}
	if prevErr != nil {
		return prevErr
	}

	if len(prevValue) > MaxIndexKeySize || len(nextValue) > MaxIndexKeySize {
		return ErrIndexKeyTooBig
	}
	if len(tx.buffer) < 48+len(prevValue)+len(nextValue) {
		return ErrIndexKeyTooBig
	}

	// Previous value?
	if !prevSkip {
		// Did values change?
		if !nextSkip && bytes.Equal(prevValue, nextValue) {
			return nil
		}

		// Set key to existing value
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0

		var (
			keyLen = uint64(12 + len(prevValue))
			key    = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  keyLen,
			}
			data mdbx.Val
		)

		// Find entry of previous value.
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == keyLen &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
				bytes.Equal(keyBytes[12:], prevValue) {
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					return ErrUniqueConstraint
				}

				if prevErr = tx.index.Delete(0); prevErr != mdbx.ErrSuccess {
					return prevErr
				} else {
					prevErr = nil
				}
			}
		}
	}

	if !nextSkip {
		// Set key to next value
		*(*uint32)(unsafe.Pointer(&tx.buffer[nextOffset])) = str.meta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[nextOffset+4])) = 0

		var (
			keyLen = uint64(12 + len(nextValue))
			key    = mdbx.Val{
				Base: &tx.buffer[nextOffset],
				Len:  keyLen,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrNotFound {
				nextErr = nil
			} else {
				return nextErr
			}
		} else {
			keyBytes := key.UnsafeBytes()
			if key.Len == keyLen &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
				bytes.Equal(keyBytes[12:], nextValue) {
				// UniqueConstraint?
				if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
					tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
					return ErrUniqueConstraint
				}
				// Key already exists
				return nil
			}
		}

		key = mdbx.Val{
			Base: &tx.buffer[nextOffset],
			Len:  keyLen,
		}
		data = mdbx.Val{}

		if nextErr = tx.index.Put(&key, &data, mdbx.PutNoOverwrite); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrKeyExist {
				return nil
			}
			return nextErr
		} else {
			nextErr = nil
		}
	}

	return nil
}

func (str *StringUnique) doDelete(tx *Tx) error {
	var (
		value, err = str.ValueOf(tx.doc, tx.docTyped, tx.buffer[12:])
	)
	if err != nil {
		if err == ErrSkip {
			return err
		}
	}

	// Set key to next value
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = str.meta.ID

	var (
		keyLen = uint64(12 + len(value))
		key    = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  keyLen,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorNextNoDup); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	keyBytes := key.UnsafeBytes()
	if key.Len == keyLen &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == str.meta.ID &&
		bytes.Equal(keyBytes[12:], value) {
		if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
			tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
			return ErrUniqueConstraint
		}

		if err = tx.index.Delete(0); err != mdbx.ErrSuccess {
			return err
		} else {
			err = nil
		}
	}
	return nil
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

////////////////////////////////////////////////////////////////////////////////////////
// FullText
////////////////////////////////////////////////////////////////////////////////////////

type FullText struct{}
