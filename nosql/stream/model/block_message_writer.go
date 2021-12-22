package model

// BlockMessageWriter reassembles blocks forged elsewhere. It adheres to the
// rules and limitations of Blocks.
type BlockMessageWriter struct {
	BlockHeaderMut
	W Writer
}

func (bw *BlockMessageWriter) IsEmpty() bool {
	return bw.Count() == 0
}

func (bw *BlockMessageWriter) Clone() BlockMessageWriter {
	return BlockMessageWriter{
		BlockHeaderMut: bw.BlockHeaderMut,
		W:              bw.W.Clone(),
	}
}

func (bw *BlockMessageWriter) Header() *BlockHeaderMut {
	return &bw.BlockHeaderMut
}

func (bw *BlockMessageWriter) Data() []byte {
	return bw.W.b[0:bw.W.i]
}

func (bw *BlockMessageWriter) Reset() {
	bw.BlockHeaderMut = BlockHeaderMut{}
	bw.W.Reset()
}

func (bw *BlockMessageWriter) SizeofRecord(record *RecordMessage) int {
	if record == nil {
		return 0
	}
	return BlockRecordHeaderSize + len(record.Data)
}

func (bw *BlockMessageWriter) HasCapacity(size int) bool {
	// Blocks can never be larger than ~64kb
	return bw.W.Size()+size <= Block64MaxDataSize
}

func (bw *BlockMessageWriter) Append(record *RecordMessage) (*BlockMessage, error) {
	if record == nil {
		return nil, ErrNil
	}

	var err error
	err = record.DecompressInline()
	if err != nil {
		return nil, err
	}

	h := &bw.BlockHeaderMut
	if h.Completed() > 0 {
		return nil, ErrBlockCompleted
	}

	if h.Count() == 0 {
		if !bw.HasCapacity(bw.SizeofRecord(record)) {
			return nil, ErrRecordTooBig
		}
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
			return nil, ErrWrongStream
		}
		if h.Id() != record.BlockID() {
			return nil, ErrWrongBlock
		}
		if record.Seq() != h.Seq()+h.Count() {
			return nil, ErrGap
		}
		if !bw.HasCapacity(bw.SizeofRecord(record) + bw.W.Size()) {
			return nil, ErrRecordTooBig
		}
		// Update header
		h.SetCount(h.Count() + 1).SetMax(record.MessageID()).
			SetEnd(record.End()).SetSeq(record.Seq())
	}

	if err = bw.W.WriteInt64(record.MessageID()); err != nil {
		return nil, err
	}
	if err = bw.W.WriteInt64(record.Timestamp()); err != nil {
		return nil, err
	}
	size := uint16(len(record.Data))
	if err = bw.W.WriteUint16(size); err != nil {
		return nil, err
	}
	if size > 0 {
		if err = bw.W.Write(record.Data); err != nil {
			return nil, err
		}
	}
	if err = bw.W.WriteUint16(size); err != nil {
		return nil, err
	}

	h.SetSize(uint16(bw.W.Size()))
	h.SetSizeU(h.Size())

	if record.Eob() {
		h.SetCompleted(record.Timestamp())
		return bw.flush(), nil
	}

	return nil, nil
}

// AppendBlock assumes the block was created elsewhere and must build "as is".
func (bw *BlockMessageWriter) AppendBlock(block *BlockMessage) (*BlockMessage, error) {
	if block == nil || block.Count() == 0 || len(block.Body) == 0 {
		return nil, nil
	}

	h := &bw.BlockHeaderMut
	// If block is already complete then return as is
	if block.IsComplete() {
		if h.Count() > 0 {
			return nil, ErrBlockCompleted
		}
		return block, nil
	}

	var err error
	// Decompress block
	err = block.DecompressInline()
	if err != nil {
		return nil, err
	}

	if h.Completed() > 0 {
		return nil, ErrBlockCompleted
	}

	if h.Count() == 0 {
		bw.W.Alloc()
		bw.BlockHeaderMut = block.BlockHeaderMut
		h.SetSizeX(0)
		err = bw.W.Write(block.Body)
		if err != nil {
			return nil, err
		}
	} else {
		// In sequence?
		if block.Seq() != h.Seq()+h.Count() {
			return nil, ErrGap
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
			return nil, err
		}
	}

	if h.Completed() > 0 {
		return bw.flush(), nil
	}

	return nil, nil
}

func (bw *BlockMessageWriter) flush() *BlockMessage {
	data := bw.W.Take()
	block := GetBlockMessage(0)
	block.Body = data
	bw.BlockHeaderMut = BlockHeaderMut{}
	return block
}

func (bw *BlockMessageWriter) Flush(completed int64) *BlockMessage {
	bw.SetCompleted(completed)
	return bw.flush()
}
