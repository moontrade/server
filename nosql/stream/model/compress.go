package model

import (
	"math"
	"sync"
	"unsafe"

	"github.com/pierrec/lz4/v4"
)

var (
	lz4Pool = &sync.Pool{New: func() interface{} {
		return &lz4Helper{}
	}}
)

var (
	MaxDataSize             = math.MaxUint16 - int(unsafe.Sizeof(BlockHeader{}))
	MaxDataSizeUncompressed = MaxDataSize - lz4.CompressBlockBound(MaxDataSize) - 1
)

type lz4Helper struct {
	buf [65536]byte
	ht  [65536]int
}

// Compress compresses the supplied buffer using LZ4
func Compress(b []byte) ([]byte, error) {
	if b == nil {
		return nil, ErrNotExist
	}
	c := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(b, c.buf[:65536], c.ht[:65536])
	lz4Pool.Put(c)
	if err != nil {
		return nil, err
	}
	out := make([]byte, n)
	//out := pbytes.GetLen(n)
	copy(out, c.buf[0:n])
	return out, nil
}

// Compress compresses the block using LZ4 and replaces current data buffer
func (b *BlockMessage) CompressInline() error {
	if b.BlockHeaderMut.Compression() != Compression_None || len(b.Body) == 0 {
		return nil
	}

	if len(b.Body) > MaxDataSizeUncompressed {
		return ErrOverflow
	}
	if len(b.Body) < MinCompressSize {
		return nil
	}
	m := b.Mut()
	m.SetSizeU(uint16(len(b.Body)))

	c := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(b.Data(), c.buf[0:], c.ht[0:])
	lz4Pool.Put(c)
	if err != nil {
		return err
	}
	// Was compression worth it?
	if n >= len(b.Body) {
		return nil
	}
	if n > MaxDataSize {
		return ErrOverflow
	}

	out := make([]byte, n)
	//out := pbytes.GetLen(n)
	copy(out, c.buf[0:n])
	//pbytes.Put(b.Body)

	m.SetCompression(Compression_LZ4).
		SetSizeX(uint16(n)).
		SetSize(m.SizeX())
	return nil
}

// Compress creates a LZ4 compressed clone if not compressed, otherwise
// it returns itself.
func (b *BlockMessage) Compress() (*BlockMessage, error) {
	if b.BlockHeaderMut.Compression() != Compression_None || len(b.Body) == 0 {
		return b, nil
	}

	if len(b.Body) > MaxDataSizeUncompressed {
		return nil, ErrOverflow
	}
	if len(b.Body) < MinCompressSize {
		return b, nil
	}

	helper := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(b.Data(), helper.buf[0:], helper.ht[0:])
	lz4Pool.Put(helper)
	if err != nil {
		return nil, err
	}
	// Was compression worth it?
	if n >= len(b.Body) {
		return b, nil
	}
	if n > MaxDataSize {
		return nil, ErrOverflow
	}

	m := GetBlockMessage(n)
	m.BlockHeaderMut = b.BlockHeaderMut
	copy(m.Body, helper.buf[0:n])
	m.SetCompression(Compression_LZ4).
		SetSizeX(uint16(n)).
		SetSize(m.SizeX())
	return m, nil
}

// DecompressInline decompresses and replaces current data buffer
func (b *BlockMessage) DecompressInline() error {
	switch b.BlockHeaderMut.Compression() {
	case Compression_None:
		return nil
	case Compression_LZ4:
		data := make([]byte, b.SizeU()) //pbytes.GetLen(int(b.SizeU()))
		n, err := lz4.UncompressBlock(b.Body, data)
		if err != nil {
			//pbytes.Put(data)
			return err
		}
		if len(data) != n {
			data = data[0:n]
		}
		//old := b.Body
		b.Body = data
		//pbytes.Put(old)
		b.BlockHeaderMut.SetSize(uint16(n)).SetCompression(Compression_None)
		return nil
	default:
		return ErrUnknownCompression
	}
}

// Decompress decompresses
func (b *BlockMessage) Decompress() (*BlockMessage, error) {
	switch b.BlockHeaderMut.Compression() {
	case Compression_None:
		return b, nil
	case Compression_LZ4:
		block := GetBlockMessage(int(b.SizeU()))
		block.BlockHeaderMut = b.BlockHeaderMut
		n, err := lz4.UncompressBlock(b.Body, block.Body)
		if err != nil {
			PutBlockMessage(block)
			return nil, err
		}
		if len(block.Body) != n {
			block.Body = block.Body[0:n]
		}

		block.BlockHeaderMut.SetSize(uint16(n)).SetCompression(Compression_None)
		return block, nil
	default:
		return nil, ErrUnknownCompression
	}
}

// CompressInline compresses the entire block using LZ4
func (r *RecordMessage) CompressInline() error {
	if r.RecordHeader.Compression() != Compression_None || len(r.Data) == 0 {
		return nil
	}

	if len(r.Data) > MaxDataSizeUncompressed {
		return ErrOverflow
	}
	if len(r.Data) < MinCompressSize {
		return nil
	}
	m := r.RecordHeader.Mut()
	m.SetSizeU(uint16(len(r.Data)))

	c := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(r.Data, c.buf[0:], c.ht[0:])
	lz4Pool.Put(c)
	if err != nil {
		return err
	}
	// Was compression worth it?
	if n >= len(r.Data) {
		return nil
	}
	if n > MaxDataSize {
		return ErrOverflow
	}

	out := make([]byte, n)
	//out := pbytes.GetLen(n)
	copy(out, c.buf[0:n])
	//pbytes.Put(r.Data)

	m.SetCompression(Compression_LZ4)
	m.SetSize(uint16(n))
	return nil
}

// CompressedClone creates a LZ4 compressed clone if not compressed, otherwise
// it returns itself.
func (r *RecordMessage) CompressedClone() (*RecordMessage, error) {
	if r.RecordHeader.Compression() != Compression_None || len(r.Data) == 0 {
		return r, nil
	}

	if len(r.Data) > MaxDataSizeUncompressed {
		return nil, ErrOverflow
	}
	if len(r.Data) < MinCompressSize {
		return r, nil
	}

	helper := lz4Pool.Get().(*lz4Helper)
	n, err := lz4.CompressBlock(r.Data, helper.buf[0:], helper.ht[0:])
	lz4Pool.Put(helper)
	if err != nil {
		return nil, err
	}
	// Was compression worth it?
	if n >= len(r.Data) {
		return r, nil
	}
	if n > MaxDataSize {
		return nil, ErrOverflow
	}

	m := GetRecord(n)
	m.RecordHeader = r.RecordHeader
	copy(m.Data, helper.buf[0:n])
	m.RecordHeader.Mut().SetCompression(Compression_LZ4).
		SetSizeX(uint16(n)).
		SetSize(m.RecordHeader.SizeX())
	return m, nil
}

func (r *RecordMessage) DecompressInline() error {
	data := r.Data
	switch r.RecordHeader.Compression() {
	case Compression_LZ4:
		b := make([]byte, r.RecordHeader.SizeU())
		//b := pbytes.GetLen(int(r.RecordHeader.SizeU()))
		n, err := lz4.UncompressBlock(data, b)
		if err != nil {
			return err
		}
		if len(b) != n {
			data = b[0:n]
		} else {
			data = b
		}
		//pbytes.Put(r.Data)
		r.Data = data
		r.RecordHeader.Mut().
			SetSize(uint16(len(data))).
			SetCompression(Compression_None)
	case Compression_None:
	default:
		return ErrUnknownCompression
	}
	return nil
}

func (r *RecordMessage) Decompress() (*RecordMessage, error) {
	switch r.RecordHeader.Compression() {
	case Compression_LZ4:
		record := GetRecord(int(r.RecordHeader.SizeU()))
		n, err := lz4.UncompressBlock(r.Data, record.Data)
		if err != nil {
			return nil, err
		}
		if len(record.Data) != n {
			record.Data = record.Data[0:n]
		}

		record.RecordHeader = r.RecordHeader
		record.RecordHeader.Mut().
			SetSize(uint16(n)).
			SetCompression(Compression_None)
		return record, nil
	case Compression_None:
		return r, nil
	default:
		return r, ErrUnknownCompression
	}
}
