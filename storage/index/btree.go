package index

/*
#cgo LDFLAGS: -L. -lstdc++
#cgo CXXFLAGS: -std=c++20 -I.
#include <stdlib.h>
#include <pthread.h>
#include "btree.h"

struct btree_record_string_t {
	void* key;
	void* data;
	int key_len;
};

struct btree_record_int64_t {
	int64_t key;
	void* data;
};

int btree_compare_record_int64_t(const void *a, const void *b, void *udata) {
	struct btree_record_int64_t* aa = (struct btree_record_int64_t*)a;
	struct btree_record_int64_t* bb = (struct btree_record_int64_t*)b;
   	return aa->key - bb->key;
}

typedef struct {
	size_t elsize;
	size_t max_items;
	int (*compare)(const void *a, const void *b, void *udata);
	size_t udata;
	size_t ptr;
} btree_new_t;

void do_btree_new(size_t arg0, size_t arg1) {
	btree_new_t* args = (btree_new_t*)(void*)arg0;
	args->ptr = (size_t)btree_new(args->elsize, args->max_items, args->compare, (void*)args->udata);
}

typedef struct {
	size_t tree;
	size_t item;
	size_t existing;
} btree_set_t;

void do_btree_set(size_t arg0, size_t arg1) {
	btree_set_t* args = (btree_set_t*)(void*)arg0;
	args->existing = (size_t)btree_set((struct btree*)(void*)args->tree, (void*)args->item);
}

typedef struct {
	size_t tree;
	size_t item;
	size_t hint;
	size_t existing;
} btree_set_hint_t;

void do_btree_set_hint(size_t arg0, size_t arg1) {
	btree_set_hint_t* args = (btree_set_hint_t*)(void*)arg0;
	args->existing = (size_t)btree_set_hint((struct btree*)(void*)args->tree, (void*)args->item, (uint64_t*)args->hint);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t result;
} btree_delete_t;

void do_btree_delete(size_t arg0, size_t arg1) {
	btree_delete_t* args = (btree_delete_t*)(void*)arg0;
	args->result = (size_t)btree_delete((struct btree*)(void*)args->tree, (void*)args->key);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t hint;
	size_t result;
} btree_delete_hint_t;

void do_btree_delete_hint(size_t arg0, size_t arg1) {
	btree_delete_hint_t* args = (btree_delete_hint_t*)(void*)arg0;
	args->result = (size_t)btree_delete_hint((struct btree*)(void*)args->tree, (void*)args->key, (uint64_t*)args->hint);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t result;
} btree_get_t;

void do_btree_get(size_t arg0, size_t arg1) {
	btree_get_t* args = (btree_get_t*)(void*)arg0;
	args->result = (size_t)btree_get((struct btree*)(void*)args->tree, (void*)args->key);
}

typedef struct {
	size_t tree;
	size_t key;
	size_t hint;
	size_t result;
} btree_get_hint_t;

void do_btree_get_hint(size_t arg0, size_t arg1) {
	btree_get_hint_t* args = (btree_get_hint_t*)(void*)arg0;
	args->result = (size_t)btree_get_hint((struct btree*)(void*)args->tree, (void*)args->key, (uint64_t*)args->hint);
}

void do_btree_free(size_t arg0, size_t arg1) {
	btree_free((struct btree*)(void*)arg0);
}

*/
import "C"
import (
	"github.com/moontrade/nogc/unsafecgo"
	"unsafe"
)

type BTree C.struct_btree
type BTreeI64 C.struct_btree

type BTreeRecordInt64 struct {
	Key   int64
	Value uintptr
}

func BTreeInt64Compare() uintptr {
	return uintptr(unsafe.Pointer(C.btree_compare_record_int64_t))
}

func NewBTree(elsize, maxItems, compare, udata uintptr) *BTree {
	args := struct {
		elsize   uintptr
		maxItems uintptr
		compare  uintptr
		udata    uintptr
		ptr      uintptr
	}{
		elsize:   elsize,
		maxItems: maxItems,
		compare:  compare,
		udata:    udata,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_new), ptr, 0)
	return (*BTree)(unsafe.Pointer(args.ptr))
}

func (m *BTree) Free() {
	ptr := uintptr(unsafe.Pointer(m))
	unsafecgo.NonBlocking((*byte)(C.do_btree_free), ptr, 0)
}

func (r *BTree) Set(key int64, value uintptr) uintptr {
	item := BTreeRecordInt64{
		Key:   key,
		Value: value,
	}
	args := struct {
		m    uintptr
		item uintptr
		old  uintptr
	}{
		m:    uintptr(unsafe.Pointer(r)),
		item: uintptr(unsafe.Pointer(&item)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_set), ptr, 0)
	return args.old
}

func (r *BTree) SetHint(key int64, value uintptr, hint *uint64) uintptr {
	item := BTreeRecordInt64{
		Key:   key,
		Value: value,
	}
	args := struct {
		m    uintptr
		item uintptr
		hint uintptr
		old  uintptr
	}{
		m:    uintptr(unsafe.Pointer(r)),
		item: uintptr(unsafe.Pointer(&item)),
		hint: uintptr(unsafe.Pointer(hint)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_set_hint), ptr, 0)
	return args.old
}

func (r *BTree) Delete(key int64) uintptr {
	args := struct {
		m      uintptr
		key    uintptr
		result uintptr
	}{
		m:   uintptr(unsafe.Pointer(r)),
		key: uintptr(unsafe.Pointer(&key)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_delete), ptr, 0)
	return args.result
}

func (r *BTree) DeleteHint(key int64, hint *uint64) uintptr {
	args := struct {
		m      uintptr
		key    uintptr
		hint   uintptr
		result uintptr
	}{
		m:    uintptr(unsafe.Pointer(r)),
		key:  uintptr(unsafe.Pointer(&key)),
		hint: uintptr(unsafe.Pointer(hint)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_delete_hint), ptr, 0)
	return args.result
}

func (r *BTree) Get(key int64) uintptr {
	args := struct {
		tree   uintptr
		key    uintptr
		result uintptr
	}{
		tree: uintptr(unsafe.Pointer(r)),
		key:  uintptr(unsafe.Pointer(&key)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_get), ptr, 0)
	return args.result
}

func (r *BTree) GetHint(key int64, hint *uint64) uintptr {
	args := struct {
		tree   uintptr
		key    uintptr
		hint   uintptr
		result uintptr
	}{
		tree: uintptr(unsafe.Pointer(r)),
		key:  uintptr(unsafe.Pointer(&key)),
		hint: uintptr(unsafe.Pointer(hint)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_btree_get_hint), ptr, 0)
	return args.result
}
