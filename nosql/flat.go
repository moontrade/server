package nosql

import (
	"reflect"
	"unsafe"
)

type FlatString16 struct {
	Data [15]byte
	Size byte
}

func (s *FlatString16) Set(value string) {
	l := len(value)
	if l > 15 {
		l = 15
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString16) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 15 {
		s.Size = 15
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *FlatString16) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type FlatString32 struct {
	Data [31]byte
	Size byte
}

func (s *FlatString32) Set(value string) {
	l := len(value)
	if l > 31 {
		l = 31
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString32) String() string {
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

func (s *FlatString32) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type FlatString64 struct {
	Data [63]byte
	Size byte
}

func (s *FlatString64) Set(value string) {
	l := len(value)
	if l > 63 {
		l = 63
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString64) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 63 {
		s.Size = 63
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *FlatString64) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type FlatString96 struct {
	Data [95]byte
	Size byte
}

func (s *FlatString96) Set(value string) {
	l := len(value)
	if l > 95 {
		l = 95
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString96) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 95 {
		s.Size = 95
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *FlatString96) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type FlatString128 struct {
	Data [127]byte
	Size byte
}

func (s *FlatString128) Set(value string) {
	l := len(value)
	if l > 127 {
		l = 127
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString128) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 127 {
		s.Size = 127
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *FlatString128) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}

type FlatString255 struct {
	Data [254]byte
	Size byte
}

func (s *FlatString255) Set(value string) {
	l := len(value)
	if l > 254 {
		l = 254
	}
	copy(s.Data[0:], value[0:l])
	s.Size = byte(l)
}

func (s *FlatString255) String() string {
	if s.Size == 0 {
		return ""
	}
	if s.Size > 254 {
		s.Size = 254
	}
	b := make([]byte, s.Size)
	copy(b, *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	})))
	return string(b)
}

func (s *FlatString255) Unsafe() string {
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&s.Data[0])),
		Len:  int(s.Size),
	}))
}
