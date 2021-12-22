//go:build 386 || amd64 || arm || arm64 || ppc64le || mips64le || mipsle || riscv64 || wasm
// +build 386 amd64 arm arm64 ppc64le mips64le mipsle riscv64 wasm

package model

import (
	"crypto/rand"
	"unsafe"
)

type StreamID int64

const mask63 = StreamID(int64(^(uint64(1) << 63)))
const mask48 = ^(int64(1) << 48)
const mask40 = ^(int64(1) << 40)
const mask32 = ^(int64(1) << 32)
const mask15 = int16(^(uint16(1) << 15))

func (s StreamID) DB() int16 {
	return int16(byte(s>>48)) | int16(byte(s>>56))<<8
}

func (s StreamID) ID() int64 {
	return int64(s) & mask48
}

func NewStreamID(shard int16, id int64) StreamID {
	b := *(*[8]byte)(unsafe.Pointer(&id))
	shard &= mask15
	b[6] = byte(shard)
	b[7] = byte(shard >> 8)
	return *(*StreamID)(unsafe.Pointer(&b))
}

func GenerateStreamID(shard int16) StreamID {
	var b [8]byte
	_, err := rand.Read(b[:6])
	if err != nil {
		panic(err)
	}

	// clear sign bit
	shard &= mask15
	// little-endian
	b[6] = byte(shard)
	b[7] = byte(shard >> 8)

	return *(*StreamID)(unsafe.Pointer(&b))
}
