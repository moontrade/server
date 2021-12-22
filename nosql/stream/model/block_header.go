package model

import "unsafe"

var (
	SizeofBlockHeader  = int(unsafe.Sizeof(BlockHeader{}))
	SizeofRecordHeader = int(unsafe.Sizeof(RecordHeader{}))
)

func (b *BlockHeader) IsComplete() bool {
	return b.Completed() > 0 && b.Seq() == 0
}

func (b *BlockHeader) IsPartial() bool {
	return b.Completed() == 0 || b.Seq() > 0
}

func (h *BlockHeader) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m    = h.Mut()
		v    = int64(0)
		vu16 = uint16(0)
		b    = byte(0)
		err  error
	)
	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetStreamID(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetId(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetCreated(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetCompleted(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetCreated(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetMin(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetMax(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetStart(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetEnd(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetSavepoint(v)

	if vu16, _, err = rd.ReadUVarint16(); err != nil {
		return err
	}
	m.SetCount(vu16)

	if vu16, _, err = rd.ReadUVarint16(); err != nil {
		return err
	}
	m.SetSeq(vu16)

	if vu16, _, err = rd.ReadUVarint16(); err != nil {
		return err
	}
	m.SetSize(vu16)

	if vu16, _, err = rd.ReadUVarint16(); err != nil {
		return err
	}
	m.SetSizeU(vu16)

	if vu16, _, err = rd.ReadUVarint16(); err != nil {
		return err
	}
	m.SetSizeX(vu16)

	if b, err = rd.ReadByte(); err != nil {
		return err
	}
	m.SetCompression(Compression(b))

	return nil
}

func (h *BlockHeader) MarshalCompactTo(w *Writer) error {
	var err error
	if _, err = w.WriteVarint64(h.StreamID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Id()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Created()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Completed()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Min()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Max()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Start()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.End()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Savepoint()); err != nil {
		return err
	}
	if _, err = w.WriteUVarint16(h.Count()); err != nil {
		return err
	}
	if _, err = w.WriteUVarint16(h.Seq()); err != nil {
		return err
	}
	if _, err = w.WriteUVarint16(h.Size()); err != nil {
		return err
	}
	if _, err = w.WriteUVarint16(h.SizeU()); err != nil {
		return err
	}
	if _, err = w.WriteUVarint16(h.SizeX()); err != nil {
		return err
	}
	if err = w.WriteByte(byte(h.Compression())); err != nil {
		return err
	}
	return nil
}
