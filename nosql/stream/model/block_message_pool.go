package model

import (
	"sync"
)

var (
	//defaultBlockMessagePool = newBlockMessagePool(128, 65536)
	blockMessagePool = &sync.Pool{New: func() interface{} {
		return &BlockMessage{}
	}}
)

func GetBlockMessage(size int) *BlockMessage {
	b := blockMessagePool.Get().(*BlockMessage)
	if size > 0 {
		b.Body = make([]byte, size)
		//b.Body = pbytes.GetLen(size)
	}
	b.BlockHeaderMut = BlockHeaderMut{}
	return b
}

func GetBlockMessageWith(data []byte) *BlockMessage {
	b := blockMessagePool.Get().(*BlockMessage)
	if len(data) > 0 {
		b.Body = make([]byte, len(data))
		//b.Body = pbytes.GetLen(len(data))
		copy(b.Body, data)
	}
	b.BlockHeaderMut = BlockHeaderMut{}
	return b
}

func PutBlockMessage(block *BlockMessage) {
	//d := block.Body
	block.Body = nil
	//pbytes.Put(d)
	blockMessagePool.Put(block)
}

//// Pool contains logic of reusing byte slices of various size.
//type blockMessagePool struct {
//	pool *pool.Pool
//}
//
//// New creates new Pool that reuses slices which size is in logarithmic range
//// [min, max].
////
//// Note that it is a shortcut for Custom() constructor with Options provided by
//// pool.WithLogSizeMapping() and pool.WithLogSizeRange(min, max) calls.
//func newBlockMessagePool(min, max int) *blockMessagePool {
//	return &blockMessagePool{pool.New(min, max)}
//}
//
//// New creates new Pool with given options.
////func Custom(opts ...pool.Option) *BlockMessagePool {
////	return &BlockMessagePool{pool.Custom(opts...)}
////}
//
//// Get returns probably reused slice of bytes with at least capacity of c and
//// exactly len of n.
//func (p *blockMessagePool) Get(n, c int) *BlockMessage {
//	if n > c {
//		panic("requested length is greater than capacity")
//	}
//
//	v, x := p.pool.Get(c)
//	if v != nil {
//		bts := v.(*BlockMessage)
//		bts.data = bts.data[:n]
//		return bts
//	}
//
//	return &BlockMessage{
//		data: make([]byte, x, c),
//	}
//}
//
//// Put returns given slice to reuse pool.
//// It does not reuse bytes whose size is not power of two or is out of pool
//// min/max range.
//func (p *blockMessagePool) Put(bts *BlockMessage) {
//	p.pool.Put(bts, cap(bts.data))
//}
//
//// GetCap returns probably reused slice of bytes with at least capacity of n.
//func (p *blockMessagePool) GetCap(c int) *BlockMessage {
//	return p.Get(0, c)
//}
//
//// GetLen returns probably reused slice of bytes with at least capacity of n
//// and exactly len of n.
//func (p *blockMessagePool) GetLen(n int) *BlockMessage {
//	return p.Get(n, n)
//}
