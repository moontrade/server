package model

func (s *Starting) Type() MessageType {
	return MessageType_Starting
}

func (s *Starting) MessageID() int64 {
	return s.RecordID().Id()
}

func (s *Progress) Type() MessageType {
	return MessageType_Progress
}

func (s *Progress) MessageID() int64 {
	return s.RecordID().Id()
}

func (s *Started) Type() MessageType {
	return MessageType_Started
}

func (s *Started) MessageID() int64 {
	return s.RecordID().Id()
}

func (s *Stopped) Type() MessageType {
	return MessageType_Stopped
}

func (s *Stopped) MessageID() int64 {
	return s.RecordID().Id()
}

func (s *Starting) UnmarshalCompactFrom(rd *Reader) error {
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

func (s *Starting) MarshalCompactTo(w *Writer) error {
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

func (s *Progress) UnmarshalCompactFrom(rd *Reader) error {
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
	m.SetStarted(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetCount(v)

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetRemaining(v)

	return nil
}

func (s *Progress) MarshalCompactTo(w *Writer) error {
	var err error
	if err = s.RecordID().MarshalCompactTo(w); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Timestamp()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Started()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Count()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Remaining()); err != nil {
		return err
	}
	return nil
}

func (s *Started) UnmarshalCompactFrom(rd *Reader) error {
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

	if v, _, err = rd.ReadVarint64(); err != nil {
		return err
	}
	m.SetStops(v)

	return nil
}

func (s *Started) MarshalCompactTo(w *Writer) error {
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
	if _, err = w.WriteVarint64(s.Stops()); err != nil {
		return err
	}
	return nil
}

func (s *Stopped) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m   = s.Mut()
		v   = int64(0)
		b   = byte(0)
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
	m.SetStarts(v)

	if b, err = rd.ReadByte(); err != nil {
		return err
	}
	m.SetReason(StopReason(b))

	return nil
}

func (s *Stopped) MarshalCompactTo(w *Writer) error {
	var err error
	if err = s.RecordID().MarshalCompactTo(w); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Timestamp()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(s.Starts()); err != nil {
		return err
	}
	if err = w.WriteByte(byte(s.Reason())); err != nil {
		return err
	}
	return nil
}
