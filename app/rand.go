package app

// Rand is a random number interface used by Machine
type Rand interface {
	Int() int
	Uint64() uint64
	Uint32() uint32
	Float64() float64
	Read([]byte) (n int, err error)
}

// #region -- pcg-family random number generator

func rincr(seed int64) int64 {
	return int64(uint64(seed)*6364136223846793005 + 1)
}

func rgen(seed int64) uint32 {
	state := uint64(seed)
	xorshifted := uint32(((state >> 18) ^ state) >> 27)
	rot := uint32(state >> 59)
	return (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
}
