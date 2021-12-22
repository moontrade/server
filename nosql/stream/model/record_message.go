package model

// RecordMessage represents a single record in a block. Appenders append as few as a single
// record at a time.
type RecordMessage struct {
	RecordHeader
	Data []byte // Record data
}

func (r *RecordMessage) Type() MessageType {
	return MessageType_Record
}

func (r *RecordMessage) UnmarshalCompactFrom(rd *Reader) error {
	if err := r.RecordHeader.UnmarshalCompactFrom(rd); err != nil {
		return err
	}
	if r.RecordHeader.Size() == 0 {
		if r.Data != nil {
			//pbytes.Put(r.Data)
			r.Data = nil
		}
		return nil
	}
	if r.Data == nil {
		r.Data = make([]byte, int(r.RecordHeader.Size()))
		//r.Data = pbytes.GetLen(int(r.RecordHeader.Size()))
	} else if cap(r.Data) < int(r.RecordHeader.Size()) {
		//pbytes.Put(r.Data)
		r.Data = make([]byte, int(r.RecordHeader.Size()))
		//r.Data = pbytes.GetLen(int(r.RecordHeader.Size()))
	} else {
		r.Data = r.Data[0:r.RecordHeader.Size()]
	}
	return rd.ReadBytes(r.Data)
}

func (r *RecordMessage) MarshalCompactTo(w *Writer) error {
	if int(r.RecordHeader.Size()) != len(r.Data) {
		return ErrSizeDataMismatch
	}
	if err := r.RecordHeader.MarshalCompactTo(w); err != nil {
		return err
	}
	return w.Write(r.Data)
}
