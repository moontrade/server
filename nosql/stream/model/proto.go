//go:build 386 || amd64 || arm || arm64 || ppc64le || mips64le || mipsle || riscv64 || wasm
// +build 386 amd64 arm arm64 ppc64le mips64le mipsle riscv64 wasm

package model

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

type Compression byte

const (
	Compression_None = Compression(0)

	Compression_LZ4 = Compression(1)
)

type StreamKind byte

const (
	StreamKind_Log = StreamKind(0)

	StreamKind_TimeSeries = StreamKind(1)

	StreamKind_Table = StreamKind(2)
)

type MessageType byte

const (
	MessageType_Record = MessageType(1)

	MessageType_Block = MessageType(2)

	MessageType_EOS = MessageType(3)

	MessageType_EOB = MessageType(4)

	MessageType_Savepoint = MessageType(5)

	MessageType_Starting = MessageType(6)

	MessageType_Progress = MessageType(7)

	MessageType_Started = MessageType(8)

	MessageType_Stopped = MessageType(9)
)

type StopReason byte

const (
	// Stream is composed from another stream or external datasource and it stopped
	StopReason_Source = StopReason(1)

	// Stream has been paused
	StopReason_Paused = StopReason(2)

	// Stream is being migrated to a new writer
	StopReason_Migrate = StopReason(3)

	// Stream has stopped unexpectedly
	StopReason_Error = StopReason(4)
)

type SchemaKind byte

const (
	SchemaKind_Bytes = SchemaKind(0)

	SchemaKind_MoonBuf = SchemaKind(1)

	SchemaKind_ProtoBuf = SchemaKind(2)

	SchemaKind_FlatBuffers = SchemaKind(3)

	SchemaKind_Json = SchemaKind(4)

	SchemaKind_MessagePack = SchemaKind(5)
)

// End of Block
type EOB struct {
	recordID  RecordID
	timestamp int64
	savepoint int64
}

func (s *EOB) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *EOB) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["savepoint"] = s.Savepoint()
	return m
}

func (s *EOB) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[40]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 40 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *EOB) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[40]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *EOB) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *EOB) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *EOB) Read(b []byte) (n int, err error) {
	if len(b) < 40 {
		return -1, io.ErrShortBuffer
	}
	v := (*EOB)(unsafe.Pointer(&b[0]))
	*v = *s
	return 40, nil
}
func (s *EOB) UnmarshalBinary(b []byte) error {
	if len(b) < 40 {
		return io.ErrShortBuffer
	}
	v := (*EOB)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *EOB) Clone() *EOB {
	v := &EOB{}
	*v = *s
	return v
}
func (s *EOB) Bytes() []byte {
	return (*(*[40]byte)(unsafe.Pointer(s)))[0:]
}
func (s *EOB) Mut() *EOBMut {
	return (*EOBMut)(unsafe.Pointer(s))
}
func (s *EOB) RecordID() *RecordID {
	return &s.recordID
}
func (s *EOB) Timestamp() int64 {
	return s.timestamp
}
func (s *EOB) Savepoint() int64 {
	return s.savepoint
}

// End of Block
type EOBMut struct {
	EOB
}

func (s *EOBMut) Clone() *EOBMut {
	v := &EOBMut{}
	*v = *s
	return v
}
func (s *EOBMut) Freeze() *EOB {
	return (*EOB)(unsafe.Pointer(s))
}
func (s *EOBMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *EOBMut) SetRecordID(v *RecordID) *EOBMut {
	s.recordID = *v
	return s
}
func (s *EOBMut) SetTimestamp(v int64) *EOBMut {
	s.timestamp = v
	return s
}
func (s *EOBMut) SetSavepoint(v int64) *EOBMut {
	s.savepoint = v
	return s
}

type RecordHeader struct {
	streamID    int64
	blockID     int64
	id          int64
	timestamp   int64
	start       int64
	end         int64
	savepoint   int64
	savepointR  int64
	seq         uint16
	size        uint16
	sizeU       uint16
	sizeX       uint16
	compression Compression
	eob         bool
	_           [6]byte // Padding
}

func (s *RecordHeader) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *RecordHeader) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["streamID"] = s.StreamID()
	m["blockID"] = s.BlockID()
	m["id"] = s.MessageID()
	m["timestamp"] = s.Timestamp()
	m["start"] = s.Start()
	m["end"] = s.End()
	m["savepoint"] = s.Savepoint()
	m["savepointR"] = s.SavepointR()
	m["seq"] = s.Seq()
	m["size"] = s.Size()
	m["sizeU"] = s.SizeU()
	m["sizeX"] = s.SizeX()
	m["compression"] = s.Compression()
	m["eob"] = s.Eob()
	return m
}

func (s *RecordHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[80]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 80 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *RecordHeader) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[80]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *RecordHeader) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[80]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *RecordHeader) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[80]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *RecordHeader) Read(b []byte) (n int, err error) {
	if len(b) < 80 {
		return -1, io.ErrShortBuffer
	}
	v := (*RecordHeader)(unsafe.Pointer(&b[0]))
	*v = *s
	return 80, nil
}
func (s *RecordHeader) UnmarshalBinary(b []byte) error {
	if len(b) < 80 {
		return io.ErrShortBuffer
	}
	v := (*RecordHeader)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *RecordHeader) Clone() *RecordHeader {
	v := &RecordHeader{}
	*v = *s
	return v
}
func (s *RecordHeader) Bytes() []byte {
	return (*(*[80]byte)(unsafe.Pointer(s)))[0:]
}
func (s *RecordHeader) Mut() *RecordHeaderMut {
	return (*RecordHeaderMut)(unsafe.Pointer(s))
}
func (s *RecordHeader) StreamID() int64 {
	return s.streamID
}
func (s *RecordHeader) BlockID() int64 {
	return s.blockID
}
func (s *RecordHeader) MessageID() int64 {
	return s.id
}
func (s *RecordHeader) Timestamp() int64 {
	return s.timestamp
}
func (s *RecordHeader) Start() int64 {
	return s.start
}
func (s *RecordHeader) End() int64 {
	return s.end
}
func (s *RecordHeader) Savepoint() int64 {
	return s.savepoint
}
func (s *RecordHeader) SavepointR() int64 {
	return s.savepointR
}
func (s *RecordHeader) Seq() uint16 {
	return s.seq
}
func (s *RecordHeader) Size() uint16 {
	return s.size
}
func (s *RecordHeader) SizeU() uint16 {
	return s.sizeU
}
func (s *RecordHeader) SizeX() uint16 {
	return s.sizeX
}
func (s *RecordHeader) Compression() Compression {
	return s.compression
}
func (s *RecordHeader) Eob() bool {
	return s.eob
}

type RecordHeaderMut struct {
	RecordHeader
}

func (s *RecordHeaderMut) Clone() *RecordHeaderMut {
	v := &RecordHeaderMut{}
	*v = *s
	return v
}
func (s *RecordHeaderMut) Freeze() *RecordHeader {
	return (*RecordHeader)(unsafe.Pointer(s))
}
func (s *RecordHeaderMut) SetStreamID(v int64) *RecordHeaderMut {
	s.streamID = v
	return s
}
func (s *RecordHeaderMut) SetBlockID(v int64) *RecordHeaderMut {
	s.blockID = v
	return s
}
func (s *RecordHeaderMut) SetId(v int64) *RecordHeaderMut {
	s.id = v
	return s
}
func (s *RecordHeaderMut) SetTimestamp(v int64) *RecordHeaderMut {
	s.timestamp = v
	return s
}
func (s *RecordHeaderMut) SetStart(v int64) *RecordHeaderMut {
	s.start = v
	return s
}
func (s *RecordHeaderMut) SetEnd(v int64) *RecordHeaderMut {
	s.end = v
	return s
}
func (s *RecordHeaderMut) SetSavepoint(v int64) *RecordHeaderMut {
	s.savepoint = v
	return s
}
func (s *RecordHeaderMut) SetSavepointR(v int64) *RecordHeaderMut {
	s.savepointR = v
	return s
}
func (s *RecordHeaderMut) SetSeq(v uint16) *RecordHeaderMut {
	s.seq = v
	return s
}
func (s *RecordHeaderMut) SetSize(v uint16) *RecordHeaderMut {
	s.size = v
	return s
}
func (s *RecordHeaderMut) SetSizeU(v uint16) *RecordHeaderMut {
	s.sizeU = v
	return s
}
func (s *RecordHeaderMut) SetSizeX(v uint16) *RecordHeaderMut {
	s.sizeX = v
	return s
}
func (s *RecordHeaderMut) SetCompression(v Compression) *RecordHeaderMut {
	s.compression = v
	return s
}
func (s *RecordHeaderMut) SetEob(v bool) *RecordHeaderMut {
	s.eob = v
	return s
}

// BlockID represents a globally unique ID of a single page of a single stream.
// String representation
type BlockID struct {
	streamID int64
	id       int64
}

func (s *BlockID) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *BlockID) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["streamID"] = s.StreamID()
	m["id"] = s.Id()
	return m
}

func (s *BlockID) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[16]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 16 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *BlockID) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[16]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *BlockID) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[16]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *BlockID) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[16]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *BlockID) Read(b []byte) (n int, err error) {
	if len(b) < 16 {
		return -1, io.ErrShortBuffer
	}
	v := (*BlockID)(unsafe.Pointer(&b[0]))
	*v = *s
	return 16, nil
}
func (s *BlockID) UnmarshalBinary(b []byte) error {
	if len(b) < 16 {
		return io.ErrShortBuffer
	}
	v := (*BlockID)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *BlockID) Clone() *BlockID {
	v := &BlockID{}
	*v = *s
	return v
}
func (s *BlockID) Bytes() []byte {
	return (*(*[16]byte)(unsafe.Pointer(s)))[0:]
}
func (s *BlockID) Mut() *BlockIDMut {
	return (*BlockIDMut)(unsafe.Pointer(s))
}
func (s *BlockID) StreamID() int64 {
	return s.streamID
}
func (s *BlockID) Id() int64 {
	return s.id
}

// BlockID represents a globally unique ID of a single page of a single stream.
// String representation
type BlockIDMut struct {
	BlockID
}

func (s *BlockIDMut) Clone() *BlockIDMut {
	v := &BlockIDMut{}
	*v = *s
	return v
}
func (s *BlockIDMut) Freeze() *BlockID {
	return (*BlockID)(unsafe.Pointer(s))
}
func (s *BlockIDMut) SetStreamID(v int64) *BlockIDMut {
	s.streamID = v
	return s
}
func (s *BlockIDMut) SetId(v int64) *BlockIDMut {
	s.id = v
	return s
}

type Stream struct {
	id        int64
	created   int64
	accountID int64
	duration  int64
	name      String32
	record    int32
	kind      StreamKind
	schema    SchemaKind
	realTime  bool
	blockSize byte
}

func (s *Stream) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Stream) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["id"] = s.Id()
	m["created"] = s.Created()
	m["accountID"] = s.AccountID()
	m["duration"] = s.Duration()
	m["name"] = s.Name()
	m["record"] = s.Record()
	m["kind"] = s.Kind()
	m["schema"] = s.Schema()
	m["realTime"] = s.RealTime()
	m["blockSize"] = s.BlockSize()
	return m
}

func (s *Stream) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[72]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 72 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Stream) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[72]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Stream) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[72]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Stream) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[72]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Stream) Read(b []byte) (n int, err error) {
	if len(b) < 72 {
		return -1, io.ErrShortBuffer
	}
	v := (*Stream)(unsafe.Pointer(&b[0]))
	*v = *s
	return 72, nil
}
func (s *Stream) UnmarshalBinary(b []byte) error {
	if len(b) < 72 {
		return io.ErrShortBuffer
	}
	v := (*Stream)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Stream) Clone() *Stream {
	v := &Stream{}
	*v = *s
	return v
}
func (s *Stream) Bytes() []byte {
	return (*(*[72]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Stream) Mut() *StreamMut {
	return (*StreamMut)(unsafe.Pointer(s))
}
func (s *Stream) Id() int64 {
	return s.id
}
func (s *Stream) Created() int64 {
	return s.created
}
func (s *Stream) AccountID() int64 {
	return s.accountID
}
func (s *Stream) Duration() int64 {
	return s.duration
}
func (s *Stream) Name() *String32 {
	return &s.name
}
func (s *Stream) Record() int32 {
	return s.record
}
func (s *Stream) Kind() StreamKind {
	return s.kind
}
func (s *Stream) Schema() SchemaKind {
	return s.schema
}
func (s *Stream) RealTime() bool {
	return s.realTime
}
func (s *Stream) BlockSize() byte {
	return s.blockSize
}

type StreamMut struct {
	Stream
}

func (s *StreamMut) Clone() *StreamMut {
	v := &StreamMut{}
	*v = *s
	return v
}
func (s *StreamMut) Freeze() *Stream {
	return (*Stream)(unsafe.Pointer(s))
}
func (s *StreamMut) SetId(v int64) *StreamMut {
	s.id = v
	return s
}
func (s *StreamMut) SetCreated(v int64) *StreamMut {
	s.created = v
	return s
}
func (s *StreamMut) SetAccountID(v int64) *StreamMut {
	s.accountID = v
	return s
}
func (s *StreamMut) SetDuration(v int64) *StreamMut {
	s.duration = v
	return s
}
func (s *StreamMut) Name() *String32Mut {
	return s.name.Mut()
}
func (s *StreamMut) SetName(v *String32) *StreamMut {
	s.name = *v
	return s
}
func (s *StreamMut) SetRecord(v int32) *StreamMut {
	s.record = v
	return s
}
func (s *StreamMut) SetKind(v StreamKind) *StreamMut {
	s.kind = v
	return s
}
func (s *StreamMut) SetSchema(v SchemaKind) *StreamMut {
	s.schema = v
	return s
}
func (s *StreamMut) SetRealTime(v bool) *StreamMut {
	s.realTime = v
	return s
}
func (s *StreamMut) SetBlockSize(v byte) *StreamMut {
	s.blockSize = v
	return s
}

// Block64
type Block64 struct {
	head BlockHeader
	body Bytes65456
}

func (s *Block64) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block64) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block64) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[65552]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 65552 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block64) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[65552]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block64) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[65552]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block64) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[65552]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block64) Read(b []byte) (n int, err error) {
	if len(b) < 65552 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block64)(unsafe.Pointer(&b[0]))
	*v = *s
	return 65552, nil
}
func (s *Block64) UnmarshalBinary(b []byte) error {
	if len(b) < 65552 {
		return io.ErrShortBuffer
	}
	v := (*Block64)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block64) Clone() *Block64 {
	v := &Block64{}
	*v = *s
	return v
}
func (s *Block64) Bytes() []byte {
	return (*(*[65552]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block64) Mut() *Block64Mut {
	return (*Block64Mut)(unsafe.Pointer(s))
}
func (s *Block64) Head() *BlockHeader {
	return &s.head
}
func (s *Block64) Body() *Bytes65456 {
	return &s.body
}

// Block64
type Block64Mut struct {
	Block64
}

func (s *Block64Mut) Clone() *Block64Mut {
	v := &Block64Mut{}
	*v = *s
	return v
}
func (s *Block64Mut) Freeze() *Block64 {
	return (*Block64)(unsafe.Pointer(s))
}
func (s *Block64Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block64Mut) SetHead(v *BlockHeader) *Block64Mut {
	s.head = *v
	return s
}
func (s *Block64Mut) SetBody(v Bytes65456) *Block64Mut {
	s.body = v
	return s
}

type Savepoint struct {
	recordID  RecordID
	timestamp int64
	writerID  int64
}

func (s *Savepoint) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Savepoint) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["writerID"] = s.WriterID()
	return m
}

func (s *Savepoint) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[40]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 40 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Savepoint) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[40]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Savepoint) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Savepoint) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Savepoint) Read(b []byte) (n int, err error) {
	if len(b) < 40 {
		return -1, io.ErrShortBuffer
	}
	v := (*Savepoint)(unsafe.Pointer(&b[0]))
	*v = *s
	return 40, nil
}
func (s *Savepoint) UnmarshalBinary(b []byte) error {
	if len(b) < 40 {
		return io.ErrShortBuffer
	}
	v := (*Savepoint)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Savepoint) Clone() *Savepoint {
	v := &Savepoint{}
	*v = *s
	return v
}
func (s *Savepoint) Bytes() []byte {
	return (*(*[40]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Savepoint) Mut() *SavepointMut {
	return (*SavepointMut)(unsafe.Pointer(s))
}
func (s *Savepoint) RecordID() *RecordID {
	return &s.recordID
}
func (s *Savepoint) Timestamp() int64 {
	return s.timestamp
}
func (s *Savepoint) WriterID() int64 {
	return s.writerID
}

type SavepointMut struct {
	Savepoint
}

func (s *SavepointMut) Clone() *SavepointMut {
	v := &SavepointMut{}
	*v = *s
	return v
}
func (s *SavepointMut) Freeze() *Savepoint {
	return (*Savepoint)(unsafe.Pointer(s))
}
func (s *SavepointMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *SavepointMut) SetRecordID(v *RecordID) *SavepointMut {
	s.recordID = *v
	return s
}
func (s *SavepointMut) SetTimestamp(v int64) *SavepointMut {
	s.timestamp = v
	return s
}
func (s *SavepointMut) SetWriterID(v int64) *SavepointMut {
	s.writerID = v
	return s
}

// Block32
type Block32 struct {
	head BlockHeader
	body Bytes32688
}

func (s *Block32) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block32) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block32) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[32784]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 32784 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block32) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[32784]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block32) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[32784]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block32) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[32784]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block32) Read(b []byte) (n int, err error) {
	if len(b) < 32784 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block32)(unsafe.Pointer(&b[0]))
	*v = *s
	return 32784, nil
}
func (s *Block32) UnmarshalBinary(b []byte) error {
	if len(b) < 32784 {
		return io.ErrShortBuffer
	}
	v := (*Block32)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block32) Clone() *Block32 {
	v := &Block32{}
	*v = *s
	return v
}
func (s *Block32) Bytes() []byte {
	return (*(*[32784]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block32) Mut() *Block32Mut {
	return (*Block32Mut)(unsafe.Pointer(s))
}
func (s *Block32) Head() *BlockHeader {
	return &s.head
}
func (s *Block32) Body() *Bytes32688 {
	return &s.body
}

// Block32
type Block32Mut struct {
	Block32
}

func (s *Block32Mut) Clone() *Block32Mut {
	v := &Block32Mut{}
	*v = *s
	return v
}
func (s *Block32Mut) Freeze() *Block32 {
	return (*Block32)(unsafe.Pointer(s))
}
func (s *Block32Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block32Mut) SetHead(v *BlockHeader) *Block32Mut {
	s.head = *v
	return s
}
func (s *Block32Mut) SetBody(v Bytes32688) *Block32Mut {
	s.body = v
	return s
}

type Stopped struct {
	recordID  RecordID
	timestamp int64
	starts    int64
	reason    StopReason
	_         [7]byte // Padding
}

func (s *Stopped) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Stopped) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["starts"] = s.Starts()
	m["reason"] = s.Reason()
	return m
}

func (s *Stopped) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[48]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 48 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Stopped) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[48]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Stopped) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Stopped) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Stopped) Read(b []byte) (n int, err error) {
	if len(b) < 48 {
		return -1, io.ErrShortBuffer
	}
	v := (*Stopped)(unsafe.Pointer(&b[0]))
	*v = *s
	return 48, nil
}
func (s *Stopped) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return io.ErrShortBuffer
	}
	v := (*Stopped)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Stopped) Clone() *Stopped {
	v := &Stopped{}
	*v = *s
	return v
}
func (s *Stopped) Bytes() []byte {
	return (*(*[48]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Stopped) Mut() *StoppedMut {
	return (*StoppedMut)(unsafe.Pointer(s))
}
func (s *Stopped) RecordID() *RecordID {
	return &s.recordID
}
func (s *Stopped) Timestamp() int64 {
	return s.timestamp
}
func (s *Stopped) Starts() int64 {
	return s.starts
}
func (s *Stopped) Reason() StopReason {
	return s.reason
}

type StoppedMut struct {
	Stopped
}

func (s *StoppedMut) Clone() *StoppedMut {
	v := &StoppedMut{}
	*v = *s
	return v
}
func (s *StoppedMut) Freeze() *Stopped {
	return (*Stopped)(unsafe.Pointer(s))
}
func (s *StoppedMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *StoppedMut) SetRecordID(v *RecordID) *StoppedMut {
	s.recordID = *v
	return s
}
func (s *StoppedMut) SetTimestamp(v int64) *StoppedMut {
	s.timestamp = v
	return s
}
func (s *StoppedMut) SetStarts(v int64) *StoppedMut {
	s.starts = v
	return s
}
func (s *StoppedMut) SetReason(v StopReason) *StoppedMut {
	s.reason = v
	return s
}

// Block16
type Block16 struct {
	head BlockHeader
	body Bytes16306
	_    [6]byte // Padding
}

func (s *Block16) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block16) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block16) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[16408]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 16408 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block16) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[16408]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block16) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[16408]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block16) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[16408]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block16) Read(b []byte) (n int, err error) {
	if len(b) < 16408 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block16)(unsafe.Pointer(&b[0]))
	*v = *s
	return 16408, nil
}
func (s *Block16) UnmarshalBinary(b []byte) error {
	if len(b) < 16408 {
		return io.ErrShortBuffer
	}
	v := (*Block16)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block16) Clone() *Block16 {
	v := &Block16{}
	*v = *s
	return v
}
func (s *Block16) Bytes() []byte {
	return (*(*[16408]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block16) Mut() *Block16Mut {
	return (*Block16Mut)(unsafe.Pointer(s))
}
func (s *Block16) Head() *BlockHeader {
	return &s.head
}
func (s *Block16) Body() *Bytes16306 {
	return &s.body
}

// Block16
type Block16Mut struct {
	Block16
}

func (s *Block16Mut) Clone() *Block16Mut {
	v := &Block16Mut{}
	*v = *s
	return v
}
func (s *Block16Mut) Freeze() *Block16 {
	return (*Block16)(unsafe.Pointer(s))
}
func (s *Block16Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block16Mut) SetHead(v *BlockHeader) *Block16Mut {
	s.head = *v
	return s
}
func (s *Block16Mut) SetBody(v Bytes16306) *Block16Mut {
	s.body = v
	return s
}

type Starting struct {
	recordID  RecordID
	timestamp int64
	writerID  int64
}

func (s *Starting) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Starting) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["writerID"] = s.WriterID()
	return m
}

func (s *Starting) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[40]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 40 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Starting) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[40]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Starting) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Starting) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[40]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Starting) Read(b []byte) (n int, err error) {
	if len(b) < 40 {
		return -1, io.ErrShortBuffer
	}
	v := (*Starting)(unsafe.Pointer(&b[0]))
	*v = *s
	return 40, nil
}
func (s *Starting) UnmarshalBinary(b []byte) error {
	if len(b) < 40 {
		return io.ErrShortBuffer
	}
	v := (*Starting)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Starting) Clone() *Starting {
	v := &Starting{}
	*v = *s
	return v
}
func (s *Starting) Bytes() []byte {
	return (*(*[40]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Starting) Mut() *StartingMut {
	return (*StartingMut)(unsafe.Pointer(s))
}
func (s *Starting) RecordID() *RecordID {
	return &s.recordID
}
func (s *Starting) Timestamp() int64 {
	return s.timestamp
}
func (s *Starting) WriterID() int64 {
	return s.writerID
}

type StartingMut struct {
	Starting
}

func (s *StartingMut) Clone() *StartingMut {
	v := &StartingMut{}
	*v = *s
	return v
}
func (s *StartingMut) Freeze() *Starting {
	return (*Starting)(unsafe.Pointer(s))
}
func (s *StartingMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *StartingMut) SetRecordID(v *RecordID) *StartingMut {
	s.recordID = *v
	return s
}
func (s *StartingMut) SetTimestamp(v int64) *StartingMut {
	s.timestamp = v
	return s
}
func (s *StartingMut) SetWriterID(v int64) *StartingMut {
	s.writerID = v
	return s
}

// Block1
type Block1 struct {
	head BlockHeader
	body Bytes944
}

func (s *Block1) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block1) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block1) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[1040]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 1040 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block1) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[1040]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block1) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[1040]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block1) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[1040]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block1) Read(b []byte) (n int, err error) {
	if len(b) < 1040 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block1)(unsafe.Pointer(&b[0]))
	*v = *s
	return 1040, nil
}
func (s *Block1) UnmarshalBinary(b []byte) error {
	if len(b) < 1040 {
		return io.ErrShortBuffer
	}
	v := (*Block1)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block1) Clone() *Block1 {
	v := &Block1{}
	*v = *s
	return v
}
func (s *Block1) Bytes() []byte {
	return (*(*[1040]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block1) Mut() *Block1Mut {
	return (*Block1Mut)(unsafe.Pointer(s))
}
func (s *Block1) Head() *BlockHeader {
	return &s.head
}
func (s *Block1) Body() *Bytes944 {
	return &s.body
}

// Block1
type Block1Mut struct {
	Block1
}

func (s *Block1Mut) Clone() *Block1Mut {
	v := &Block1Mut{}
	*v = *s
	return v
}
func (s *Block1Mut) Freeze() *Block1 {
	return (*Block1)(unsafe.Pointer(s))
}
func (s *Block1Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block1Mut) SetHead(v *BlockHeader) *Block1Mut {
	s.head = *v
	return s
}
func (s *Block1Mut) SetBody(v Bytes944) *Block1Mut {
	s.body = v
	return s
}

// Block8
type Block8 struct {
	head BlockHeader
	body Bytes8112
}

func (s *Block8) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block8) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block8) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[8208]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 8208 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block8) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[8208]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block8) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[8208]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block8) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[8208]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block8) Read(b []byte) (n int, err error) {
	if len(b) < 8208 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block8)(unsafe.Pointer(&b[0]))
	*v = *s
	return 8208, nil
}
func (s *Block8) UnmarshalBinary(b []byte) error {
	if len(b) < 8208 {
		return io.ErrShortBuffer
	}
	v := (*Block8)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block8) Clone() *Block8 {
	v := &Block8{}
	*v = *s
	return v
}
func (s *Block8) Bytes() []byte {
	return (*(*[8208]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block8) Mut() *Block8Mut {
	return (*Block8Mut)(unsafe.Pointer(s))
}
func (s *Block8) Head() *BlockHeader {
	return &s.head
}
func (s *Block8) Body() *Bytes8112 {
	return &s.body
}

// Block8
type Block8Mut struct {
	Block8
}

func (s *Block8Mut) Clone() *Block8Mut {
	v := &Block8Mut{}
	*v = *s
	return v
}
func (s *Block8Mut) Freeze() *Block8 {
	return (*Block8)(unsafe.Pointer(s))
}
func (s *Block8Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block8Mut) SetHead(v *BlockHeader) *Block8Mut {
	s.head = *v
	return s
}
func (s *Block8Mut) SetBody(v Bytes8112) *Block8Mut {
	s.body = v
	return s
}

// BlockHeader
type BlockHeader struct {
	streamID    int64
	id          int64
	created     int64
	completed   int64
	min         int64
	max         int64
	start       int64
	end         int64
	savepoint   int64
	savepointR  int64
	count       uint16
	seq         uint16
	size        uint16
	sizeU       uint16
	sizeX       uint16
	compression Compression
	_           [5]byte // Padding
}

func (s *BlockHeader) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *BlockHeader) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["streamID"] = s.StreamID()
	m["id"] = s.Id()
	m["created"] = s.Created()
	m["completed"] = s.Completed()
	m["min"] = s.Min()
	m["max"] = s.Max()
	m["start"] = s.Start()
	m["end"] = s.End()
	m["savepoint"] = s.Savepoint()
	m["savepointR"] = s.SavepointR()
	m["count"] = s.Count()
	m["seq"] = s.Seq()
	m["size"] = s.Size()
	m["sizeU"] = s.SizeU()
	m["sizeX"] = s.SizeX()
	m["compression"] = s.Compression()
	return m
}

func (s *BlockHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[96]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 96 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *BlockHeader) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[96]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *BlockHeader) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[96]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *BlockHeader) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[96]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *BlockHeader) Read(b []byte) (n int, err error) {
	if len(b) < 96 {
		return -1, io.ErrShortBuffer
	}
	v := (*BlockHeader)(unsafe.Pointer(&b[0]))
	*v = *s
	return 96, nil
}
func (s *BlockHeader) UnmarshalBinary(b []byte) error {
	if len(b) < 96 {
		return io.ErrShortBuffer
	}
	v := (*BlockHeader)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *BlockHeader) Clone() *BlockHeader {
	v := &BlockHeader{}
	*v = *s
	return v
}
func (s *BlockHeader) Bytes() []byte {
	return (*(*[96]byte)(unsafe.Pointer(s)))[0:]
}
func (s *BlockHeader) Mut() *BlockHeaderMut {
	return (*BlockHeaderMut)(unsafe.Pointer(s))
}
func (s *BlockHeader) StreamID() int64 {
	return s.streamID
}
func (s *BlockHeader) Id() int64 {
	return s.id
}
func (s *BlockHeader) Created() int64 {
	return s.created
}
func (s *BlockHeader) Completed() int64 {
	return s.completed
}
func (s *BlockHeader) Min() int64 {
	return s.min
}
func (s *BlockHeader) Max() int64 {
	return s.max
}
func (s *BlockHeader) Start() int64 {
	return s.start
}
func (s *BlockHeader) End() int64 {
	return s.end
}
func (s *BlockHeader) Savepoint() int64 {
	return s.savepoint
}
func (s *BlockHeader) SavepointR() int64 {
	return s.savepointR
}
func (s *BlockHeader) Count() uint16 {
	return s.count
}
func (s *BlockHeader) Seq() uint16 {
	return s.seq
}
func (s *BlockHeader) Size() uint16 {
	return s.size
}
func (s *BlockHeader) SizeU() uint16 {
	return s.sizeU
}
func (s *BlockHeader) SizeX() uint16 {
	return s.sizeX
}
func (s *BlockHeader) Compression() Compression {
	return s.compression
}

// BlockHeader
type BlockHeaderMut struct {
	BlockHeader
}

func (s *BlockHeaderMut) Clone() *BlockHeaderMut {
	v := &BlockHeaderMut{}
	*v = *s
	return v
}
func (s *BlockHeaderMut) Freeze() *BlockHeader {
	return (*BlockHeader)(unsafe.Pointer(s))
}
func (s *BlockHeaderMut) SetStreamID(v int64) *BlockHeaderMut {
	s.streamID = v
	return s
}
func (s *BlockHeaderMut) SetId(v int64) *BlockHeaderMut {
	s.id = v
	return s
}
func (s *BlockHeaderMut) SetCreated(v int64) *BlockHeaderMut {
	s.created = v
	return s
}
func (s *BlockHeaderMut) SetCompleted(v int64) *BlockHeaderMut {
	s.completed = v
	return s
}
func (s *BlockHeaderMut) SetMin(v int64) *BlockHeaderMut {
	s.min = v
	return s
}
func (s *BlockHeaderMut) SetMax(v int64) *BlockHeaderMut {
	s.max = v
	return s
}
func (s *BlockHeaderMut) SetStart(v int64) *BlockHeaderMut {
	s.start = v
	return s
}
func (s *BlockHeaderMut) SetEnd(v int64) *BlockHeaderMut {
	s.end = v
	return s
}
func (s *BlockHeaderMut) SetSavepoint(v int64) *BlockHeaderMut {
	s.savepoint = v
	return s
}
func (s *BlockHeaderMut) SetSavepointR(v int64) *BlockHeaderMut {
	s.savepointR = v
	return s
}
func (s *BlockHeaderMut) SetCount(v uint16) *BlockHeaderMut {
	s.count = v
	return s
}
func (s *BlockHeaderMut) SetSeq(v uint16) *BlockHeaderMut {
	s.seq = v
	return s
}
func (s *BlockHeaderMut) SetSize(v uint16) *BlockHeaderMut {
	s.size = v
	return s
}
func (s *BlockHeaderMut) SetSizeU(v uint16) *BlockHeaderMut {
	s.sizeU = v
	return s
}
func (s *BlockHeaderMut) SetSizeX(v uint16) *BlockHeaderMut {
	s.sizeX = v
	return s
}
func (s *BlockHeaderMut) SetCompression(v Compression) *BlockHeaderMut {
	s.compression = v
	return s
}

type Stats struct {
	size   int64
	count  int64
	blocks int64
}

func (s *Stats) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Stats) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["size"] = s.Size()
	m["count"] = s.Count()
	m["blocks"] = s.Blocks()
	return m
}

func (s *Stats) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Stats) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Stats) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Stats) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Stats) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*Stats)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *Stats) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*Stats)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Stats) Clone() *Stats {
	v := &Stats{}
	*v = *s
	return v
}
func (s *Stats) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Stats) Mut() *StatsMut {
	return (*StatsMut)(unsafe.Pointer(s))
}
func (s *Stats) Size() int64 {
	return s.size
}
func (s *Stats) Count() int64 {
	return s.count
}
func (s *Stats) Blocks() int64 {
	return s.blocks
}

type StatsMut struct {
	Stats
}

func (s *StatsMut) Clone() *StatsMut {
	v := &StatsMut{}
	*v = *s
	return v
}
func (s *StatsMut) Freeze() *Stats {
	return (*Stats)(unsafe.Pointer(s))
}
func (s *StatsMut) SetSize(v int64) *StatsMut {
	s.size = v
	return s
}
func (s *StatsMut) SetCount(v int64) *StatsMut {
	s.count = v
	return s
}
func (s *StatsMut) SetBlocks(v int64) *StatsMut {
	s.blocks = v
	return s
}

type AccountStats struct {
	id       int64
	storage  Stats
	appender Stats
	streams  int64
}

func (s *AccountStats) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *AccountStats) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["id"] = s.Id()
	m["storage"] = s.Storage().MarshalMap(nil)
	m["appender"] = s.Appender().MarshalMap(nil)
	m["streams"] = s.Streams()
	return m
}

func (s *AccountStats) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[64]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 64 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *AccountStats) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[64]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *AccountStats) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[64]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *AccountStats) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[64]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *AccountStats) Read(b []byte) (n int, err error) {
	if len(b) < 64 {
		return -1, io.ErrShortBuffer
	}
	v := (*AccountStats)(unsafe.Pointer(&b[0]))
	*v = *s
	return 64, nil
}
func (s *AccountStats) UnmarshalBinary(b []byte) error {
	if len(b) < 64 {
		return io.ErrShortBuffer
	}
	v := (*AccountStats)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *AccountStats) Clone() *AccountStats {
	v := &AccountStats{}
	*v = *s
	return v
}
func (s *AccountStats) Bytes() []byte {
	return (*(*[64]byte)(unsafe.Pointer(s)))[0:]
}
func (s *AccountStats) Mut() *AccountStatsMut {
	return (*AccountStatsMut)(unsafe.Pointer(s))
}
func (s *AccountStats) Id() int64 {
	return s.id
}
func (s *AccountStats) Storage() *Stats {
	return &s.storage
}
func (s *AccountStats) Appender() *Stats {
	return &s.appender
}
func (s *AccountStats) Streams() int64 {
	return s.streams
}

type AccountStatsMut struct {
	AccountStats
}

func (s *AccountStatsMut) Clone() *AccountStatsMut {
	v := &AccountStatsMut{}
	*v = *s
	return v
}
func (s *AccountStatsMut) Freeze() *AccountStats {
	return (*AccountStats)(unsafe.Pointer(s))
}
func (s *AccountStatsMut) SetId(v int64) *AccountStatsMut {
	s.id = v
	return s
}
func (s *AccountStatsMut) Storage() *StatsMut {
	return s.storage.Mut()
}
func (s *AccountStatsMut) SetStorage(v *Stats) *AccountStatsMut {
	s.storage = *v
	return s
}
func (s *AccountStatsMut) Appender() *StatsMut {
	return s.appender.Mut()
}
func (s *AccountStatsMut) SetAppender(v *Stats) *AccountStatsMut {
	s.appender = *v
	return s
}
func (s *AccountStatsMut) SetStreams(v int64) *AccountStatsMut {
	s.streams = v
	return s
}

type RecordID struct {
	streamID int64
	blockID  int64
	id       int64
}

func (s *RecordID) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *RecordID) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["streamID"] = s.StreamID()
	m["blockID"] = s.BlockID()
	m["id"] = s.Id()
	return m
}

func (s *RecordID) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *RecordID) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *RecordID) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *RecordID) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *RecordID) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*RecordID)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *RecordID) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*RecordID)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *RecordID) Clone() *RecordID {
	v := &RecordID{}
	*v = *s
	return v
}
func (s *RecordID) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *RecordID) Mut() *RecordIDMut {
	return (*RecordIDMut)(unsafe.Pointer(s))
}
func (s *RecordID) StreamID() int64 {
	return s.streamID
}
func (s *RecordID) BlockID() int64 {
	return s.blockID
}
func (s *RecordID) Id() int64 {
	return s.id
}

type RecordIDMut struct {
	RecordID
}

func (s *RecordIDMut) Clone() *RecordIDMut {
	v := &RecordIDMut{}
	*v = *s
	return v
}
func (s *RecordIDMut) Freeze() *RecordID {
	return (*RecordID)(unsafe.Pointer(s))
}
func (s *RecordIDMut) SetStreamID(v int64) *RecordIDMut {
	s.streamID = v
	return s
}
func (s *RecordIDMut) SetBlockID(v int64) *RecordIDMut {
	s.blockID = v
	return s
}
func (s *RecordIDMut) SetId(v int64) *RecordIDMut {
	s.id = v
	return s
}

// Block2
type Block2 struct {
	head BlockHeader
	body Bytes1968
}

func (s *Block2) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block2) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block2) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[2064]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 2064 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block2) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[2064]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block2) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[2064]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block2) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[2064]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block2) Read(b []byte) (n int, err error) {
	if len(b) < 2064 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block2)(unsafe.Pointer(&b[0]))
	*v = *s
	return 2064, nil
}
func (s *Block2) UnmarshalBinary(b []byte) error {
	if len(b) < 2064 {
		return io.ErrShortBuffer
	}
	v := (*Block2)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block2) Clone() *Block2 {
	v := &Block2{}
	*v = *s
	return v
}
func (s *Block2) Bytes() []byte {
	return (*(*[2064]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block2) Mut() *Block2Mut {
	return (*Block2Mut)(unsafe.Pointer(s))
}
func (s *Block2) Head() *BlockHeader {
	return &s.head
}
func (s *Block2) Body() *Bytes1968 {
	return &s.body
}

// Block2
type Block2Mut struct {
	Block2
}

func (s *Block2Mut) Clone() *Block2Mut {
	v := &Block2Mut{}
	*v = *s
	return v
}
func (s *Block2Mut) Freeze() *Block2 {
	return (*Block2)(unsafe.Pointer(s))
}
func (s *Block2Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block2Mut) SetHead(v *BlockHeader) *Block2Mut {
	s.head = *v
	return s
}
func (s *Block2Mut) SetBody(v Bytes1968) *Block2Mut {
	s.body = v
	return s
}

type Progress struct {
	recordID  RecordID
	timestamp int64
	writerID  int64
	started   int64
	count     int64
	remaining int64
}

func (s *Progress) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Progress) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["writerID"] = s.WriterID()
	m["started"] = s.Started()
	m["count"] = s.Count()
	m["remaining"] = s.Remaining()
	return m
}

func (s *Progress) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[64]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 64 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Progress) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[64]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Progress) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[64]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Progress) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[64]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Progress) Read(b []byte) (n int, err error) {
	if len(b) < 64 {
		return -1, io.ErrShortBuffer
	}
	v := (*Progress)(unsafe.Pointer(&b[0]))
	*v = *s
	return 64, nil
}
func (s *Progress) UnmarshalBinary(b []byte) error {
	if len(b) < 64 {
		return io.ErrShortBuffer
	}
	v := (*Progress)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Progress) Clone() *Progress {
	v := &Progress{}
	*v = *s
	return v
}
func (s *Progress) Bytes() []byte {
	return (*(*[64]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Progress) Mut() *ProgressMut {
	return (*ProgressMut)(unsafe.Pointer(s))
}
func (s *Progress) RecordID() *RecordID {
	return &s.recordID
}
func (s *Progress) Timestamp() int64 {
	return s.timestamp
}
func (s *Progress) WriterID() int64 {
	return s.writerID
}
func (s *Progress) Started() int64 {
	return s.started
}
func (s *Progress) Count() int64 {
	return s.count
}
func (s *Progress) Remaining() int64 {
	return s.remaining
}

type ProgressMut struct {
	Progress
}

func (s *ProgressMut) Clone() *ProgressMut {
	v := &ProgressMut{}
	*v = *s
	return v
}
func (s *ProgressMut) Freeze() *Progress {
	return (*Progress)(unsafe.Pointer(s))
}
func (s *ProgressMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *ProgressMut) SetRecordID(v *RecordID) *ProgressMut {
	s.recordID = *v
	return s
}
func (s *ProgressMut) SetTimestamp(v int64) *ProgressMut {
	s.timestamp = v
	return s
}
func (s *ProgressMut) SetWriterID(v int64) *ProgressMut {
	s.writerID = v
	return s
}
func (s *ProgressMut) SetStarted(v int64) *ProgressMut {
	s.started = v
	return s
}
func (s *ProgressMut) SetCount(v int64) *ProgressMut {
	s.count = v
	return s
}
func (s *ProgressMut) SetRemaining(v int64) *ProgressMut {
	s.remaining = v
	return s
}

// Block4
type Block4 struct {
	head BlockHeader
	body Bytes4016
}

func (s *Block4) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Block4) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["head"] = s.Head().MarshalMap(nil)
	m["body"] = s.Body()
	return m
}

func (s *Block4) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[4112]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 4112 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Block4) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[4112]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Block4) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[4112]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Block4) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[4112]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Block4) Read(b []byte) (n int, err error) {
	if len(b) < 4112 {
		return -1, io.ErrShortBuffer
	}
	v := (*Block4)(unsafe.Pointer(&b[0]))
	*v = *s
	return 4112, nil
}
func (s *Block4) UnmarshalBinary(b []byte) error {
	if len(b) < 4112 {
		return io.ErrShortBuffer
	}
	v := (*Block4)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Block4) Clone() *Block4 {
	v := &Block4{}
	*v = *s
	return v
}
func (s *Block4) Bytes() []byte {
	return (*(*[4112]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Block4) Mut() *Block4Mut {
	return (*Block4Mut)(unsafe.Pointer(s))
}
func (s *Block4) Head() *BlockHeader {
	return &s.head
}
func (s *Block4) Body() *Bytes4016 {
	return &s.body
}

// Block4
type Block4Mut struct {
	Block4
}

func (s *Block4Mut) Clone() *Block4Mut {
	v := &Block4Mut{}
	*v = *s
	return v
}
func (s *Block4Mut) Freeze() *Block4 {
	return (*Block4)(unsafe.Pointer(s))
}
func (s *Block4Mut) Head() *BlockHeaderMut {
	return s.head.Mut()
}
func (s *Block4Mut) SetHead(v *BlockHeader) *Block4Mut {
	s.head = *v
	return s
}
func (s *Block4Mut) SetBody(v Bytes4016) *Block4Mut {
	s.body = v
	return s
}

type StreamStats struct {
	storage  Stats
	appender Stats
}

func (s *StreamStats) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *StreamStats) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["storage"] = s.Storage().MarshalMap(nil)
	m["appender"] = s.Appender().MarshalMap(nil)
	return m
}

func (s *StreamStats) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[48]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 48 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *StreamStats) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[48]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *StreamStats) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *StreamStats) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *StreamStats) Read(b []byte) (n int, err error) {
	if len(b) < 48 {
		return -1, io.ErrShortBuffer
	}
	v := (*StreamStats)(unsafe.Pointer(&b[0]))
	*v = *s
	return 48, nil
}
func (s *StreamStats) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return io.ErrShortBuffer
	}
	v := (*StreamStats)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *StreamStats) Clone() *StreamStats {
	v := &StreamStats{}
	*v = *s
	return v
}
func (s *StreamStats) Bytes() []byte {
	return (*(*[48]byte)(unsafe.Pointer(s)))[0:]
}
func (s *StreamStats) Mut() *StreamStatsMut {
	return (*StreamStatsMut)(unsafe.Pointer(s))
}
func (s *StreamStats) Storage() *Stats {
	return &s.storage
}
func (s *StreamStats) Appender() *Stats {
	return &s.appender
}

type StreamStatsMut struct {
	StreamStats
}

func (s *StreamStatsMut) Clone() *StreamStatsMut {
	v := &StreamStatsMut{}
	*v = *s
	return v
}
func (s *StreamStatsMut) Freeze() *StreamStats {
	return (*StreamStats)(unsafe.Pointer(s))
}
func (s *StreamStatsMut) Storage() *StatsMut {
	return s.storage.Mut()
}
func (s *StreamStatsMut) SetStorage(v *Stats) *StreamStatsMut {
	s.storage = *v
	return s
}
func (s *StreamStatsMut) Appender() *StatsMut {
	return s.appender.Mut()
}
func (s *StreamStatsMut) SetAppender(v *Stats) *StreamStatsMut {
	s.appender = *v
	return s
}

type Started struct {
	recordID  RecordID
	timestamp int64
	writerID  int64
	stops     int64
}

func (s *Started) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Started) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["writerID"] = s.WriterID()
	m["stops"] = s.Stops()
	return m
}

func (s *Started) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[48]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 48 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Started) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[48]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Started) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Started) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Started) Read(b []byte) (n int, err error) {
	if len(b) < 48 {
		return -1, io.ErrShortBuffer
	}
	v := (*Started)(unsafe.Pointer(&b[0]))
	*v = *s
	return 48, nil
}
func (s *Started) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return io.ErrShortBuffer
	}
	v := (*Started)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Started) Clone() *Started {
	v := &Started{}
	*v = *s
	return v
}
func (s *Started) Bytes() []byte {
	return (*(*[48]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Started) Mut() *StartedMut {
	return (*StartedMut)(unsafe.Pointer(s))
}
func (s *Started) RecordID() *RecordID {
	return &s.recordID
}
func (s *Started) Timestamp() int64 {
	return s.timestamp
}
func (s *Started) WriterID() int64 {
	return s.writerID
}
func (s *Started) Stops() int64 {
	return s.stops
}

type StartedMut struct {
	Started
}

func (s *StartedMut) Clone() *StartedMut {
	v := &StartedMut{}
	*v = *s
	return v
}
func (s *StartedMut) Freeze() *Started {
	return (*Started)(unsafe.Pointer(s))
}
func (s *StartedMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *StartedMut) SetRecordID(v *RecordID) *StartedMut {
	s.recordID = *v
	return s
}
func (s *StartedMut) SetTimestamp(v int64) *StartedMut {
	s.timestamp = v
	return s
}
func (s *StartedMut) SetWriterID(v int64) *StartedMut {
	s.writerID = v
	return s
}
func (s *StartedMut) SetStops(v int64) *StartedMut {
	s.stops = v
	return s
}

// End of Stream
// The reader is caught up on the stream.
type EOS struct {
	recordID  RecordID
	timestamp int64
	writerID  int64
	closed    bool
	waiting   bool
	_         [6]byte // Padding
}

func (s *EOS) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *EOS) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["recordID"] = s.RecordID().MarshalMap(nil)
	m["timestamp"] = s.Timestamp()
	m["writerID"] = s.WriterID()
	m["closed"] = s.Closed()
	m["waiting"] = s.Waiting()
	return m
}

func (s *EOS) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[48]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 48 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *EOS) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[48]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *EOS) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *EOS) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *EOS) Read(b []byte) (n int, err error) {
	if len(b) < 48 {
		return -1, io.ErrShortBuffer
	}
	v := (*EOS)(unsafe.Pointer(&b[0]))
	*v = *s
	return 48, nil
}
func (s *EOS) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return io.ErrShortBuffer
	}
	v := (*EOS)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *EOS) Clone() *EOS {
	v := &EOS{}
	*v = *s
	return v
}
func (s *EOS) Bytes() []byte {
	return (*(*[48]byte)(unsafe.Pointer(s)))[0:]
}
func (s *EOS) Mut() *EOSMut {
	return (*EOSMut)(unsafe.Pointer(s))
}
func (s *EOS) RecordID() *RecordID {
	return &s.recordID
}
func (s *EOS) Timestamp() int64 {
	return s.timestamp
}
func (s *EOS) WriterID() int64 {
	return s.writerID
}
func (s *EOS) Closed() bool {
	return s.closed
}
func (s *EOS) Waiting() bool {
	return s.waiting
}

// End of Stream
// The reader is caught up on the stream.
type EOSMut struct {
	EOS
}

func (s *EOSMut) Clone() *EOSMut {
	v := &EOSMut{}
	*v = *s
	return v
}
func (s *EOSMut) Freeze() *EOS {
	return (*EOS)(unsafe.Pointer(s))
}
func (s *EOSMut) RecordID() *RecordIDMut {
	return s.recordID.Mut()
}
func (s *EOSMut) SetRecordID(v *RecordID) *EOSMut {
	s.recordID = *v
	return s
}
func (s *EOSMut) SetTimestamp(v int64) *EOSMut {
	s.timestamp = v
	return s
}
func (s *EOSMut) SetWriterID(v int64) *EOSMut {
	s.writerID = v
	return s
}
func (s *EOSMut) SetClosed(v bool) *EOSMut {
	s.closed = v
	return s
}
func (s *EOSMut) SetWaiting(v bool) *EOSMut {
	s.waiting = v
	return s
}

type Bytes944 [944]byte

func NewBytes944(s string) *Bytes944 {
	v := Bytes944{}
	v.set(s)
	return &v
}
func (s *Bytes944) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes944) Len() int {
	return 944
}
func (s *Bytes944) Cap() int {
	return 944
}
func (s *Bytes944) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes944) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes944) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes944) Clone() *Bytes944 {
	v := Bytes944{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes944) Mut() *Bytes944Mut {
	return *(**Bytes944Mut)(unsafe.Pointer(&s))
}
func (s *Bytes944) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[944]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 944 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes944) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[944]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes944) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[944]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes944) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[944]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes944) UnmarshalBinary(b []byte) error {
	if len(b) < 944 {
		return io.ErrShortBuffer
	}
	v := (*Bytes944)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes944Mut struct {
	Bytes944
}

func (s *Bytes944Mut) Set(v string) {
	s.set(v)
}

type Bytes65456 [65456]byte

func NewBytes65456(s string) *Bytes65456 {
	v := Bytes65456{}
	v.set(s)
	return &v
}
func (s *Bytes65456) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes65456) Len() int {
	return 65456
}
func (s *Bytes65456) Cap() int {
	return 65456
}
func (s *Bytes65456) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes65456) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes65456) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes65456) Clone() *Bytes65456 {
	v := Bytes65456{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes65456) Mut() *Bytes65456Mut {
	return *(**Bytes65456Mut)(unsafe.Pointer(&s))
}
func (s *Bytes65456) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[65456]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 65456 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes65456) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[65456]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes65456) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[65456]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes65456) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[65456]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes65456) UnmarshalBinary(b []byte) error {
	if len(b) < 65456 {
		return io.ErrShortBuffer
	}
	v := (*Bytes65456)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes65456Mut struct {
	Bytes65456
}

func (s *Bytes65456Mut) Set(v string) {
	s.set(v)
}

type Bytes32688 [32688]byte

func NewBytes32688(s string) *Bytes32688 {
	v := Bytes32688{}
	v.set(s)
	return &v
}
func (s *Bytes32688) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes32688) Len() int {
	return 32688
}
func (s *Bytes32688) Cap() int {
	return 32688
}
func (s *Bytes32688) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes32688) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes32688) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes32688) Clone() *Bytes32688 {
	v := Bytes32688{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes32688) Mut() *Bytes32688Mut {
	return *(**Bytes32688Mut)(unsafe.Pointer(&s))
}
func (s *Bytes32688) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[32688]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 32688 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes32688) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[32688]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes32688) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[32688]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes32688) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[32688]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes32688) UnmarshalBinary(b []byte) error {
	if len(b) < 32688 {
		return io.ErrShortBuffer
	}
	v := (*Bytes32688)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes32688Mut struct {
	Bytes32688
}

func (s *Bytes32688Mut) Set(v string) {
	s.set(v)
}

type Bytes1968 [1968]byte

func NewBytes1968(s string) *Bytes1968 {
	v := Bytes1968{}
	v.set(s)
	return &v
}
func (s *Bytes1968) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes1968) Len() int {
	return 1968
}
func (s *Bytes1968) Cap() int {
	return 1968
}
func (s *Bytes1968) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes1968) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes1968) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes1968) Clone() *Bytes1968 {
	v := Bytes1968{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes1968) Mut() *Bytes1968Mut {
	return *(**Bytes1968Mut)(unsafe.Pointer(&s))
}
func (s *Bytes1968) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[1968]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 1968 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes1968) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[1968]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes1968) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[1968]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes1968) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[1968]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes1968) UnmarshalBinary(b []byte) error {
	if len(b) < 1968 {
		return io.ErrShortBuffer
	}
	v := (*Bytes1968)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes1968Mut struct {
	Bytes1968
}

func (s *Bytes1968Mut) Set(v string) {
	s.set(v)
}

type Bytes16306 [16306]byte

func NewBytes16306(s string) *Bytes16306 {
	v := Bytes16306{}
	v.set(s)
	return &v
}
func (s *Bytes16306) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes16306) Len() int {
	return 16306
}
func (s *Bytes16306) Cap() int {
	return 16306
}
func (s *Bytes16306) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes16306) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes16306) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes16306) Clone() *Bytes16306 {
	v := Bytes16306{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes16306) Mut() *Bytes16306Mut {
	return *(**Bytes16306Mut)(unsafe.Pointer(&s))
}
func (s *Bytes16306) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[16306]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 16306 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes16306) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[16306]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes16306) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[16306]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes16306) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[16306]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes16306) UnmarshalBinary(b []byte) error {
	if len(b) < 16306 {
		return io.ErrShortBuffer
	}
	v := (*Bytes16306)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes16306Mut struct {
	Bytes16306
}

func (s *Bytes16306Mut) Set(v string) {
	s.set(v)
}

type Bytes4016 [4016]byte

func NewBytes4016(s string) *Bytes4016 {
	v := Bytes4016{}
	v.set(s)
	return &v
}
func (s *Bytes4016) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes4016) Len() int {
	return 4016
}
func (s *Bytes4016) Cap() int {
	return 4016
}
func (s *Bytes4016) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes4016) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes4016) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes4016) Clone() *Bytes4016 {
	v := Bytes4016{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes4016) Mut() *Bytes4016Mut {
	return *(**Bytes4016Mut)(unsafe.Pointer(&s))
}
func (s *Bytes4016) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[4016]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 4016 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes4016) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[4016]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes4016) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[4016]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes4016) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[4016]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes4016) UnmarshalBinary(b []byte) error {
	if len(b) < 4016 {
		return io.ErrShortBuffer
	}
	v := (*Bytes4016)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes4016Mut struct {
	Bytes4016
}

func (s *Bytes4016Mut) Set(v string) {
	s.set(v)
}

type Bytes8112 [8112]byte

func NewBytes8112(s string) *Bytes8112 {
	v := Bytes8112{}
	v.set(s)
	return &v
}
func (s *Bytes8112) set(v string) {
	copy(s[0:], v)
}
func (s *Bytes8112) Len() int {
	return 8112
}
func (s *Bytes8112) Cap() int {
	return 8112
}
func (s *Bytes8112) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *Bytes8112) String() string {
	return string(s[0:s.Len()])
}
func (s *Bytes8112) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *Bytes8112) Clone() *Bytes8112 {
	v := Bytes8112{}
	copy(s[0:], v[0:])
	return &v
}
func (s *Bytes8112) Mut() *Bytes8112Mut {
	return *(**Bytes8112Mut)(unsafe.Pointer(&s))
}
func (s *Bytes8112) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[8112]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 8112 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *Bytes8112) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[8112]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *Bytes8112) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[8112]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *Bytes8112) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[8112]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *Bytes8112) UnmarshalBinary(b []byte) error {
	if len(b) < 8112 {
		return io.ErrShortBuffer
	}
	v := (*Bytes8112)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type Bytes8112Mut struct {
	Bytes8112
}

func (s *Bytes8112Mut) Set(v string) {
	s.set(v)
}

type String32 [32]byte

func NewString32(s string) *String32 {
	v := String32{}
	v.set(s)
	return &v
}
func (s *String32) set(v string) {
	copy(s[0:31], v)
	c := 31
	l := len(v)
	if l > c {
		s[31] = byte(c)
	} else {
		s[31] = byte(l)
	}
}
func (s *String32) Len() int {
	return int(s[31])
}
func (s *String32) Cap() int {
	return 31
}
func (s *String32) Unsafe() string {
	return *(*string)(unsafe.Pointer(s))
}
func (s *String32) String() string {
	return string(s[0:s[31]])
}
func (s *String32) Bytes() []byte {
	return s[0:s.Len()]
}
func (s *String32) Clone() *String32 {
	v := String32{}
	copy(s[0:], v[0:])
	return &v
}
func (s *String32) Mut() *String32Mut {
	return *(**String32Mut)(unsafe.Pointer(&s))
}
func (s *String32) ReadFrom(r io.Reader) error {
	n, err := io.ReadFull(r, (*(*[32]byte)(unsafe.Pointer(&s)))[0:])
	if err != nil {
		return err
	}
	if n != 32 {
		return io.ErrShortBuffer
	}
	return nil
}
func (s *String32) WriteTo(w io.Writer) (n int, err error) {
	return w.Write((*(*[32]byte)(unsafe.Pointer(&s)))[0:])
}
func (s *String32) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[32]byte)(unsafe.Pointer(&s)))[0:]...)
}
func (s *String32) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[32]byte)(unsafe.Pointer(&s)))[0:]...), nil
}
func (s *String32) UnmarshalBinary(b []byte) error {
	if len(b) < 32 {
		return io.ErrShortBuffer
	}
	v := (*String32)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}

type String32Mut struct {
	String32
}

func (s *String32Mut) Set(v string) {
	s.set(v)
}
func init() {
	{
		var b [2]byte
		v := uint16(1)
		b[0] = byte(v)
		b[1] = byte(v >> 8)
		if *(*uint16)(unsafe.Pointer(&b[0])) != 1 {
			panic("BigEndian detected... compiled for LittleEndian only!!!")
		}
	}
	to := reflect.TypeOf
	type sf struct {
		n string
		o uintptr
		s uintptr
	}
	ss := func(tt interface{}, mtt interface{}, s uintptr, fl []sf) {
		t := to(tt)
		mt := to(mtt)
		if t.Size() != s {
			panic(fmt.Sprintf("sizeof %s = %d, expected = %d", t.Name(), t.Size(), s))
		}
		if mt.Size() != s {
			panic(fmt.Sprintf("sizeof %s = %d, expected = %d", mt.Name(), mt.Size(), s))
		}
		if t.NumField() != len(fl) {
			panic(fmt.Sprintf("%s field count = %d: expected %d", t.Name(), t.NumField(), len(fl)))
		}
		for i, ef := range fl {
			f := t.Field(i)
			if f.Offset != ef.o {
				panic(fmt.Sprintf("%s.%s offset = %d, expected = %d", t.Name(), f.Name, f.Offset, ef.o))
			}
			if f.Type.Size() != ef.s {
				panic(fmt.Sprintf("%s.%s size = %d, expected = %d", t.Name(), f.Name, f.Type.Size(), ef.s))
			}
			if f.Name != ef.n {
				panic(fmt.Sprintf("%s.%s expected field: %s", t.Name(), f.Name, ef.n))
			}
		}
	}

	ss(EOB{}, EOBMut{}, 40, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"savepoint", 32, 8},
	})
	ss(RecordHeader{}, RecordHeaderMut{}, 80, []sf{
		{"streamID", 0, 8},
		{"blockID", 8, 8},
		{"id", 16, 8},
		{"timestamp", 24, 8},
		{"start", 32, 8},
		{"end", 40, 8},
		{"savepoint", 48, 8},
		{"savepointR", 56, 8},
		{"seq", 64, 2},
		{"size", 66, 2},
		{"sizeU", 68, 2},
		{"sizeX", 70, 2},
		{"compression", 72, 1},
		{"eob", 73, 1},
		{"_", 74, 6},
	})
	ss(BlockID{}, BlockIDMut{}, 16, []sf{
		{"streamID", 0, 8},
		{"id", 8, 8},
	})
	ss(Stream{}, StreamMut{}, 72, []sf{
		{"id", 0, 8},
		{"created", 8, 8},
		{"accountID", 16, 8},
		{"duration", 24, 8},
		{"name", 32, 32},
		{"record", 64, 4},
		{"kind", 68, 1},
		{"schema", 69, 1},
		{"realTime", 70, 1},
		{"blockSize", 71, 1},
	})
	ss(Block64{}, Block64Mut{}, 65552, []sf{
		{"head", 0, 96},
		{"body", 96, 65456},
	})
	ss(Savepoint{}, SavepointMut{}, 40, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"writerID", 32, 8},
	})
	ss(Block32{}, Block32Mut{}, 32784, []sf{
		{"head", 0, 96},
		{"body", 96, 32688},
	})
	ss(Stopped{}, StoppedMut{}, 48, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"starts", 32, 8},
		{"reason", 40, 1},
		{"_", 41, 7},
	})
	ss(Block16{}, Block16Mut{}, 16408, []sf{
		{"head", 0, 96},
		{"body", 96, 16306},
		{"_", 16402, 6},
	})
	ss(Starting{}, StartingMut{}, 40, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"writerID", 32, 8},
	})
	ss(Block1{}, Block1Mut{}, 1040, []sf{
		{"head", 0, 96},
		{"body", 96, 944},
	})
	ss(Block8{}, Block8Mut{}, 8208, []sf{
		{"head", 0, 96},
		{"body", 96, 8112},
	})
	ss(BlockHeader{}, BlockHeaderMut{}, 96, []sf{
		{"streamID", 0, 8},
		{"id", 8, 8},
		{"created", 16, 8},
		{"completed", 24, 8},
		{"min", 32, 8},
		{"max", 40, 8},
		{"start", 48, 8},
		{"end", 56, 8},
		{"savepoint", 64, 8},
		{"savepointR", 72, 8},
		{"count", 80, 2},
		{"seq", 82, 2},
		{"size", 84, 2},
		{"sizeU", 86, 2},
		{"sizeX", 88, 2},
		{"compression", 90, 1},
		{"_", 91, 5},
	})
	ss(Stats{}, StatsMut{}, 24, []sf{
		{"size", 0, 8},
		{"count", 8, 8},
		{"blocks", 16, 8},
	})
	ss(AccountStats{}, AccountStatsMut{}, 64, []sf{
		{"id", 0, 8},
		{"storage", 8, 24},
		{"appender", 32, 24},
		{"streams", 56, 8},
	})
	ss(RecordID{}, RecordIDMut{}, 24, []sf{
		{"streamID", 0, 8},
		{"blockID", 8, 8},
		{"id", 16, 8},
	})
	ss(Block2{}, Block2Mut{}, 2064, []sf{
		{"head", 0, 96},
		{"body", 96, 1968},
	})
	ss(Progress{}, ProgressMut{}, 64, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"writerID", 32, 8},
		{"started", 40, 8},
		{"count", 48, 8},
		{"remaining", 56, 8},
	})
	ss(Block4{}, Block4Mut{}, 4112, []sf{
		{"head", 0, 96},
		{"body", 96, 4016},
	})
	ss(StreamStats{}, StreamStatsMut{}, 48, []sf{
		{"storage", 0, 24},
		{"appender", 24, 24},
	})
	ss(Started{}, StartedMut{}, 48, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"writerID", 32, 8},
		{"stops", 40, 8},
	})
	ss(EOS{}, EOSMut{}, 48, []sf{
		{"recordID", 0, 24},
		{"timestamp", 24, 8},
		{"writerID", 32, 8},
		{"closed", 40, 1},
		{"waiting", 41, 1},
		{"_", 42, 6},
	})

}
