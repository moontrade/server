package compress

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/DataDog/zstd"
	"github.com/pierrec/lz4/v4"
)

func TestBar(t *testing.T) {
	fmt.Println(unsafe.Sizeof(Bar{}))
	//p, u, c := generatePage(pageSize)
	p, u, c := generatePage(1024*8 - 192)
	_ = p
	fmt.Println("records:	", len(p.Bars))
	fmt.Println("uncompressed:	", len(u))
	cc := encodePage(p, u)
	lz4Encoded := compress(cc)
	fmt.Printf("lz4:				%d %.2f\n", len(c), float32(len(c))/float32(len(u)))
	fmt.Printf("lz4 opt layout:		%d %.2f\n", len(lz4Encoded), float32(len(lz4Encoded))/float32(len(u)))

	for i := -10; i < 23; i++ {
		z := compressZSTDLevel(cc, i)
		zz := compressZSTDLevel(u, i)
		pad := ""
		if i >= 0 && i < 10 {
			pad = " "
		}
		fmt.Printf("zstd %s%d:			opt: %d %.2f    std: %d %.2f    improvement: %d %.2f\n", pad, i, len(z), float32(len(z))/float32(len(u)), len(zz), float32(len(zz))/float32(len(u)), len(zz)-len(z), float32(len(z))/float32(len(zz)))
	}
}

func BenchmarkCompress(b *testing.B) {
	pageSize := 8192 - 192
	_, cached, _ := generatePage(pageSize)

	genPage := func(pageSize int) []byte {
		//return generatePage(pageSize)
		return cached
	}

	b.Run("lz4", func(b *testing.B) {
		u := genPage(pageSize)
		dst := make([]byte, lz4.CompressBlockBound(len(u)))

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			compressTo(u, dst)
		}
	})
	b.Run("lz4 HC", func(b *testing.B) {
		u := genPage(pageSize)
		dst := make([]byte, lz4.CompressBlockBound(len(u)))

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			compressLZ4To(u, dst)
		}
	})
	b.Run("lz4 decompress", func(b *testing.B) {
		u := genPage(pageSize)
		dst := make([]byte, lz4.CompressBlockBound(len(u)))
		compressTo(u, dst)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lz4.UncompressBlock(dst, u)
		}
	})

	for _, i := range []int{-7, -2, -1, 0, 1} {
		b.Run("zstd "+strconv.Itoa(i), func(b *testing.B) {
			u := genPage(pageSize)
			dst := make([]byte, zstd.CompressBound(len(u)))

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				compressZSTDToLevel(u, dst, -7)
			}
		})

		b.Run("zstd decompress "+strconv.Itoa(i), func(b *testing.B) {
			u := genPage(pageSize)
			dst := make([]byte, zstd.CompressBound(len(u)))
			compressZSTDToLevel(u, dst, -7)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				zstd.Decompress(u, dst)
			}
		})
	}
}

type OHLC struct {
	Open, High, Low, Close float64
}
type Volume struct {
	Total, Buy, Sell float64
}
type Spread struct {
	Open, High, Low, Close, Avg float32
}

type Bar struct {
	Begin    int64
	Bid, Ask OHLC
	Volume   Volume
	Spread   Spread
}

const EPOCH = 946684800

func randomOHLC() OHLC {
	return OHLC{
		Open:  rand.Float64(),
		High:  rand.Float64(),
		Low:   rand.Float64(),
		Close: rand.Float64(),
	}
}

func emulatedOHLC() OHLC {
	open := rand.Float64() / 1000.0
	for open > 1000.0 {
		open /= 1000.0
	}
	if rand.Int()%2 == 0 {
		high := open + float64(rand.Intn(1000))
		low := open - float64(rand.Intn(100))
		closed := open + float64(rand.Intn(500))
		if closed > high {
			c := closed
			closed = high
			high = c
		}

		return OHLC{
			Open:  open,
			High:  high,
			Low:   low,
			Close: closed,
		}
	}
	high := open + float64(rand.Intn(100))
	low := open - float64(rand.Intn(1000))
	closed := open - float64(rand.Intn(500))
	if closed < low {
		c := closed
		closed = low
		low = c
	}

	return OHLC{
		Open:  open,
		High:  high,
		Low:   low,
		Close: closed,
	}
}

func emulatedAsk(bid OHLC) OHLC {
	diff := float64(rand.Intn(100))
	return OHLC{
		Open:  bid.Open + diff,
		High:  bid.High + diff,
		Low:   bid.Low + diff,
		Close: bid.Close + diff,
	}
}

func emulatedVolume() Volume {
	total := float64(rand.Intn(5000000))
	buys := float64(rand.Intn(100))
	if buys == 0 {
		return Volume{
			Total: total,
			Buy:   0,
			Sell:  total,
		}
	}

	buys = (buys / 100) * total
	return Volume{
		Total: total,
		Buy:   buys,
		Sell:  total - buys,
	}
}

func emulatedSpread() Spread {
	open := rand.Float32()
	if rand.Int()%2 == 0 {
		high := open + float32(rand.Intn(1000))
		low := open - float32(rand.Intn(100))
		closed := open + float32(rand.Intn(500))
		if closed > high {
			c := closed
			closed = high
			high = c
		}

		return Spread{
			Open:  open,
			High:  high,
			Low:   low,
			Close: closed,
			Avg:   (high + low) / 2,
		}
	}
	high := open + float32(rand.Intn(100))
	low := open - float32(rand.Intn(1000))
	closed := open - float32(rand.Intn(500))
	if closed < low {
		c := closed
		closed = low
		low = c
	}

	return Spread{
		Open:  open,
		High:  high,
		Low:   low,
		Close: closed,
		Avg:   (high + low) / 2,
	}
}

func emulatedBar() Bar {
	bid := emulatedOHLC()
	bar := Bar{
		Begin:  time.Now().Unix(),
		Bid:    bid,
		Ask:    emulatedAsk(bid),
		Volume: emulatedVolume(),
		Spread: emulatedSpread(),
	}
	return bar
}

func emulateNextBar(prev Bar) Bar {
	bar := prev
	bar.Begin = bar.Begin + 60

	spreadDiff := float32(rand.Intn(5))
	priceDiff := float64(rand.Intn(5))

	priceDiff += priceDiff / 5
	spreadDiff = float32(priceDiff / 5)

	volumeDiff := float64(rand.Intn(200000))
	volumeDiff = 0

	if rand.Int()%2 == 0 {
		spreadDiff = -spreadDiff
		priceDiff = -priceDiff
	}

	bar.Bid.Open += priceDiff
	bar.Bid.High += priceDiff
	bar.Bid.Low += priceDiff
	bar.Bid.Close += priceDiff
	bar.Spread.Open += spreadDiff
	bar.Spread.High += spreadDiff
	bar.Spread.Low += spreadDiff
	bar.Spread.Close += spreadDiff
	bar.Spread.Avg = (bar.Spread.High + bar.Spread.Low) / 2
	bar.Volume.Total += volumeDiff
	bar.Volume.Buy += volumeDiff / 2
	bar.Volume.Sell += volumeDiff / 2
	return bar
}

type PG struct {
	PageHeader
	Bars []Bar
}

type PageHeader struct {
	ID       uint64
	Database uint64
	StreamID uint64
	Created  int64
	Seq      uint64
	First    uint64
	Last     uint64
	Begin    int64
	End      int64
	Size     uint32
	Count    uint32
	Duration uint32
	Encoding uint32
}

func generatePage(pageSize int) (PG, []byte, []byte) {
	bars := make([]Bar, 0, 128)
	previous := emulatedBar()
	page := make([]byte, 0, pageSize)

	header := PageHeader{
		ID:       1,
		Database: 5,
		StreamID: 5,
		Created:  time.Now().Unix(),
		Seq:      1,
		First:    90,
		Last:     110,
		Begin:    previous.Begin,
		End:      previous.Begin,
		Size:     0,
		Count:    0,
		Duration: 0,
		Encoding: 0,
	}

	page = append(page, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&header)),
		Len:  int(unsafe.Sizeof(PageHeader{})),
		Cap:  int(unsafe.Sizeof(PageHeader{})),
	}))...)

	for unsafe.Sizeof(PageHeader{})+(uintptr(len(bars)+1)*unsafe.Sizeof(Bar{})) < uintptr(pageSize) {
		bar := emulateNextBar(previous)
		bars = append(bars, bar)
		previous = bar

		page = append(page, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(&bar)),
			Len:  int(unsafe.Sizeof(Bar{})),
			Cap:  int(unsafe.Sizeof(Bar{})),
		}))...)
	}

	return PG{PageHeader: header, Bars: bars}, page, compress(page)
}

func encodePage(pg PG, uncompressed []byte) []byte {
	page := make([]byte, len(uncompressed))

	copy(page, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&pg.PageHeader)),
		Len:  int(unsafe.Sizeof(PageHeader{})),
		Cap:  int(unsafe.Sizeof(PageHeader{})),
	})))

	const barSize = unsafe.Sizeof(Bar{})

	var (
		float32Segment = 4 * len(pg.Bars)
		float64Segment = 8 * len(pg.Bars)
		bidOpen        = float64Segment
		bidHigh        = bidOpen + float64Segment
		bidLow         = bidHigh + float64Segment
		bidClose       = bidLow + float64Segment
		askOpen        = bidClose + float64Segment
		askHigh        = askOpen + float64Segment
		askLow         = askHigh + float64Segment
		askClose       = askLow + float64Segment
		volumeTotal    = askClose + float64Segment
		volumeBuys     = volumeTotal + float64Segment
		volumeSells    = volumeBuys + float64Segment
		spreadOpen     = volumeSells + float64Segment
		spreadHigh     = spreadOpen + float32Segment
		spreadLow      = spreadHigh + float32Segment
		spreadClose    = spreadLow + float32Segment
		spreadAvg      = spreadClose + float32Segment
	)

	for i, bar := range pg.Bars {
		binary.LittleEndian.PutUint64(page[i*8:], uint64(bar.Begin))
		binary.LittleEndian.PutUint64(page[bidOpen+(i*8):], uint64(bar.Bid.Open))
		binary.LittleEndian.PutUint64(page[bidHigh+(i*8):], uint64(bar.Bid.High))
		binary.LittleEndian.PutUint64(page[bidLow+(i*8):], uint64(bar.Bid.Low))
		binary.LittleEndian.PutUint64(page[bidClose+(i*8):], uint64(bar.Bid.Close))
		binary.LittleEndian.PutUint64(page[askOpen+(i*8):], uint64(bar.Ask.Open))
		binary.LittleEndian.PutUint64(page[askHigh+(i*8):], uint64(bar.Ask.High))
		binary.LittleEndian.PutUint64(page[askLow+(i*8):], uint64(bar.Ask.Low))
		binary.LittleEndian.PutUint64(page[askClose+(i*8):], uint64(bar.Ask.Close))
		binary.LittleEndian.PutUint64(page[volumeTotal+(i*8):], uint64(bar.Volume.Total))
		binary.LittleEndian.PutUint64(page[volumeBuys+(i*8):], uint64(bar.Volume.Buy))
		binary.LittleEndian.PutUint64(page[volumeSells+(i*8):], uint64(bar.Volume.Sell))
		binary.LittleEndian.PutUint64(page[spreadOpen+(i*8):], uint64(bar.Spread.Open))
		binary.LittleEndian.PutUint64(page[spreadHigh+(i*8):], uint64(bar.Spread.High))
		binary.LittleEndian.PutUint64(page[spreadLow+(i*8):], uint64(bar.Spread.Low))
		binary.LittleEndian.PutUint64(page[spreadClose+(i*8):], uint64(bar.Spread.Close))
		binary.LittleEndian.PutUint64(page[spreadAvg+(i*8):], uint64(bar.Spread.Avg))
	}

	return page
}

func compress(block []byte) []byte {
	dst := make([]byte, lz4.CompressBlockBound(len(block)))
	n, err := lz4.CompressBlock(block, dst, nil)
	//n, err := lz4.CompressBlockHC(block, dst, lz4.Fast, nil, nil)
	if err != nil {
		panic(err)
	}
	return dst[0:n]
}

func compressZSTD(block []byte) []byte {
	out, err := zstd.CompressLevel(nil, block, -10) //zstd.BestSpeed)
	if err != nil {
		panic(err)
	}
	return out
}

func compressZSTDLevel(block []byte, level int) []byte {
	out, err := zstd.CompressLevel(nil, block, level)
	if err != nil {
		panic(err)
	}
	return out
}

func compressTo(block, dst []byte) []byte {
	n, err := lz4.CompressBlock(block, dst, nil)
	if err != nil {
		panic(err)
	}
	return dst[0:n]
}

func compressLZ4To(block, dst []byte) []byte {
	n, err := lz4.CompressBlockHC(block, dst, lz4.Fast, nil, nil)
	if err != nil {
		panic(err)
	}
	return dst[0:n]
}

func compressZSTDTo(block, dst []byte) []byte {
	out, err := zstd.CompressLevel(dst, block, -10) // zstd.BestSpeed)
	if err != nil {
		panic(err)
	}
	return out
}

func compressZSTDToLevel(block, dst []byte, level int) []byte {
	out, err := zstd.CompressLevel(dst, block, level) // zstd.BestSpeed)
	if err != nil {
		panic(err)
	}
	return out
}
