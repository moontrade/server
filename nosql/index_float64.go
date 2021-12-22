package nosql

import (
	"encoding/binary"
	"github.com/moontrade/mdbx-go"
	"unsafe"
)

var (
	_ Index = (*Float64)(nil)
)

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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.IndexMeta.ID &&
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
		*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.IndexMeta.ID &&
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
