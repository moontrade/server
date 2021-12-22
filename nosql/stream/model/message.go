package model

type Message interface {
	Type() MessageType

	// Sequential ID of message
	MessageID() int64

	// Timestamp of the message
	Timestamp() int64

	MarshalBinary() ([]byte, error)

	UnmarshalBinary(b []byte) error

	MarshalCompactTo(w *Writer) error

	UnmarshalCompactFrom(rd *Reader) error
}

func MarshalMessageCompactTo(message Message, w *Writer) error {
	if message == nil {
		return nil
	}
	_ = w.WriteByte(byte(message.Type()))
	return message.MarshalCompactTo(w)
}

func UnmarshalMessageCompact(rd *Reader) (Message, error) {
	b, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}

	var message Message
	switch MessageType(b) {
	case MessageType_Record:
		message = GetRecord(0)
	case MessageType_Block:
		message = GetBlockMessage(0)
	case MessageType_EOS:
		message = GetEOS()
	case MessageType_EOB:
		message = GetEOB()
	case MessageType_Savepoint:
		message = GetSavepoint()
	case MessageType_Starting:
		message = GetStarting()
	case MessageType_Progress:
		message = GetProgress()
	case MessageType_Started:
		message = GetStarted()
	case MessageType_Stopped:
		message = GetStopped()
	}

	if message == nil {
		return nil, ErrUnknownMessage
	}

	if err = message.UnmarshalCompactFrom(rd); err != nil {
		return nil, err
	}
	return message, nil
}
