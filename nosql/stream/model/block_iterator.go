package model

import (
	"encoding/binary"
	"io"
)

type BlockIterator interface {
	Reset(d []byte) BlockIterator

	Prev() (*Record, error)

	Next() (*Record, error)
}

func NewBlockReader(d []byte) BlockIterator {
	return &varBlockReader{
		at:   0,
		size: len(d),
		d:    d,
	}
	//if record > 0 {
	//	return &fixedBlockReader{
	//		record: record + FixedHeader,
	//		at:     0,
	//		size:   len(d),
	//		d:      d,
	//	}
	//} else {
	//
	//}
}

type varBlockReader struct {
	at   int
	size int
	d    []byte
}

func (f *varBlockReader) Reset(d []byte) BlockIterator {
	//if record == 0 {
	//	f.at = 0
	//	f.size = len(d)
	//	return f
	//}
	//return NewBlockReader(record, d)
	f.at = 0
	f.size = len(d)
	f.d = d
	return f
}

func (b *varBlockReader) Prev() (*Record, error) {
	at := b.at
	if at-BlockRecordHeaderSize < 0 {
		return nil, io.EOF
	}

	l := int(binary.LittleEndian.Uint16(b.d[at-2:]))

	if at-BlockRecordHeaderSize-l < 0 {
		return nil, ErrCorrupted
	}

	at = at - BlockRecordHeaderSize - l
	id := int64(binary.LittleEndian.Uint16(b.d[at:]))
	timestamp := int64(binary.LittleEndian.Uint16(b.d[at+8:]))
	start := int64(binary.LittleEndian.Uint16(b.d[at+16:]))
	end := int64(binary.LittleEndian.Uint16(b.d[at+24:]))
	l2 := int(binary.LittleEndian.Uint16(b.d[at+32:]))

	if l != l2 {
		return nil, ErrCorrupted
	}

	b.at = at

	return &Record{
		ID:        id,
		Timestamp: timestamp,
		Start:     start,
		End:       end,
		Data:      b.d[at+34 : at+34+l],
	}, nil
}

func (b *varBlockReader) Next() (*Record, error) {
	at := b.at
	if at+BlockRecordHeaderSize > b.size {
		return nil, io.EOF
	}

	id := int64(binary.LittleEndian.Uint16(b.d[at:]))
	timestamp := int64(binary.LittleEndian.Uint16(b.d[at+8:]))
	start := int64(binary.LittleEndian.Uint16(b.d[at+16:]))
	end := int64(binary.LittleEndian.Uint16(b.d[at+24:]))
	l := int(binary.LittleEndian.Uint16(b.d[at+32:]))

	if at+BlockRecordHeaderSize+l > b.size {
		return nil, ErrCorrupted
	}

	l2 := int(binary.LittleEndian.Uint16(b.d[at+34+l:]))

	if l != l2 {
		return nil, ErrCorrupted
	}

	b.at += BlockRecordHeaderSize + l
	return &Record{
		ID:        id,
		Timestamp: timestamp,
		Start:     start,
		End:       end,
		Data:      b.d[at+34 : at+34+l],
	}, nil
}

//// Reads the records in a block with fixed record sizes
//type fixedBlockReader struct {
//	record int
//	at     int
//	size   int
//	d      []byte
//}
//
//func (f *fixedBlockReader) Reset(record int, d []byte) BlockIterator {
//	if record > 0 && record+FixedHeader == f.record {
//		f.at = 0
//		f.size = len(d)
//		f.d = d
//		return f
//	}
//	return NewBlockReader(record, d)
//}
//
//func (b *fixedBlockReader) Prev() (*Record, error) {
//	at := b.at - b.record
//	if at < 0 {
//		return nil, io.EOF
//	}
//	b.at -= b.record
//	return &Record{
//		ID:        int64(binary.LittleEndian.Uint64(b.d[at:])),
//		Timestamp: int64(binary.LittleEndian.Uint64(b.d[at+8:])),
//		Data:      b.d[at+FixedHeader : at+b.record],
//	}, nil
//}
//
//func (b *fixedBlockReader) Next() (*Record, error) {
//	at := b.at
//	if at+b.record > b.size {
//		return nil, io.EOF
//	}
//	b.at += b.record
//	return &Record{
//		ID:        int64(binary.LittleEndian.Uint64(b.d[at:])),
//		Timestamp: int64(binary.LittleEndian.Uint64(b.d[at+8:])),
//		Data:      b.d[at+16 : at+16+b.record],
//	}, nil
//}
