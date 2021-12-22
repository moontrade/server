package model

import (
	"sync"
)

var (
	recordPool = &sync.Pool{New: func() interface{} {
		return &RecordMessage{}
	}}
)

func GetRecord(size int) *RecordMessage {
	record := recordPool.Get().(*RecordMessage)
	if size > 0 {
		record.Data = make([]byte, size)
		//record.Data = pbytes.GetLen(size)
	}
	return record
}

func PutRecord(record *RecordMessage) {
	if record.Data != nil {
		//pbytes.Put(record.Data)
		record.Data = nil
	}
	record.RecordHeader = RecordHeader{}
	recordPool.Put(record)
}
