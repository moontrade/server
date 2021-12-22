package model

import (
	"fmt"
	"strconv"
	"testing"
)

func TestStreamID_Shard(t *testing.T) {
	//id, err := strconv.ParseInt("561358757089511", 10, 64)
	//if err != nil {
	//	t.Fatal(err)
	//}
	id := GenerateStreamID(15)
	shard := StreamID(id).DB()

	id2 := StreamID(id).ID()
	fmt.Println(id, shard, id2)

	fmt.Println(len(strconv.Itoa(int(id))))
}
