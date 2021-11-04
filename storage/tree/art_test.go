package tree

import (
	"fmt"
	"github.com/moontrade/nogc"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestLock(t *testing.T) {
	l := NewLock()
	l.LockUnsafe()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		println("go 1 before")
		l.LockUnsafe()
		println("go 1 after. waiting 2 seconds...")
		SleepUnsafe(time.Second * 2)
		l.UnlockUnsafe()
	}()

	go func() {
		defer wg.Done()
		println("go 2 before. waiting 2 seconds...")
		SleepUnsafe(time.Second * 2)
		//time.Sleep(time.Second*2)
		l.Unlock()
		SleepUnsafe(time.Second)
		l.LockUnsafe()
		println("go 2 after")
		l.UnlockUnsafe()
	}()

	wg.Wait()
}

func TestNew(t *testing.T) {
	println("sizeof Leaf", unsafe.Sizeof(Leaf{}))
	art, _ := NewART()
	art.InsertBytes(memory.WrapString("hello"), Value{
		File:   1,
		Offset: 2,
		Data:   3,
	})
	art.InsertBytes(memory.WrapString("below"), Value{
		File:   1,
		Offset: 2,
		Data:   3,
	})
	art.InsertString("hello01", Value{
		File:   1,
		Offset: 2,
		Data:   3,
	})
	found := art.FindBytes(memory.WrapString("hello"))
	fmt.Println("found", found)
	println("size", art.Size())
	min := art.Minimum()
	println("min", uint(min.Data()), "key", min.Key().String())
	max := art.Maximum()
	println("max", uint(max.Data()), "key", max.Key().String())
	art.Free()
}

func TestBytes(t *testing.T) {
	tree, _ := NewART()
	key := memory.AllocBytes(8)

	//println(tree.String())
	for i := 0; i < 1000; i++ {
		key.SetInt64BE(0, int64(i))
		tree.InsertBytes(key, Value{
			File:   1,
			Offset: 2,
			Data:   3,
		})
	}
	fmt.Println("1000:			", tree.Bytes())
	for i := 1000; i < 10000; i++ {
		key.SetInt64BE(0, int64(i))
		tree.InsertBytes(key, Value{
			File:   1,
			Offset: 2,
			Data:   3,
		})
	}
	fmt.Println("10000:			", tree.Bytes())
	for i := 10000; i < 100000; i++ {
		key.SetInt64BE(0, int64(i))
		tree.InsertBytes(key, Value{
			File:   1,
			Offset: 2,
			Data:   3,
		})
	}
	fmt.Println("100000:			", tree.Bytes())
	for i := 100000; i < 1000000; i++ {
		key.SetInt64BE(0, int64(i))
		tree.InsertBytes(key, Value{
			File:   1,
			Offset: 2,
			Data:   3,
		})
	}
	fmt.Println("10000:			", tree.Bytes())
}

func BenchmarkTree_Insert(b *testing.B) {
	b.Run("insert int32BE", func(b *testing.B) {
		tree, _ := NewART()
		key := memory.AllocBytes(4)

		//println(tree.String())
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			key.SetInt32BE(0, int32(i))
			tree.InsertBytes(key, Value{
				File:   1,
				Offset: 2,
				Data:   3,
			})
		}
		b.StopTimer()
		tree.Free()
		key.Free()
	})
	b.Run("insert int32LE", func(b *testing.B) {
		tree, _ := NewART()
		key := memory.AllocBytes(4)

		//println(tree.String())
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			key.SetInt32LE(0, int32(i))
			tree.InsertBytes(key, Value{
				File:   1,
				Offset: 2,
				Data:   3,
			})
		}
		b.StopTimer()
		fmt.Println("tree bytes", tree.Bytes())
		fmt.Println("tree size", tree.Size())
		tree.Free()
		key.Free()
	})
	b.Run("insert int64BE", func(b *testing.B) {
		tree, _ := NewART()
		key := memory.AllocBytes(8)

		//println(tree.String())
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			key.SetInt64BE(0, int64(i))
			tree.InsertBytes(key, Value{
				File:   1,
				Offset: 2,
				Data:   3,
			})
		}
		b.StopTimer()
		fmt.Println("tree bytes", tree.Bytes())
		fmt.Println("tree size", tree.Size())
		tree.Free()
		key.Free()
	})
	b.Run("insert int64LE", func(b *testing.B) {
		tree, _ := NewART()
		key := memory.AllocBytes(8)

		//println(tree.String())
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			key.SetInt64LE(0, int64(i))
			tree.InsertBytes(key, Value{
				File:   1,
				Offset: 2,
				Data:   3,
			})
		}
		b.StopTimer()
		tree.Free()
		key.Free()
	})
	b.Run("insert int64 PointerSet", func(b *testing.B) {
		m := memory.NewPointerSet(uintptr(b.N * 2))
		key := memory.AllocBytes(8)

		var mu sync.Mutex
		//println(tree.String())
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			key.SetInt64BE(0, int64(i))
			mu.Lock()
			m.Set(uintptr(i))
			mu.Unlock()
		}
		b.StopTimer()
		key.Free()
		_ = m.Close()
	})
	b.Run("insert int64 gomap", func(b *testing.B) {
		m := make(map[int64]struct{}, b.N*2)
		key := memory.AllocBytes(8)

		var mu sync.Mutex
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			mu.Lock()
			m[int64(i)] = struct{}{}
			mu.Unlock()
		}
		b.StopTimer()
		key.Free()
	})
	b.Run("insert int64 BTree HINT", func(b *testing.B) {
		m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 64, BTreeInt64Compare(), 0)
		key := memory.AllocBytes(8)

		var mu sync.Mutex
		hint := uint64(0)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			mu.Lock()
			m.SetHint(int64(i), 101, &hint)
			mu.Unlock()
		}
		b.StopTimer()
		key.Free()
	})
}

func BenchmarkTree_Find(b *testing.B) {
	runInt64BE := func(entries int) {
		b.Run("int64BE "+strconv.Itoa(entries)+" entries", func(b *testing.B) {
			tree, _ := NewART()
			defer tree.Free()
			key := memory.AllocBytes(8)
			defer key.Free()

			for i := 0; i < entries; i++ {
				key.SetInt64BE(0, int64(i))
				tree.InsertBytes(key, Value{
					File:   1,
					Offset: 2,
					Data:   3,
				})
			}
			//key.SetInt64BE(0, int64(entries/2))
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				key.SetInt64BE(0, int64(i))
				tree.FindBytes(key)
			}
			b.StopTimer()
		})
	}
	runInt64BE(10)
	runInt64BE(100)
	runInt64BE(1000)
	runInt64BE(10000)
	runInt64BE(100000)
	runInt64BE(1000000)
	//runInt64BE(10000000)
}
