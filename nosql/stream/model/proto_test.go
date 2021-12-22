package model

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestPage_Bytes(t *testing.T) {
	printLayout(page{})
	fmt.Println()
	printLayout(Block64{})

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, 10)
	fmt.Println(binary.LittleEndian.Uint64(b))
}

type page struct {
	first    int64
	duration int64
	last     int64
	count    uint16
	record   uint16
	size     uint16
	xsize    uint16
}

func printLayout(t interface{}) {
	// First ask Go to give us some information about the MyData type
	typ := reflect.TypeOf(t)
	fmt.Printf("Struct is %d bytes long\n", typ.Size())
	// We can run through the fields in the structure in order
	n := typ.NumField()
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		fmt.Printf("%s at offset %v, size=%d, align=%d\n",
			field.Name, field.Offset, field.Type.Size(),
			field.Type.Align())
	}
}
