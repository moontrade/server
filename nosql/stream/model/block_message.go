package model

import (
	"unsafe"
)

const (
	BlockRecordHeaderSize = 36 // ID (8 bytes) | Timestamp (8 bytes) | Start (8 bytes) | End (8 bytes) | Size (2 bytes) ... 2 (bytes)
)

var (
	sizeofBlockMessageStruct = int(unsafe.Sizeof(BlockMessage{}))
	sizeofBlockMessageBuffer = int(unsafe.Sizeof(BlockMessageBuffer{}))
)

type BlockMessage struct {
	BlockHeaderMut
	Body []byte
}

func (s *BlockMessage) MessageID() int64 {
	return s.BlockHeaderMut.Min()
}

func (s *BlockMessage) Timestamp() int64 {
	return s.BlockHeaderMut.Created()
}

func (b *BlockMessage) Clone() *BlockMessage {
	clone := GetBlockMessage(len(b.Body))
	clone.BlockHeaderMut = b.BlockHeaderMut
	return clone
}

func (b *BlockMessage) Unmarshal(header []byte, data []byte) error {
	if err := b.BlockHeaderMut.UnmarshalBinary(header); err != nil {
		return err
	}
	if len(b.Body) < len(data) {
		if cap(b.Body) >= len(data) {
			b.Body = b.Body[0:len(data)]
		}
		b.Body = make([]byte, len(data))
	} else if len(b.Body) > len(data) {
		b.Body = b.Body[0:len(data)]
	}
	copy(b.Body, data)
	return nil
}

func (b *BlockMessage) Sizeof() int {
	return sizeofBlockMessageStruct + 24 + len(b.Body)
}

func (b *BlockMessage) Header() *BlockHeader {
	return &b.BlockHeader
}

func (b *BlockMessage) HeaderMut() *BlockHeaderMut {
	return &b.BlockHeaderMut
}

func (b *BlockMessage) Data() []byte {
	return b.Body[0:b.Size()]
}

func (b *BlockMessage) Type() MessageType {
	return MessageType_Block
}

func (b *BlockMessage) MarshalTo(w *Writer) error {
	if int(b.Size()) != len(b.Body) {
		return ErrSizeDataMismatch
	}
	if err := w.Write(b.BlockHeader.Bytes()); err != nil {
		return err
	}
	return w.Write(b.Body)
}

func (b *BlockMessage) UnmarshalFrom(rd *Reader) error {
	if err := rd.ReadBytes(b.BlockHeader.Bytes()); err != nil {
		return err
	}
	if b.BlockHeader.Size() == 0 {
		return nil
	}

	if len(b.Body) != int(b.BlockHeader.Size()) {
		b.Body = make([]byte, b.BlockHeader.Size())
	}

	return rd.ReadBytes(b.Body)
}

func (b *BlockMessage) UnmarshalCompactFrom(rd *Reader) error {
	if err := b.BlockHeader.UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	if b.BlockHeader.Size() == 0 {
		if b.Body != nil {
			b.Body = nil
		}
		return nil
	}
	if b.Body == nil {
		b.Body = make([]byte, b.BlockHeader.Size())
	} else if cap(b.Body) < int(b.BlockHeader.Size()) {
		b.Body = make([]byte, b.BlockHeader.Size())
	} else {
		b.Body = b.Body[0:b.BlockHeader.Size()]
	}
	return rd.ReadBytes(b.Body)
}

func (b *BlockMessage) MarshalCompactTo(w *Writer) error {
	err := b.BlockHeader.MarshalCompactTo(w)
	if err != nil {
		return err
	}
	if int(b.BlockHeader.Size()) != len(b.Body) {
		return ErrSizeDataMismatch
	}
	return w.Write(b.Body)
}
