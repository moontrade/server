package model

func (e *EOS) Type() MessageType {
	return MessageType_EOS
}

func (s *EOS) MessageID() int64 {
	return s.RecordID().Id()
}

func (r *EOS) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m   = r.Mut()
		v   = int64(0)
		err error
	)
	if err = r.RecordID().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetTimestamp(v)

	return nil
}

func (r *EOS) MarshalCompactTo(w *Writer) error {
	var err error
	if err = r.RecordID().MarshalCompactTo(w); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(r.Timestamp()); err != nil {
		return err
	}
	return nil
}
