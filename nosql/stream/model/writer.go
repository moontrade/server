package model

import (
	"encoding/binary"
	"errors"
)

type Writer struct {
	b        []byte
	i        int
	lastSize int
}

func (w *Writer) Clone() Writer {
	clone := Writer{
		i:        w.i,
		lastSize: w.lastSize,
	}
	if len(w.b) > 0 {
		clone.b = make([]byte, len(w.b))
		//clone.b = pbytes.GetLen(len(w.b))
		copy(clone.b, w.b)
	}
	return clone
}

func (w *Writer) At() int {
	return w.i
}

func (w *Writer) Reset() {
	w.lastSize = w.i
	w.i = 0
}

func (w *Writer) Take() []byte {
	if w.b == nil {
		return nil
	}
	b := w.b[0:w.i]
	w.b = nil
	w.i = 0
	w.lastSize = len(b)
	return b
}

func (w *Writer) Alloc() {
	if w.lastSize > 0 {
		w.b = make([]byte, w.lastSize)
		//w.b = pbytes.GetLen(w.lastSize)
	} else {
		w.b = make([]byte, 128)
		//w.b = pbytes.GetLen(128)
	}
}

func (w *Writer) Close() error {
	if w.b != nil {
		//pbytes.Put(w.b)
		w.b = nil
		w.i = 0
	}
	return nil
}

func (w *Writer) Size() int {
	return w.i
}

func (w *Writer) Remaining() int {
	return len(w.b) - w.i
}

func (w *Writer) Ensure(n int) error {
	if w.Remaining() >= n {
		return nil
	}
	if cap(w.b)-w.i >= n {
		w.b = w.b[0:cap(w.b)]
	}
	b := make([]byte, len(w.b)*2)
	//b := pbytes.GetLen(len(w.b) * 2)
	if len(b) == 0 {
		return errors.New("out of memory")
	}
	copy(b, w.b)
	//pbytes.Put(w.b)
	w.b = b
	return nil
}

func (w *Writer) WriteByte(value byte) error {
	if err := w.Ensure(1); err != nil {
		panic(err)
	}
	w.b[w.i] = value
	w.i += 1
	return nil
}

func (w *Writer) Write(value []byte) error {
	if len(value) == 0 {
		return nil
	}
	if err := w.Ensure(len(value)); err != nil {
		panic(err)
	}
	copy(w.b[w.i:], value)
	w.i += len(value)
	return nil
}

func (w *Writer) WriteString(s string) error {
	if len(s) == 0 {
		return nil
	}
	if err := w.Ensure(len(s)); err != nil {
		return err
	}
	copy(w.b[w.i:], s)
	w.i += len(s)
	return nil
}

func (w *Writer) WriteBool(value bool) error {
	if err := w.Ensure(1); err != nil {
		return err
	}
	if value {
		w.b[w.i] = 1
	} else {
		w.b[w.i] = 0
	}
	w.i += 1
	return nil
}

func (w *Writer) WriteInt16(value int16) error {
	if err := w.Ensure(2); err != nil {
		return err
	}
	binary.LittleEndian.PutUint16(w.b[w.i:], uint16(value))
	w.i += 2
	return nil
}

func (w *Writer) WriteUint16(value uint16) error {
	if err := w.Ensure(2); err != nil {
		return err
	}
	binary.LittleEndian.PutUint16(w.b[w.i:], value)
	w.i += 2
	return nil
}

func (w *Writer) WriteInt32(value int32) error {
	if err := w.Ensure(4); err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(w.b[w.i:], uint32(value))
	w.i += 4
	return nil
}

func (w *Writer) WriteUint32(value uint32) error {
	if err := w.Ensure(4); err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(w.b[w.i:], value)
	w.i += 4
	return nil
}

func (w *Writer) WriteInt64(value int64) error {
	if err := w.Ensure(8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(w.b[w.i:], uint64(value))
	w.i += 8
	return nil
}

func (w *Writer) WriteUint64(value uint64) error {
	if err := w.Ensure(8); err != nil {
		return nil
	}
	binary.LittleEndian.PutUint64(w.b[w.i:], value)
	w.i += 8
	return nil
}

func (w *Writer) WriteVarint16(value int16) (int, error) {
	return w.WriteVarint64(int64(value))
}

func (w *Writer) WriteUVarint16(value uint16) (int, error) {
	return w.WriteUVarint64(uint64(value))
}

func (w *Writer) WriteVarint32(value int32) (int, error) {
	return w.WriteVarint64(int64(value))
}

func (w *Writer) WriteUVarint32(value uint32) (int, error) {
	return w.WriteUVarint64(uint64(value))
}

func (w *Writer) WriteVarintZigZag64(value int64) (int, error) {
	return w.WriteUVarint64(uint64(value<<1) ^ uint64(value>>63))
}

func (w *Writer) WriteVarint64(value int64) (int, error) {
	ux := uint64(value) << 1
	if value < 0 {
		ux = ^ux
	}
	return w.WriteUVarint64(ux)
}

func (w *Writer) WriteUVarint64(value uint64) (int, error) {
	switch {
	case value < 1<<7:
		if err := w.Ensure(1); err != nil {
			return 0, err
		}
		w.b[w.i] = byte(value)
		w.i += 1
		return 1, nil

	case value < 1<<14:
		if err := w.Ensure(2); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte(value >> 7)
		w.i += 2
		return 2, nil

	case value < 1<<21:
		if err := w.Ensure(3); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte(value >> 14)
		w.i += 3
		return 3, nil

	case value < 1<<28:
		if err := w.Ensure(4); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte(value >> 21)
		w.i += 4
		return 4, nil

	case value < 1<<35:
		if err := w.Ensure(5); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte(value >> 28)
		w.i += 5
		return 5, nil

	case value < 1<<42:
		if err := w.Ensure(6); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte((value>>28)&0x7f | 0x80)
		w.b[w.i+5] = byte(value >> 35)
		w.i += 6
		return 6, nil

	case value < 1<<49:
		if err := w.Ensure(7); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte((value>>28)&0x7f | 0x80)
		w.b[w.i+5] = byte((value>>35)&0x7f | 0x80)
		w.b[w.i+6] = byte(value >> 42)
		w.i += 7
		return 7, nil

	case value < 1<<56:
		if err := w.Ensure(8); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte((value>>28)&0x7f | 0x80)
		w.b[w.i+5] = byte((value>>35)&0x7f | 0x80)
		w.b[w.i+6] = byte((value>>42)&0x7f | 0x80)
		w.b[w.i+7] = byte(value >> 49)
		w.i += 8
		return 8, nil

	case value < 1<<63:
		if err := w.Ensure(9); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte((value>>28)&0x7f | 0x80)
		w.b[w.i+5] = byte((value>>35)&0x7f | 0x80)
		w.b[w.i+6] = byte((value>>42)&0x7f | 0x80)
		w.b[w.i+7] = byte((value>>49)&0x7f | 0x80)
		w.b[w.i+8] = byte(value >> 56)
		w.i += 9
		return 9, nil

	default:
		if err := w.Ensure(10); err != nil {
			return 0, err
		}
		w.b[w.i] = byte((value>>0)&0x7f | 0x80)
		w.b[w.i+1] = byte((value>>7)&0x7f | 0x80)
		w.b[w.i+2] = byte((value>>14)&0x7f | 0x80)
		w.b[w.i+3] = byte((value>>21)&0x7f | 0x80)
		w.b[w.i+4] = byte((value>>28)&0x7f | 0x80)
		w.b[w.i+5] = byte((value>>35)&0x7f | 0x80)
		w.b[w.i+6] = byte((value>>42)&0x7f | 0x80)
		w.b[w.i+7] = byte((value>>49)&0x7f | 0x80)
		w.b[w.i+8] = byte((value>>56)&0x7f | 0x80)
		w.b[w.i+9] = 1
		w.i += 10
		return 10, nil
	}
}

// WriteTag encodes the field Number and wire Type into its unified form.
func (w *Writer) WriteTag(num WireNumber, typ WireType) (int, error) {
	return w.WriteUVarint64(uint64(num)<<3 | uint64(typ&7))
}
