package index

import (
	"fmt"
	"strconv"
	"testing"
	"unsafe"
)

func TestNewBTreeI64(t *testing.T) {
	m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 256, BTreeInt64Compare(), 0)
	defer m.Free()

	m.Set(100, 101)
	result := m.Get(100)

	fmt.Println(result)
}

func BenchmarkBTreeI64_Set(b *testing.B) {
	runInt64 := func(entries int) {
		b.Run("int64 "+strconv.Itoa(entries)+" entries", func(b *testing.B) {
			m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 64, BTreeInt64Compare(), 0)
			defer m.Free()

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.Set(int64(i), uintptr(i))
			}
			b.StopTimer()
		})
		b.Run("int64 HINT "+strconv.Itoa(entries)+" entries", func(b *testing.B) {
			m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 64, BTreeInt64Compare(), 0)
			defer m.Free()

			b.ResetTimer()
			b.ReportAllocs()
			hint := uint64(0)
			for i := 0; i < b.N; i++ {
				m.SetHint(int64(i), uintptr(i), &hint)
			}
			b.StopTimer()
		})
	}
	runInt64(10)
	runInt64(100)
	runInt64(1000)
	runInt64(10000)
	runInt64(100000)
	runInt64(1000000)
	//runInt64(10000000)
}

func BenchmarkBTreeI64_Get(b *testing.B) {
	runInt64 := func(entries int) {
		b.Run("int64 "+strconv.Itoa(entries)+" entries", func(b *testing.B) {
			m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 64, BTreeInt64Compare(), 0)
			defer m.Free()

			for i := 0; i < entries; i++ {
				m.Set(int64(i), uintptr(i))
			}
			//key.SetInt64BE(0, int64(entries/2))
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.Get(int64(i))
			}
			b.StopTimer()
		})
		b.Run("int64 HINT "+strconv.Itoa(entries)+" entries", func(b *testing.B) {
			m := NewBTree(unsafe.Sizeof(BTreeRecordInt64{}), 64, BTreeInt64Compare(), 0)
			defer m.Free()

			hint := uint64(0)
			for i := 0; i < entries; i++ {
				m.SetHint(int64(i), uintptr(i), &hint)
			}
			//key.SetInt64BE(0, int64(entries/2))
			hint = 0
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.GetHint(int64(i), &hint)
			}
			b.StopTimer()
		})
	}
	runInt64(10)
	runInt64(100)
	runInt64(1000)
	runInt64(10000)
	runInt64(100000)
	runInt64(1000000)
	//runInt64(10000000)
}
