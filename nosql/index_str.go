package nosql

import (
	"bytes"
	"encoding/binary"
	"github.com/moontrade/mdbx-go"
	"unsafe"
)

var (
	_ Index = (*String)(nil)
)

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
		if prevErr = tx.index.Get(&key, &data, mdbx.CursorSetRange); prevErr != mdbx.ErrSuccess {
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

	if err = tx.index.Get(&key, &data, mdbx.CursorSetRange); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			return nil
		}
		return err
	}

	err = nil
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
