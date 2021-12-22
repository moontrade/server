package model

func GetBlockMut(size int) BlockMut {
	if size > 64 {
		if size > 32768 {
			return GetBlock64Mut()
		}
		if size > 16384 {
			return GetBlock32Mut()
		}
		if size > 8192 {
			return GetBlock16Mut()
		}
		if size > 4096 {
			return GetBlock8Mut()
		}
		if size > 2048 {
			return GetBlock4Mut()
		}
		if size > 1024 {
			return GetBlock2Mut()
		}
		return GetBlock1Mut()
	}

	if size > 32 {
		return GetBlock64Mut()
	}
	if size > 16 {
		return GetBlock32Mut()
	}
	if size >= 8 {
		return GetBlock16Mut()
	}
	if size > 4 {
		return GetBlock8Mut()
	}
	if size > 2 {
		return GetBlock4Mut()
	}
	if size > 1 {
		return GetBlock2Mut()
	}
	return GetBlock1Mut()
}

func PutBlockMut(b BlockMut) {
	if b == nil {
		return
	}
	switch t := b.(type) {
	case *Block64Mut:
		block64Pool.Put(t)

	case *Block32Mut:
		block32Pool.Put(t)

	case *Block16Mut:
		block16Pool.Put(t)

	case *Block8Mut:
		block8Pool.Put(t)

	case *Block4Mut:
		block4Pool.Put(t)

	case *Block2Mut:
		block2Pool.Put(t)

	case *Block1Mut:
		block1Pool.Put(t)
	}
}

func GetBlock1Mut() *Block1Mut {
	return block1Pool.Get().(*Block1Mut)
}

func PutBlock1Mut(b *Block1Mut) {
	// Clear it
	*b = Block1Mut{}
	block1Pool.Put(b)
}

func GetBlock2Mut() *Block2Mut {
	return block2Pool.Get().(*Block2Mut)
}

func PutBlock2Mut(b *Block2Mut) {
	// Clear it
	*b = Block2Mut{}
	block2Pool.Put(b)
}

func GetBlock4Mut() *Block4Mut {
	return block4Pool.Get().(*Block4Mut)
}

func PutBlock4Mut(b *Block4Mut) {
	// Clear it
	*b = Block4Mut{}
	block4Pool.Put(b)
}

func GetBlock8Mut() *Block8Mut {
	return block8Pool.Get().(*Block8Mut)
}

func PutBlock8Mut(b *Block8Mut) {
	// Clear it
	*b = Block8Mut{}
	block8Pool.Put(b)
}

func GetBlock16Mut() *Block16Mut {
	return block16Pool.Get().(*Block16Mut)
}

func PutBlock16Mut(b *Block16Mut) {
	// Clear it
	*b = Block16Mut{}
	block16Pool.Put(b)
}

func GetBlock32Mut() *Block32Mut {
	return block32Pool.Get().(*Block32Mut)
}

func PutBlock32Mut(b *Block32Mut) {
	// Clear it
	*b = Block32Mut{}
	block32Pool.Put(b)
}

func GetBlock64Mut() *Block64Mut {
	return block64Pool.Get().(*Block64Mut)
}

func PutBlock64Mut(b *Block64Mut) {
	// Clear it
	*b = Block64Mut{}
	block64Pool.Put(b)
}

var (
	PooledBlock1Allocator  BlockAllocator = block1Allocator{}
	PooledBlock2Allocator  BlockAllocator = block2Allocator{}
	PooledBlock4Allocator  BlockAllocator = block4Allocator{}
	PooledBlock8Allocator  BlockAllocator = block8Allocator{}
	PooledBlock16Allocator BlockAllocator = block16Allocator{}
	PooledBlock32Allocator BlockAllocator = block32Allocator{}
	PooledBlock64Allocator BlockAllocator = block64Allocator{}
)

func GetBlockAllocator(size int) BlockAllocator {
	switch size {
	case 64:
		return PooledBlock64Allocator
	case 32:
		return PooledBlock32Allocator
	case 16:
		return PooledBlock16Allocator
	case 8:
		return PooledBlock8Allocator
	case 4:
		return PooledBlock4Allocator
	case 2:
		return PooledBlock2Allocator
	case 1:
		return PooledBlock1Allocator
	}
	return PooledBlock64Allocator
}

type block1Allocator struct{}

func (block1Allocator) Alloc() BlockMut {
	return GetBlock1Mut()
}

func (block1Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block2Allocator struct{}

func (block2Allocator) Alloc() BlockMut {
	return GetBlock2Mut()
}

func (block2Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block4Allocator struct{}

func (block4Allocator) Alloc() BlockMut {
	return GetBlock4Mut()
}

func (block4Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block8Allocator struct{}

func (block8Allocator) Alloc() BlockMut {
	return GetBlock8Mut()
}

func (block8Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block16Allocator struct{}

func (block16Allocator) Alloc() BlockMut {
	return GetBlock16Mut()
}

func (block16Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block32Allocator struct{}

func (block32Allocator) Alloc() BlockMut {
	return GetBlock32Mut()
}

func (block32Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}

type block64Allocator struct{}

func (block64Allocator) Alloc() BlockMut {
	return GetBlock64Mut()
}

func (block64Allocator) Release(b BlockMut) {
	PutBlockMut(b)
}
