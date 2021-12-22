package model

import "time"

// BlockMessageBuilder forges new blocks for a stream
type BlockMessageBuilder struct {
	BlockHeaderMut
	sizeOffset int
	maxSize    int
	W          Writer
}

func (bw *BlockMessageBuilder) SetMaxSize(maxSize int) {
	if maxSize < Block1MaxDataSize {
		maxSize = Block1MaxDataSize
	} else if maxSize > Block64MaxDataSize {
		maxSize = Block64MaxDataSize
	}
	bw.maxSize = maxSize
}

func (bw *BlockMessageBuilder) IsEmpty() bool {
	return bw.Count() == 0
}

func (bw *BlockMessageBuilder) Clone() BlockMessageBuilder {
	return BlockMessageBuilder{
		BlockHeaderMut: bw.BlockHeaderMut,
		sizeOffset:     bw.sizeOffset,
		maxSize:        bw.maxSize,
		W:              bw.W.Clone(),
	}
}

func (bw *BlockMessageBuilder) Header() *BlockHeaderMut {
	return &bw.BlockHeaderMut
}

func (bw *BlockMessageBuilder) Data() []byte {
	return bw.W.b[0:bw.W.i]
}

func (bw *BlockMessageBuilder) Alloc(maxSize int) error {
	bw.SetMaxSize(maxSize)
	return bw.W.Ensure(bw.maxSize - bw.W.i)
}

func (bw *BlockMessageBuilder) Reset() {
	bw.BlockHeaderMut = BlockHeaderMut{}
	bw.W.Reset()
}

func (bw *BlockMessageBuilder) SizeofRecord(record *RecordMessage) int {
	if record == nil {
		return 0
	}
	return BlockRecordHeaderSize + len(record.Data)
}

func (bw *BlockMessageBuilder) HasCapacity(size int) bool {
	if bw.maxSize <= 0 {
		bw.maxSize = Block64MaxDataSize
	}
	return bw.sizeOffset+bw.W.Size()+size <= bw.maxSize
}

func (bw *BlockMessageBuilder) Append(record *RecordMessage, completed []*BlockMessage) ([]*BlockMessage, int, error) {
	if record == nil {
		return completed, 0, ErrNil
	}

	var err error

	err = record.DecompressInline()
	if err != nil {
		return completed, 0, err
	}

	h := &bw.BlockHeaderMut
	count := 0
	if h.Completed() > 0 {
		completed = append(completed, bw.flush())
		count++
	} else if !bw.HasCapacity(len(record.Data) + BlockRecordHeaderSize) {
		completed = append(completed, bw.Flush(record.Timestamp()))
		count++
	}

	if h.Count() == 0 {
		bw.W.Alloc()
		h.
			SetStreamID(record.StreamID()).
			SetId(record.BlockID()).
			SetCreated(record.Timestamp()).
			SetSeq(record.Seq()).
			SetMin(record.MessageID()).
			SetMax(record.MessageID()).
			SetStart(record.Start()).
			SetEnd(record.End()).
			SetCount(1)
	} else {
		if h.StreamID() != record.StreamID() {
			return completed, count, ErrWrongStream
		}
		if h.Id() != record.BlockID() {
			return completed, count, ErrWrongBlock
		}
		if record.Seq() != h.Seq()+h.Count() {
			return completed, count, ErrGap
		}
		// Update header
		h.SetCount(h.Count() + 1).SetMax(record.MessageID()).
			SetEnd(record.End()).SetSeq(record.Seq())
	}

	if record.Eob() {
		h.SetCompleted(record.Timestamp())
	}

	if err = bw.W.WriteInt64(record.MessageID()); err != nil {
		return completed, count, err
	}
	if err = bw.W.WriteInt64(record.Timestamp()); err != nil {
		return completed, count, err
	}
	if err = bw.W.WriteInt64(record.Start()); err != nil {
		return completed, count, err
	}
	if err = bw.W.WriteInt64(record.End()); err != nil {
		return completed, count, err
	}
	size := uint16(len(record.Data))
	if err = bw.W.WriteUint16(size); err != nil {
		return completed, count, err
	}
	if size > 0 {
		if err = bw.W.Write(record.Data); err != nil {
			return completed, count, err
		}
	}
	if err = bw.W.WriteUint16(size); err != nil {
		return completed, count, err
	}

	h.SetSize(uint16(bw.W.Size()))
	h.SetSizeU(h.Size())

	if h.Completed() > 0 {
		completed = append(completed, bw.Flush(h.Completed()))
		count++
	}

	return completed, count, nil
}

func (bw *BlockMessageBuilder) AppendBlock(block *BlockMessage, completed []*BlockMessage) ([]*BlockMessage, int, error) {
	if block == nil || block.Count() == 0 || len(block.Body) == 0 {
		return completed, 0, nil
	}

	var err error
	count := 0
	// Decompress block
	err = block.DecompressInline()
	if err != nil {
		return completed, 0, err
	}

	h := &bw.BlockHeaderMut
	if h.Completed() > 0 {
		completed = append(completed, bw.flush())
		count++
	} else if !bw.HasCapacity(len(block.Data())) {
		completed = append(completed, bw.Flush(time.Now().UnixNano()))
		count++
	}

	if h.Count() == 0 {
		bw.W.Alloc()
		bw.BlockHeaderMut = block.BlockHeaderMut
		h.SetSizeX(0)
		err = bw.W.Write(block.Body)
		if err != nil {
			return completed, count, err
		}
	} else {
		// In sequence?
		if block.Seq() != h.Seq()+h.Count() {
			return completed, count, ErrGap
		}

		// Update header
		h.SetCompleted(block.Completed()).
			SetMax(block.Max()).
			SetEnd(block.End()).
			SetCount(h.Count() + block.Count()).
			SetSize(h.Size() + uint16(len(block.Body))).
			SetSizeU(h.SizeU() + block.SizeU())

		err = bw.W.Write(block.Body)
		if err != nil {
			return completed, count, err
		}
	}

	if h.Completed() > 0 {
		completed = append(completed, bw.flush())
		count++
	}

	return completed, count, nil
}

func (bw *BlockMessageBuilder) flush() *BlockMessage {
	if bw.Size() == 0 {
		return nil
	}
	bw.sizeOffset = 0
	data := bw.W.Take()
	block := GetBlockMessage(0)
	block.Body = data
	bw.BlockHeaderMut = BlockHeaderMut{}
	bw.BlockHeaderMut.
		SetStreamID(block.StreamID()).
		SetId(block.MessageID() + 1)
	return block
}

func (bw *BlockMessageBuilder) Flush(completed int64) *BlockMessage {
	if bw.Size() == 0 {
		return nil
	}
	bw.SetCompleted(completed)
	return bw.flush()
}
