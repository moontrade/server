package model

func (r *RecordID) UnmarshalCompactFrom(rd *Reader) error {
	var (
		m   = r.Mut()
		v   = int64(0)
		err error
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

	return nil
}

func (r *RecordID) MarshalCompactTo(w *Writer) error {
	var err error
	if _, err = w.WriteVarint64(r.StreamID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(r.BlockID()); err != nil {
		return err
	}
	if _, err = w.WriteVarint64(r.Id()); err != nil {
		return err
	}
	return nil
}
