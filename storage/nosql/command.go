package nosql

import (
	"reflect"
	"unsafe"
)

type String32 struct {
	Data [31]byte
	Size byte
}

func (s *String32) Set(value string) {
	l := len(value)
	if l > 31 {
		l = 31
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *String32) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 31 {
		s.Size = 31
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *String32) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type InsertCommand struct {
	Collection String32
	Data       []byte
}
