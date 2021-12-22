package model

func (s *Savepoint) Type() MessageType {
	return MessageType_Savepoint
}

func (s *Savepoint) MessageID() int64 {
	return s.RecordID().Id()
}

func (s *Savepoint) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m   = s.Mut()
		v   = int64(0)
		err error
	)
	if err = s.RecordID().UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetTimestamp(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetWriterID(v)

	return nil
}

func (s *Savepoint) MarshalCompactTo(w *Writer) error {
	var err error
	if err = s.RecordID().MarshalCompactTo(w); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Timestamp()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.WriterID()); err != nil {
		return err
	}
	return nil
}
