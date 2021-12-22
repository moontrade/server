package model

import (
	"fmt"
	"testing"
	"time"
)

func TestBlockMut_Compress(t *testing.T) {
	b := &Block64Mut{}
	b.HeaderMut().SetStreamID(1).SetId(1).SetStart(time.Now().UnixNano())
	x, err := Compress(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(x))

	bb := &Block64Mut{}
	err = bb.Uncompress(x)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(b.Header().Id(), bb.Header().Id(), b.Header(),
		bb.Header(), b.Header().Start(), bb.Header().Start())
	fmt.Println(*b == *bb)
}
