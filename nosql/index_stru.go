package nosql

import (
	"bytes"
	"github.com/moontrade/mdbx-go"
	"unsafe"
)

var (
	_ Index = (*StringUnique)(nil)
)

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

	if err = tx.index.Get(&key, &data, mdbx.CursorSetRange); err != mdbx.ErrSuccess {
		if err == mdbx.ErrNotFound {
			err = nil
		} else {
			return err
		}
	} else {
		err = nil
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

		if nextErr = tx.index.Get(&key, &data, mdbx.CursorSetRange); nextErr != mdbx.ErrSuccess {
			if nextErr == mdbx.ErrNotFound {
				nextErr = nil
			} else {
				return nextErr
			}
		} else {
			nextErr = nil
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
