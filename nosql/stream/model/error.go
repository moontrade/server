package model

import (
	"errors"
)

var (
	ErrUnknownCompression = errors.New("unknown compression algorithm")
	ErrOverflow           = errors.New("overflow")
	ErrVarintOverflow     = errors.New("varint overflow")
	ErrFieldTagTooBig     = errors.New("field tag too big")
	ErrIDTooSmall         = errors.New("id is equal or smaller than top item")
	ErrTimeIsPast         = errors.New("time is less than top item")
	ErrRecordTooBig       = errors.New("record too big")
	ErrRecordNotFixed     = errors.New("expected a fixed record size")
	ErrCorrupted          = errors.New("corrupted")
	ErrNil                = errors.New("nil")
	ErrNotExist           = errors.New("corrupted")
	ErrSizeDataMismatch   = errors.New("size != len(data)")
	ErrBlockAllocatorOOM  = errors.New("blockAllocator.Alloc() returned nil")
	ErrOutOfMemory        = errors.New("out of memory")
	ErrOutOfSync          = errors.New("out of sync")
	ErrOutOfOrder         = errors.New("out of order")
	ErrRecordOrBlocks     = errors.New("oneof record or blocks")
	ErrGap                = errors.New("gap")
	ErrWrongStream        = errors.New("wrong stream")
	ErrWrongBlock         = errors.New("wrong block")
	ErrBlockCompleted     = errors.New("block completed")
	ErrConcurrentWriters  = errors.New("concurrent writers")
	ErrUnknownMessage     = errors.New("unknown message")
	ErrNotEnd             = errors.New("not end")
)
