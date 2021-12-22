package model

func (h *RecordHeader) MarshalCompactTo(w *Writer) error {
	var err error
	if _, err = w.WriteVarint64(h.StreamID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.BlockID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.MessageID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(h.Timestamp()); err != nil {
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

func (r *RecordHeader) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m    = r.Mut()
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
	m.SetBlockID(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetId(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetTimestamp(v)

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
