package nosql

import (
	"encoding/binary"
	"github.com/moontrade/mdbx-go"
	"unsafe"
)

var (
	_ Index = (*Float64Unique)(nil)
)

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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
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
			err = nil
		} else {
			return err
		}
	} else {
		err = nil
		keyBytes := key.UnsafeBytes()
		if len(keyBytes) == 20 &&
			*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.IndexMeta.ID &&
			bigEndianF64(keyBytes[12:]) == value {
			if *(*DocID)(unsafe.Pointer(&keyBytes[4])) != tx.docID {
				tx.errDocID = *(*DocID)(unsafe.Pointer(&keyBytes[4]))
				return ErrUniqueConstraint
			}
			return nil
		}
	}

	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
		*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
		binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&nextValue)))

		var (
			key = mdbx.Val{
				Base: &tx.buffer[0],
				Len:  20,
			}
			data mdbx.Val
		)

		if nextErr = tx.index.Get(&key, &data, mdbx.CursorSetRange); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrNotFound {
				nextErr = nil
			} else {
				return nextErr
			}
		} else {
			nextErr = nil
			keyBytes := key.UnsafeBytes()
			if len(keyBytes) == 20 &&
				*(*uint32)(unsafe.Pointer(&keyBytes[0])) == f64.IndexMeta.ID &&
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

		*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
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
	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = f64.IndexMeta.ID
	*(*DocID)(unsafe.Pointer(&tx.buffer[4])) = 0
	binary.BigEndian.PutUint64(tx.buffer[12:], *(*uint64)(unsafe.Pointer(&value)))

	var (
		key = mdbx.Val{
			Base: &tx.buffer[0],
			Len:  12,
		}
		data = tx.docID.Key()
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
