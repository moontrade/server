package model

import "sync"

type OnBlockMessageEvicted func(block *BlockMessage)

type BlockMessageBuffer struct {
	buffer    []*BlockMessage
	length    int
	counter   int
	size      int // Total size in bytes of all BlockMessages in the buffer
	maxBytes  int
	onEvicted OnBlockMessageEvicted
	mu        sync.Mutex
}

func NewBlockMessageBuffer(maxLength, maxSize int, evicted OnBlockMessageEvicted) *BlockMessageBuffer {
	return &BlockMessageBuffer{
		onEvicted: evicted,
		buffer:    make([]*BlockMessage, maxLength, maxLength),
		length:    0,
		counter:   0,
		size:      sizeofBlockMessageBuffer + 24 + (8 * maxLength),
		maxBytes:  maxSize,
		mu:        sync.Mutex{},
	}
}

func (b *BlockMessageBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
}
func (b *BlockMessageBuffer) clear0() {
	if b.length == 0 {
		return
	}
	for i, block := range b.buffer {
		if block != nil {
			b.buffer[i] = nil
			if b.onEvicted != nil {
				b.onEvicted(block)
			}
			b.size -= block.Sizeof()
		}
	}
}

func (b *BlockMessageBuffer) IsFull() bool {
	return b.length == len(b.buffer)
}

func (b *BlockMessageBuffer) Copy(to []*BlockMessage) []*BlockMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		return nil
	}
	for i := b.counter - b.length; i < b.counter; i++ {
		to = append(to, b.buffer[i])
	}
	return to
}

// PushSavepoint Clears all blocks <= supplied Block ID
func (b *BlockMessageBuffer) Savepoint(id int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var (
		index  int
		block  *BlockMessage
		length = b.length
	)
	for i := length; i > -1; i-- {
		index = (b.counter - i) % len(b.buffer)
		block = b.buffer[index]
		if block == nil {
			b.length--
			continue
		}
		if block.MessageID() <= id {
			b.buffer[index] = nil
			if b.onEvicted != nil {
				b.onEvicted(block)
			}
			if block.MessageID() <= id {
				continue
			}
		}

		break
	}
}

func (b *BlockMessageBuffer) First() *BlockMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		return nil
	}
	index := (b.counter - b.length) % len(b.buffer)
	return b.buffer[index]
}

func (b *BlockMessageBuffer) PopFirst() *BlockMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		return nil
	}
	index := (b.counter - b.length) % len(b.buffer)
	b.length--
	block := b.buffer[index]
	b.buffer[index] = nil
	if block != nil {
		b.size -= block.Sizeof()
	}
	return block
}

func (b *BlockMessageBuffer) Last() *BlockMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		return nil
	}
	index := (b.counter - 1) % len(b.buffer)
	return b.buffer[index]
}

func (b *BlockMessageBuffer) PopLast() *BlockMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		return nil
	}
	index := (b.counter - 1) % len(b.buffer)
	b.length--
	block := b.buffer[index]
	b.buffer[index] = nil
	if block != nil {
		b.size -= block.Sizeof()
	}
	return block
}

func (b *BlockMessageBuffer) PushSequential(block *BlockMessage) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == 0 {
		b.push0(block)
		return true
	}
	last := b.buffer[(b.counter-1)%len(b.buffer)]
	if last == nil {
		b.push0(block)
		return true
	}
	if last.MessageID() != block.MessageID()-1 {
		b.clear0()
		return false
	}
	b.push0(block)
	return true
}

func (b *BlockMessageBuffer) Push(block *BlockMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.push0(block)
}

func (b *BlockMessageBuffer) push0(block *BlockMessage) {
	if block == nil {
		return
	}

	index := b.counter % len(b.buffer)
	existing := b.buffer[index]
	if existing != nil {
		b.size -= existing.Sizeof()
		if b.onEvicted != nil {
			b.onEvicted(existing)
		}
	} else {
		b.length++
	}
	b.size += block.Sizeof()
	b.buffer[index] = block

	b.counter++

	// Prune or compress blocks until maxBytes is satisfied
	for b.maxBytes > 0 && b.size > b.maxBytes && b.length > 1 {
		index = (b.counter - b.length) % len(b.buffer)
		existing = b.buffer[index]
		if existing == nil {
			break
		}
		size := existing.Sizeof()
		compressed, _ := existing.Compress()
		sizeAfter := compressed.Sizeof()

		if sizeAfter < size && b.size-(size-sizeAfter) <= b.maxBytes {
			b.buffer[index] = compressed
			b.size -= size - sizeAfter
			break
		}

		b.length--
		if b.onEvicted != nil {
			b.onEvicted(existing)
		}
		b.size -= size
	}
}

func (b *BlockMessageBuffer) PushNoEvict(block *BlockMessage) bool {
	if block == nil {
		return true
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.length == len(b.buffer) {
		return false
	}
	b.buffer[b.counter%len(b.buffer)] = block
	b.length++
	b.counter++
	return true
}
