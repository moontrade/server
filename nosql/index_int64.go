package nosql

import (
	"encoding/binary"
	"github.com/moontrade/mdbx-go"
	"unsafe"
)

var (
	_ Index = (*Int64)(nil)
)

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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.IndexMeta.ID
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.IndexMeta.ID
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
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorSetRange); prevErr != mdbx.ErrSuccess {
			if prevErr == mdbx.ErrNotFound {
				prevErr = nil
			} else {
				return nil
			}
		} else {
			prevErr = nil
			keyBytes := key.UnsafeBytes()
			if key.Len == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.IndexMeta.ID &&
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.IndexMeta.ID
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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.IndexMeta.ID
	binary.BigEndian.PutUint64(tx.buffer[4:], uint64(tx.docID))
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  20,
		}
		data mdbx.Val
	)

	if err = tx.index.Get(&key, &data, mdbx.CursorSetRange); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	err = nil
	keyBytes := key.UnsafeBytes()
	if key.Len == 20 &&
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == i64.IndexMeta.ID &&
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
