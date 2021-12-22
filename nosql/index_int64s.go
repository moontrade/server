package nosql

import (
	"encoding/binary"
	"github.com/moontrade/mdbx-go"
	"sort"
	"unsafe"
)

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

	*(*uint32)(unsafe.Pointer(&tx.buffer[0])) = i64.IndexMeta.ID
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
