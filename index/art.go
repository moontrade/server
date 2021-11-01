//go:build !tinygo && (amd64 || arm64)
// +build !tinygo
// +build amd64 arm64

package index

/*
#cgo LDFLAGS: -L. -lstdc++
#cgo CXXFLAGS: -std=c++20 -I.
#include <stdlib.h>
#include <pthread.h>
#include "art.h"

typedef struct {
	size_t thread_safe;
	size_t ptr;
	size_t code;
} art_new_t;

void do_art_new(size_t arg0, size_t arg1) {
	art_new_t* args = (art_new_t*)(void*)arg0;
	art_tree* tree = (art_tree*)malloc(sizeof(art_tree));
	args->code = (size_t)art_tree_init(tree);
	if (args->thread_safe > 0) {
		art_tree_init_lock(tree); // Enable RWSpinLock
	}
	args->ptr = (size_t)(void*)tree;
}

void do_art_destroy(size_t arg0, size_t arg1) {
	art_tree* tree = (art_tree*)(void*)arg0;
	art_tree_destroy(tree);
	free((void*)tree);
}

typedef struct {
	size_t tree;
	size_t size;
} art_size_t;

void do_art_size(size_t arg0, size_t arg1) {
	art_size_t* args = (art_size_t*)(void*)arg0;
	args->size = (size_t)art_size((art_tree*)(void*)args->tree);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t len;
	art_value data;
	art_value old;
} art_insert_t;

void do_art_insert(size_t arg0, size_t arg1) {
	art_insert_t* args = (art_insert_t*)(void*)arg0;
	args->old = art_insert((art_tree*)(void*)args->tree, (unsigned char*)args->key, (int)args->len, args->data);
}

void do_art_insert_no_replace(size_t arg0, size_t arg1) {
	art_insert_t* args = (art_insert_t*)(void*)arg0;
	args->old = art_insert_no_replace((art_tree*)(void*)args->tree, (unsigned char*)args->key, (int)args->len, args->data);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t len;
	art_value value;
} art_delete_t;

void do_art_delete(size_t arg0, size_t arg1) {
	art_delete_t* args = (art_delete_t*)(void*)arg0;
	args->value = art_delete((art_tree*)(void*)args->tree, (unsigned char*)args->key, (int)args->len);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t len;
	art_value result;
} art_search_t;

void do_art_search(size_t arg0, size_t arg1) {
	art_search_t* args = (art_search_t*)(void*)arg0;
	art_tree* tree = (art_tree*)(void*)args->tree;
	args->result = art_search((art_tree*)(void*)args->tree, (unsigned char*)args->key, (int)args->len);
}

typedef struct {
	size_t tree;
	size_t result;
	size_t result2;
} art_minmax_t;

void do_art_minimum(size_t arg0, size_t arg1) {
	art_minmax_t* args = (art_minmax_t*)(void*)arg0;
	args->result = (size_t)art_minimum((art_tree*)(void*)args->tree);
}

void do_art_maximum(size_t arg0, size_t arg1) {
	art_minmax_t* args = (art_minmax_t*)(void*)arg0;
	args->result = (size_t)art_maximum((art_tree*)(void*)args->tree);
}

void do_art_minmax(size_t arg0, size_t arg1) {
	art_minmax_t* args = (art_minmax_t*)(void*)arg0;
	args->result = (size_t)art_minimum((art_tree*)(void*)args->tree);
	args->result2 = (size_t)art_maximum((art_tree*)(void*)args->tree);
}

*/
import "C"
import (
	"github.com/moontrade/nogc"
	"github.com/moontrade/nogc/unsafecgo"
	"reflect"
	"time"
	"unsafe"
)

type RWLock C.void

func NewLock() *RWLock {
	return (*RWLock)(unsafe.Pointer(C.moontrade_rwlock_new()))
}

func Sleep(nanos time.Duration) {
	C.art_sleep(C.uint64_t(nanos))
}

func SleepUnsafe(nanos time.Duration) {
	unsafecgo.NonBlocking((*byte)(C.art_sleep), uintptr(nanos), 0)
}

func (l *RWLock) Free() {
	C.moontrade_rwlock_destroy(unsafe.Pointer(l))
}

func (l *RWLock) Lock() {
	C.moontrade_rwlock_lock(unsafe.Pointer(l))
}

func (l *RWLock) LockUnsafe() {
	unsafecgo.NonBlocking((*byte)(C.moontrade_rwlock_lock), uintptr(unsafe.Pointer(l)), 0)
}

func (l *RWLock) Unlock() {
	C.moontrade_rwlock_unlock(unsafe.Pointer(l))
}

func (l *RWLock) UnlockUnsafe() {
	unsafecgo.NonBlocking((*byte)(C.moontrade_rwlock_unlock), uintptr(unsafe.Pointer(l)), 0)
}

func (l *RWLock) LockShared() {
	C.moontrade_rwlock_lock_shared(unsafe.Pointer(l))
}

func (l *RWLock) LockSharedUnsafe() {
	unsafecgo.NonBlocking((*byte)(C.moontrade_rwlock_lock_shared), uintptr(unsafe.Pointer(l)), 0)
}

func (l *RWLock) UnlockShared() {
	C.moontrade_rwlock_unlock_shared(unsafe.Pointer(l))
}

func (l *RWLock) UnlockSharedUnsafe() {
	unsafecgo.NonBlocking((*byte)(C.moontrade_rwlock_unlock_shared), uintptr(unsafe.Pointer(l)), 0)
}

type ART C.art_tree

func (t *ART) Lock() *RWLock {
	tree := (*C.art_tree)(unsafe.Pointer(t))
	return (*RWLock)(tree.lock)
}

func (t *ART) Size() int64 {
	return int64((*C.art_tree)(unsafe.Pointer(t)).size)
}

func (t *ART) Bytes() int64 {
	return int64((*C.art_tree)(unsafe.Pointer(t)).bytes)
}

type Leaf C.art_leaf

func (l *Leaf) Data() memory.Pointer {
	return *(*memory.Pointer)(unsafe.Pointer(l))
}
func (l *Leaf) Key() memory.FatPointer {
	return memory.FatPointerOf(
		memory.Pointer(uintptr(unsafe.Pointer(l))+unsafe.Sizeof(uintptr(0))+4),
		uintptr(*(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(l)) + unsafe.Sizeof(uintptr(0))))))
}

type artNewT struct {
	threadSafe uintptr
	ptr        uintptr
	code       uintptr
}

func NewART() (*ART, int) {
	return NewARTThreadSafe(false)
}

func NewARTThreadSafe(threadSafe bool) (*ART, int) {
	args := artNewT{}
	if threadSafe {
		args.threadSafe = 1
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_new), ptr, 0)
	return (*ART)(unsafe.Pointer(args.ptr)), int(args.code)
}

func (r *ART) Free() {
	ptr := uintptr(unsafe.Pointer(r))
	unsafecgo.NonBlocking((*byte)(C.do_art_destroy), ptr, 0)
}

//type artSizeT struct {
//	ptr  uintptr
//	size uintptr
//}
//
//func (r *ART) Size() int {
//	args := artSizeT{ptr: uintptr(unsafe.Pointer(r))}
//	ptr := uintptr(unsafe.Pointer(&args))
//	unsafecgo.NonBlocking((*byte)(C.do_art_size), ptr, 0)
//	return int(args.size)
//}

//func (t *ART) Size() int {
//	return int(*(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(t)) + unsafe.Sizeof(uintptr(0)))))
//}

//func (r *Art) Print() {
//	unsafecgo.NonBlocking((*byte)(C.do_rax_show), uintptr(unsafe.Pointer(r)), 0)
//}

type Value struct {
	File   uint32
	Offset uint32
	Data   uintptr
}

type artInsertT struct {
	tree  uintptr
	key   uintptr
	len   uintptr
	value Value
	old   Value
}

func (r *ART) Insert(key memory.Pointer, size int, value Value) Value {
	args := artInsertT{
		tree:  uintptr(unsafe.Pointer(r)),
		key:   uintptr(key),
		len:   uintptr(size),
		value: value,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_insert), ptr, 0)
	return args.old
}

func (r *ART) InsertBytes(key memory.Bytes, value Value) Value {
	return r.Insert(key.Pointer, key.Len(), value)
}

func (r *ART) InsertString(key string, value Value) Value {
	k := (*reflect.StringHeader)(unsafe.Pointer(&key))
	return r.Insert(memory.Pointer(k.Data), int(k.Len), value)
}

func (r *ART) InsertSlice(key []byte, value Value) Value {
	k := (*reflect.SliceHeader)(unsafe.Pointer(&key))
	return r.Insert(memory.Pointer(k.Data), int(k.Len), value)
}

func (r *ART) InsertNoReplace(key memory.Pointer, size int, value Value) Value {
	args := artInsertT{
		tree:  uintptr(unsafe.Pointer(r)),
		key:   uintptr(key),
		len:   uintptr(size),
		value: value,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_insert_no_replace), ptr, 0)
	return args.old
}

func (r *ART) InsertNoReplaceBytes(key memory.Bytes, value Value) Value {
	return r.InsertNoReplace(key.Pointer, key.Len(), value)
}

func (r *ART) InsertNoReplaceString(key string, value Value) Value {
	k := (*reflect.StringHeader)(unsafe.Pointer(&key))
	return r.InsertNoReplace(memory.Pointer(k.Data), int(k.Len), value)
}

func (r *ART) InsertNoReplaceSlice(key []byte, value Value) Value {
	k := (*reflect.SliceHeader)(unsafe.Pointer(&key))
	return r.InsertNoReplace(memory.Pointer(k.Data), int(k.Len), value)
}

type artDeleteT struct {
	tree uintptr
	key  uintptr
	len  uintptr
	item uintptr
}

func (r *ART) Delete(key memory.Pointer, size int) memory.Pointer {
	args := artDeleteT{
		tree: uintptr(unsafe.Pointer(r)),
		key:  uintptr(key),
		len:  uintptr(size),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_delete), ptr, 0)
	return memory.Pointer(args.item)
}

func (r *ART) DeleteBytes(key memory.Bytes) memory.Pointer {
	return r.Delete(key.Pointer, key.Len())
}

type artSearchT struct {
	tree   uintptr
	s      uintptr
	len    uintptr
	result Value
}

func (r *ART) Find(key memory.Pointer, size int) Value {
	args := artSearchT{
		tree: uintptr(unsafe.Pointer(r)),
		s:    uintptr(key),
		len:  uintptr(size),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_search), ptr, 0)
	return args.result
}

func (r *ART) FindBytes(key memory.Bytes) Value {
	return r.Find(key.Pointer, key.Len())
}

type artTreeMinmaxT struct {
	tree    uintptr
	result  uintptr
	result2 uintptr
}

func (r *ART) Minimum() *Leaf {
	args := artTreeMinmaxT{
		tree: uintptr(unsafe.Pointer(r)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_minimum), ptr, 0)
	return (*Leaf)(unsafe.Pointer(args.result))
}

// Maximum Returns the maximum valued leaf
// @return The maximum leaf or NULL
func (r *ART) Maximum() *Leaf {
	args := artTreeMinmaxT{
		tree: uintptr(unsafe.Pointer(r)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_maximum), ptr, 0)
	return (*Leaf)(unsafe.Pointer(args.result))
}

// MinMax Returns the minimum and maximum valued leaf
// @return The minimum and maximum leaf or NULL, NULL
func (r *ART) MinMax() (*Leaf, *Leaf) {
	args := artTreeMinmaxT{
		tree: uintptr(unsafe.Pointer(r)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_art_minmax), ptr, 0)
	return (*Leaf)(unsafe.Pointer(args.result)), (*Leaf)(unsafe.Pointer(args.result2))
}
