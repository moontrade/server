package model

import (
	"encoding/binary"
	"io"
	"math"
)

type Reader struct {
	b []byte
	i int
}

func ReaderOf(b []byte) Reader {
	return Reader{
		b: b,
	}
}

func NewReader(b []byte) *Reader {
	return &Reader{
		b: b,
		i: 0,
	}
}

func (r *Reader) Reset(b []byte) {
	r.b = b
	r.i = 0
}

func (r *Reader) Remaining() int {
	return len(r.b) - r.i
}

func (r *Reader) At() int {
	return r.i
}

func (r *Reader) ReadBool() (bool, error) {
	if r.Remaining() < 1 {
		return false, io.ErrShortBuffer
	}
	v := r.b[r.i]
	r.i += 1
	return v > 0, nil
}

func (r *Reader) ReadByte() (byte, error) {
	if r.Remaining() < 1 {
		return 0, io.ErrShortBuffer
	}
	v := r.b[r.i]
	r.i += 1
	return v, nil
}

func (r *Reader) ReadBytes(b []byte) error {
	if r.Remaining() < len(b) {
		return io.ErrShortBuffer
	}
	copy(b, r.b[r.i:])
	r.i += len(b)
	return nil
}

func (r *Reader) ReadBytesUnsafe(size int) ([]byte, error) {
	if r.Remaining() < size {
		return nil, io.ErrShortBuffer
	}
	b := r.b[r.i : r.i+size]
	r.i += size
	return b, nil
}

func (r *Reader) ReadInt16() (int16, error) {
	if r.Remaining() < 2 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint32(r.b[r.i:])
	r.i += 2
	return int16(v), nil
}

func (r *Reader) ReadUint16() (uint16, error) {
	if r.Remaining() < 2 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint64(r.b[r.i:])
	r.i += 2
	return uint16(v), nil
}

func (r *Reader) ReadInt32() (int32, error) {
	if r.Remaining() < 4 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint32(r.b[r.i:])
	r.i += 4
	return int32(v), nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	if r.Remaining() < 4 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint64(r.b[r.i:])
	r.i += 4
	return uint32(v), nil
}

func (r *Reader) ReadInt64() (int64, error) {
	if r.Remaining() < 8 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint64(r.b[r.i:])
	r.i += 8
	return int64(v), nil
}

func (r *Reader) ReadUint64() (uint64, error) {
	if r.Remaining() < 8 {
		return 0, io.ErrShortBuffer
	}
	v := binary.LittleEndian.Uint64(r.b[r.i:])
	r.i += 8
	return v, nil
}

func (r *Reader) ReadVarint16() (int16, int, error) {
	v, n, err := r.ReadVarint64()
	return int16(v), n, err
}

func (r *Reader) ReadUVarint16() (uint16, int, error) {
	v, n, err := r.ReadUVarint64()
	return uint16(v), n, err
}

func (r *Reader) ReadVarint32() (int32, int, error) {
	v, n, err := r.ReadVarint64()
	return int32(v), n, err
}

func (r *Reader) ReadUVarint32() (uint32, int, error) {
	v, n, err := r.ReadUVarint64()
	return uint32(v), n, err
}

func (r *Reader) ReadVarintZigZag64() (int64, int, error) {
	x, n, err := r.ReadUVarint64()
	if err != nil {
		return int64(x), n, err
	}
	return int64(x>>1) ^ int64(x)<<63>>63, n, nil
}

func (r *Reader) ReadVarint64() (int64, int, error) {
	ux, n, err := r.ReadUVarint64()
	if err != nil {
		return int64(ux), n, err
	}
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, n, nil
}

func (r *Reader) ReadUVarint64() (uint64, int, error) {
	var (
		y   uint64
		v   uint64
		err error
		b   byte
	)

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}

	v = uint64(b)
	if v < 0x80 {
		return v, 1, nil
	}
	v -= 0x80

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 7
	if y < 0x80 {
		return v, 2, nil
	}
	v -= 0x80 << 7

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 14
	if y < 0x80 {
		return v, 3, nil
	}
	v -= 0x80 << 14

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 21
	if y < 0x80 {
		return v, 4, nil
	}
	v -= 0x80 << 21

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 28
	if y < 0x80 {
		return v, 5, nil
	}
	v -= 0x80 << 28

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 35
	if y < 0x80 {
		return v, 6, nil
	}
	v -= 0x80 << 35

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 42
	if y < 0x80 {
		return v, 7, nil
	}
	v -= 0x80 << 42

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 49
	if y < 0x80 {
		return v, 8, nil
	}
	v -= 0x80 << 49

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 56
	if y < 0x80 {
		return v, 9, nil
	}
	v -= 0x80 << 56

	if b, err = r.ReadByte(); err != nil {
		return 0, 0, err
	}
	y = uint64(b)
	v += y << 63
	if y < 2 {
		return v, 10, nil
	}
	return 0, 0, ErrVarintOverflow
}

// ReadTag decodes the field Number and wire Type from its unified form.
// The Number is -1 if the decoded field number overflows int32.
// Other than overflow, this does not check for field number validity.
func (r *Reader) ReadTag(x uint64) (WireNumber, WireType, error) {
	x, _, err := r.ReadUVarint64()
	if err != nil {
		return 0, 0, err
	}
	// NOTE: MessageSet allows for larger field numbers than normal.
	if x>>3 > uint64(math.MaxInt32) {
		return -1, 0, ErrFieldTagTooBig
	}
	return WireNumber(x >> 3), WireType(x & 7), nil
}
