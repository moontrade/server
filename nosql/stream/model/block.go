package model

import (
	"encoding/binary"
	"github.com/pierrec/lz4/v4"
	"sync"
)

var (
	Block1MaxDataSize  = (1024 - SizeofBlockHeader) - (lz4.CompressBlockBound(1024-SizeofBlockHeader) + 1)
	Block2MaxDataSize  = (2048 - SizeofBlockHeader) - (lz4.CompressBlockBound(2048-SizeofBlockHeader) + 1)
	Block4MaxDataSize  = (4096 - SizeofBlockHeader) - (lz4.CompressBlockBound(4096-SizeofBlockHeader) + 1)
	Block8MaxDataSize  = (8192 - SizeofBlockHeader) - (lz4.CompressBlockBound(8192-SizeofBlockHeader) + 1)
	Block16MaxDataSize = (16384 - SizeofBlockHeader) - (lz4.CompressBlockBound(16384-SizeofBlockHeader) + 1)
	Block32MaxDataSize = (32768 - SizeofBlockHeader) - (lz4.CompressBlockBound(32768-SizeofBlockHeader) + 1)
	Block64MaxDataSize = (65536 - SizeofBlockHeader) - (lz4.CompressBlockBound(65536-SizeofBlockHeader) + 1)

	block1Pool = &sync.Pool{New: func() interface{} {
		return &Block1Mut{}
	}}
	block2Pool = &sync.Pool{New: func() interface{} {
		return &Block2Mut{}
	}}
	block4Pool = &sync.Pool{New: func() interface{} {
		return &Block4Mut{}
	}}
	block8Pool = &sync.Pool{New: func() interface{} {
		return &Block8Mut{}
	}}
	block16Pool = &sync.Pool{New: func() interface{} {
		return &Block16Mut{}
	}}
	block32Pool = &sync.Pool{New: func() interface{} {
		return &Block32Mut{}
	}}
	block64Pool = &sync.Pool{New: func() interface{} {
		return &Block64Mut{}
	}}
)

const (
	MinCompressSize = 128
)

// BlockAllocator allocates a new Block
type BlockAllocator interface {
	Alloc() BlockMut

	Release(b BlockMut)
}

type Block interface {
	Message
	//Type() MessageType

	Header() *BlockHeader

	Data() []byte
}

type BlockMut interface {
	Block

	HeaderMut() *BlockHeaderMut

	// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
	Uncompress(b []byte) error

	AppendRecord(id, created, time, end int64, data []byte) error
}

func NewBlockID(streamID, id int64) *BlockID {
	return (&BlockIDMut{}).SetStreamID(streamID).SetId(id).Freeze()
}

func (b1 *Block1) Type() MessageType {
	return MessageType_Block
}

func (b *Block1) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block1) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block1) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block1) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (b1 *Block1) Header() *BlockHeader {
	return b1.Head()
}

func (b1 *Block1Mut) HeaderMut() *BlockHeaderMut {
	return b1.Head()
}

func (bm *Block1) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func (*Block2) Type() MessageType {
	return MessageType_Block
}

func (b *Block2) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block2) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block2) Header() *BlockHeader {
	return b.Head()
}

func (b *Block2Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (b *Block2) Data() []byte {
	return b.Body()[0:b.Head().Size()]
}

func (b *Block2) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block2) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (*Block4) Type() MessageType {
	return MessageType_Block
}

func (b *Block4) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block4) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block4) Header() *BlockHeader {
	return b.Head()
}

func (b *Block4Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (bm *Block4) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func (b *Block4) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block4) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (*Block8) Type() MessageType {
	return MessageType_Block
}

func (b *Block8) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block8) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block8) Header() *BlockHeader {
	return b.Head()
}

func (b *Block8Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (b *Block8) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block8) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (*Block16) Type() MessageType {
	return MessageType_Block
}

func (b *Block16) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block16) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block16) Header() *BlockHeader {
	return b.Head()
}

func (b *Block16Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (b *Block16) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block16) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (*Block32) Type() MessageType {
	return MessageType_Block
}

func (b *Block32) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block32) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block32) Header() *BlockHeader {
	return b.Head()
}

func (b *Block32Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (b *Block32) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block32) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (*Block64) Type() MessageType {
	return MessageType_Block
}

func (b *Block64) MessageID() int64 {
	return b.Head().Min()
}

func (b *Block64) Timestamp() int64 {
	return b.Head().Created()
}

func (b *Block64) Header() *BlockHeader {
	return b.Head()
}

func (b *Block64Mut) HeaderMut() *BlockHeaderMut {
	return b.Head()
}

func (b *Block64) MarshalCompactTo(w *Writer) error {
	var err error
	if err = b.Head().MarshalCompactTo(w); err != nil {
		return err
	}
	data := b.Data()
	if len(data) != int(b.Head().Size()) {
		return ErrCorrupted
	}
	return w.Write(data)
}

func (b *Block64) UnmarshalCompactFrom(rd *Reader) error {
	var err error
	if err = b.Head().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	return rd.ReadBytes(b.body[0:b.Head().Size()])
}

func (bm *Block8) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func (bm *Block16) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func (bm *Block32) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func (bm *Block64) Data() []byte {
	return bm.Body()[0:bm.Head().Size()]
}

func ToBlockMessage(b Block, compress bool) (*BlockMessage, error) {
	bdata := b.Data()
	size := len(bdata)
	if size == 0 {
		// Empty BlockMessage
		return &BlockMessage{
			BlockHeaderMut: *b.Header().Mut(),
		}, nil
	}

	//
	if size < MinCompressSize || !compress || b.Header().Compression() != Compression_None {
		m := GetBlockMessage(size)
		if m == nil || m.Body == nil {
			return nil, ErrOutOfMemory
		}
		m.Body = m.Body[0:size]
		copy(m.Body, bdata)
		return m, nil
	}

	c := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(bdata, c.buf[:65536], c.ht[:65536])
	lz4Pool.Put(c)
	if err != nil {
		return nil, err
	}

	// Was compression worth it?
	if n >= len(bdata) {
		m := GetBlockMessage(len(bdata))
		if m == nil || m.Body == nil {
			return nil, ErrOutOfMemory
		}
		// Copy data uncompressed
		m.Body = m.Body[0:len(bdata)]
		copy(m.Body, bdata)
		// Copy header as is
		m.BlockHeaderMut = *b.Header().Mut()
		return m, nil
	}

	m := GetBlockMessage(n)
	if m == nil || m.Body == nil {
		return nil, ErrOutOfMemory
	}
	copy(m.Body, c.buf[0:n])
	// Modify header to reflect compression
	m.BlockHeaderMut = *b.Header().Mut()
	m.BlockHeaderMut.SetCompression(Compression_LZ4)
	m.BlockHeaderMut.SetSizeX(uint16(n))
	m.BlockHeaderMut.SetSize(m.BlockHeaderMut.SizeX())
	m.BlockHeaderMut.SetSizeU(uint16(len(bdata)))
	return m, nil
}

func (bm *Block1Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block2Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block4Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block8Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block16Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block32Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

func (bm *Block64Mut) ToMessage(compress bool) (*BlockMessage, error) {
	return ToBlockMessage(bm, compress)
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block1Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block2Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block4Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block8Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block16Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block32Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Uncompress decompresses the LZ4 encoded input into the memory of BlockMut
func (bm *Block64Mut) Uncompress(b []byte) error {
	_, err := lz4.UncompressBlock(b, bm.Bytes())
	return err
}

// Append appends the next record into the block.
func appendRecord(header *BlockHeaderMut, block []byte, id, timestamp, start, end int64, maxSize int, data []byte) error {
	if id <= header.Max() {
		return ErrIDTooSmall
	}
	if start < header.End() || end < header.End() {
		return ErrTimeIsPast
	}

	//var recordHeader int
	//if header.Record() > 0 {
	//	if int(header.Record()) != len(data) {
	//		return ErrRecordNotFixed
	//	}
	//	recordHeader = FixedHeader
	//} else {
	//	recordHeader = BlockRecordHeaderSize
	//}

	recordSize := len(data) + BlockRecordHeaderSize
	// Overflow?
	if recordSize > len(block) {
		// Is the record able to fit inside 1 block?
		if recordSize > maxSize {
			return ErrRecordTooBig
		}
		return ErrOverflow
	}

	// Is it the first record?
	if header.Count() == 0 {
		header.SetMin(id).
			SetStart(start)
	}
	header.SetMax(id).
		SetEnd(end).
		SetCompleted(timestamp).
		SetSize(header.Size() + uint16(BlockRecordHeaderSize) + uint16(len(data))).
		SetCount(header.Count() + 1)

	// Write Record ID
	binary.LittleEndian.PutUint64(block, uint64(id))
	// Write timestamp
	binary.LittleEndian.PutUint64(block[8:], uint64(timestamp))
	// Write start
	binary.LittleEndian.PutUint64(block[16:], uint64(start))
	// Write start
	binary.LittleEndian.PutUint64(block[24:], uint64(end))
	// Write Record data size
	binary.LittleEndian.PutUint16(block[32:], uint16(len(data)))
	copy(block[34:], data)
	// Write Record data size to support reverse iteration
	binary.LittleEndian.PutUint16(block[34+len(data):], uint16(len(data)))

	return nil
}

func (b *Block1Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block1MaxDataSize,
		data,
	)
}

func (b *Block2Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block2MaxDataSize,
		data,
	)
}

func (b *Block4Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block4MaxDataSize,
		data,
	)
}

func (b *Block8Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block8MaxDataSize,
		data,
	)
}

func (b *Block16Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block16MaxDataSize,
		data,
	)
}

func (b *Block32Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block32MaxDataSize,
		data,
	)
}

func (b *Block64Mut) AppendRecord(id, created, time, end int64, data []byte) error {
	return appendRecord(
		b.HeaderMut(),
		b.Body()[b.Header().Size():],
		id,
		created,
		time,
		end,
		Block64MaxDataSize,
		data,
	)
}
