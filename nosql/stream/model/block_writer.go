package model

// BlockWriter simplifies building Blocks by appending 1 record at a time
type BlockWriter struct {
	header    BlockHeaderMut
	current   BlockMut
	allocator BlockAllocator
}

func (bw *BlockWriter) Current() BlockMut {
	return bw.current
}

func NewBlockWriter(max BlockHeaderMut, allocator BlockAllocator) *BlockWriter {
	if allocator == nil {
		return nil
	}
	return &BlockWriter{
		header:    max,
		allocator: allocator,
	}
}

func (w *BlockWriter) NextBlock() BlockMut {
	w.header.SetId(w.header.Id() + 1)
	b := w.allocator.Alloc()
	if b == nil {
		return nil
	}
	copy(b.Header().Bytes(), w.header.Bytes())
	return b
}

// Record tries to append the next record returning a full Block once filled.
func (w *BlockWriter) AppendRecord(record *RecordMessage) (BlockMut, int64, uint16, error) {
	if record.RecordHeader.Compression() != Compression_None {
		var err error
		record, err = record.Decompress()
		if err != nil {
			return nil, 0, 0, err
		}
	}
	return w.Append(
		record.RecordHeader.MessageID(),
		record.RecordHeader.Timestamp(),
		record.RecordHeader.Start(),
		record.RecordHeader.End(),
		record.Data,
	)
}

// Append tries to append the next record returning a full Block once filled.
func (w *BlockWriter) Append(id, created, time, end int64, data []byte) (BlockMut, int64, uint16, error) {
	if w.current == nil {
		w.current = w.NextBlock()
		if w.current == nil {
			return nil, 0, 0, ErrBlockAllocatorOOM
		}
	}

	err := w.current.AppendRecord(id, created, time, end, data)
	if err != nil {
		if err != ErrOverflow {
			return nil, 0, 0, err
		}

		flush := w.current
		w.current = w.NextBlock()
		if w.current == nil {
			return flush, 0, 0, ErrBlockAllocatorOOM
		}
		err = w.current.AppendRecord(id, created, time, end, data)
		return flush, w.current.Header().Id(), w.current.Header().Count(), err
	}

	return nil, w.current.Header().Id(), w.current.Header().Count(), nil
}
