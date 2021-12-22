package model

type Record struct {
	ID        int64
	Timestamp int64
	Start     int64
	End       int64
	Data      []byte
}
