//go:build 386 || amd64 || arm || arm64 || ppc64le || mips64le || mipsle || riscv64 || wasm
// +build 386 amd64 arm arm64 ppc64le mips64le mipsle riscv64 wasm

package pricing

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

type VolumeSide struct {
	total    float64
	interest float64
	percent  float64
}

func (s *VolumeSide) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *VolumeSide) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["total"] = s.Total()
	m["interest"] = s.Interest()
	m["percent"] = s.Percent()
	return m
}

func (s *VolumeSide) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *VolumeSide) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *VolumeSide) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *VolumeSide) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *VolumeSide) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*VolumeSide)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *VolumeSide) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*VolumeSide)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *VolumeSide) Clone() *VolumeSide {
	v := &VolumeSide{}
	*v = *s
	return v
}
func (s *VolumeSide) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *VolumeSide) Mut() *VolumeSideMut {
	return (*VolumeSideMut)(unsafe.Pointer(s))
}
func (s *VolumeSide) Total() float64 {
	return s.total
}
func (s *VolumeSide) Interest() float64 {
	return s.interest
}
func (s *VolumeSide) Percent() float64 {
	return s.percent
}

type VolumeSideMut struct {
	VolumeSide
}

func (s *VolumeSideMut) Clone() *VolumeSideMut {
	v := &VolumeSideMut{}
	*v = *s
	return v
}
func (s *VolumeSideMut) Freeze() *VolumeSide {
	return (*VolumeSide)(unsafe.Pointer(s))
}
func (s *VolumeSideMut) SetTotal(v float64) *VolumeSideMut {
	s.total = v
	return s
}
func (s *VolumeSideMut) SetInterest(v float64) *VolumeSideMut {
	s.interest = v
	return s
}
func (s *VolumeSideMut) SetPercent(v float64) *VolumeSideMut {
	s.percent = v
	return s
}

type Volume struct {
	total float64
	buy   VolumeSide
	sell  VolumeSide
}

func (s *Volume) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Volume) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["total"] = s.Total()
	m["buy"] = s.Buy().MarshalMap(nil)
	m["sell"] = s.Sell().MarshalMap(nil)
	return m
}

func (s *Volume) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[56]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 56 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Volume) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[56]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Volume) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[56]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Volume) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[56]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Volume) Read(b []byte) (n int, err error) {
	if len(b) < 56 {
		return -1, io.ErrShortBuffer
	}
	v := (*Volume)(unsafe.Pointer(&b[0]))
	*v = *s
	return 56, nil
}
func (s *Volume) UnmarshalBinary(b []byte) error {
	if len(b) < 56 {
		return io.ErrShortBuffer
	}
	v := (*Volume)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Volume) Clone() *Volume {
	v := &Volume{}
	*v = *s
	return v
}
func (s *Volume) Bytes() []byte {
	return (*(*[56]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Volume) Mut() *VolumeMut {
	return (*VolumeMut)(unsafe.Pointer(s))
}
func (s *Volume) Total() float64 {
	return s.total
}
func (s *Volume) Buy() *VolumeSide {
	return &s.buy
}
func (s *Volume) Sell() *VolumeSide {
	return &s.sell
}

type VolumeMut struct {
	Volume
}

func (s *VolumeMut) Clone() *VolumeMut {
	v := &VolumeMut{}
	*v = *s
	return v
}
func (s *VolumeMut) Freeze() *Volume {
	return (*Volume)(unsafe.Pointer(s))
}
func (s *VolumeMut) SetTotal(v float64) *VolumeMut {
	s.total = v
	return s
}
func (s *VolumeMut) Buy() *VolumeSideMut {
	return s.buy.Mut()
}
func (s *VolumeMut) SetBuy(v *VolumeSide) *VolumeMut {
	s.buy = *v
	return s
}
func (s *VolumeMut) Sell() *VolumeSideMut {
	return s.sell.Mut()
}
func (s *VolumeMut) SetSell(v *VolumeSide) *VolumeMut {
	s.sell = *v
	return s
}

type Liquidations struct {
	trades int64
	min    float64
	avg    float64
	max    float64
	buys   float64
	sells  float64
	value  float64
}

func (s *Liquidations) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Liquidations) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["trades"] = s.Trades()
	m["min"] = s.Min()
	m["avg"] = s.Avg()
	m["max"] = s.Max()
	m["buys"] = s.Buys()
	m["sells"] = s.Sells()
	m["value"] = s.Value()
	return m
}

func (s *Liquidations) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[56]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 56 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Liquidations) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[56]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Liquidations) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[56]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Liquidations) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[56]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Liquidations) Read(b []byte) (n int, err error) {
	if len(b) < 56 {
		return -1, io.ErrShortBuffer
	}
	v := (*Liquidations)(unsafe.Pointer(&b[0]))
	*v = *s
	return 56, nil
}
func (s *Liquidations) UnmarshalBinary(b []byte) error {
	if len(b) < 56 {
		return io.ErrShortBuffer
	}
	v := (*Liquidations)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Liquidations) Clone() *Liquidations {
	v := &Liquidations{}
	*v = *s
	return v
}
func (s *Liquidations) Bytes() []byte {
	return (*(*[56]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Liquidations) Mut() *LiquidationsMut {
	return (*LiquidationsMut)(unsafe.Pointer(s))
}
func (s *Liquidations) Trades() int64 {
	return s.trades
}
func (s *Liquidations) Min() float64 {
	return s.min
}
func (s *Liquidations) Avg() float64 {
	return s.avg
}
func (s *Liquidations) Max() float64 {
	return s.max
}
func (s *Liquidations) Buys() float64 {
	return s.buys
}
func (s *Liquidations) Sells() float64 {
	return s.sells
}
func (s *Liquidations) Value() float64 {
	return s.value
}

type LiquidationsMut struct {
	Liquidations
}

func (s *LiquidationsMut) Clone() *LiquidationsMut {
	v := &LiquidationsMut{}
	*v = *s
	return v
}
func (s *LiquidationsMut) Freeze() *Liquidations {
	return (*Liquidations)(unsafe.Pointer(s))
}
func (s *LiquidationsMut) SetTrades(v int64) *LiquidationsMut {
	s.trades = v
	return s
}
func (s *LiquidationsMut) SetMin(v float64) *LiquidationsMut {
	s.min = v
	return s
}
func (s *LiquidationsMut) SetAvg(v float64) *LiquidationsMut {
	s.avg = v
	return s
}
func (s *LiquidationsMut) SetMax(v float64) *LiquidationsMut {
	s.max = v
	return s
}
func (s *LiquidationsMut) SetBuys(v float64) *LiquidationsMut {
	s.buys = v
	return s
}
func (s *LiquidationsMut) SetSells(v float64) *LiquidationsMut {
	s.sells = v
	return s
}
func (s *LiquidationsMut) SetValue(v float64) *LiquidationsMut {
	s.value = v
	return s
}

type Bar struct {
	time         Time
	precision    float64
	price        Candle
	bid          Candle
	ask          Candle
	spread       Spread
	ticks        int64
	volume       Volume
	trades       Trades
	greeks       Greeks
	liquidations Liquidations
}

func (s *Bar) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Bar) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["time"] = s.Time().MarshalMap(nil)
	m["precision"] = s.Precision()
	m["price"] = s.Price().MarshalMap(nil)
	m["bid"] = s.Bid().MarshalMap(nil)
	m["ask"] = s.Ask().MarshalMap(nil)
	m["spread"] = s.Spread().MarshalMap(nil)
	m["ticks"] = s.Ticks()
	m["volume"] = s.Volume().MarshalMap(nil)
	m["trades"] = s.Trades().MarshalMap(nil)
	m["greeks"] = s.Greeks().MarshalMap(nil)
	m["liquidations"] = s.Liquidations().MarshalMap(nil)
	return m
}

func (s *Bar) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[344]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 344 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Bar) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[344]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Bar) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[344]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Bar) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[344]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Bar) Read(b []byte) (n int, err error) {
	if len(b) < 344 {
		return -1, io.ErrShortBuffer
	}
	v := (*Bar)(unsafe.Pointer(&b[0]))
	*v = *s
	return 344, nil
}
func (s *Bar) UnmarshalBinary(b []byte) error {
	if len(b) < 344 {
		return io.ErrShortBuffer
	}
	v := (*Bar)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Bar) Clone() *Bar {
	v := &Bar{}
	*v = *s
	return v
}
func (s *Bar) Bytes() []byte {
	return (*(*[344]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Bar) Mut() *BarMut {
	return (*BarMut)(unsafe.Pointer(s))
}
func (s *Bar) Time() *Time {
	return &s.time
}
func (s *Bar) Precision() float64 {
	return s.precision
}
func (s *Bar) Price() *Candle {
	return &s.price
}
func (s *Bar) Bid() *Candle {
	return &s.bid
}
func (s *Bar) Ask() *Candle {
	return &s.ask
}
func (s *Bar) Spread() *Spread {
	return &s.spread
}
func (s *Bar) Ticks() int64 {
	return s.ticks
}
func (s *Bar) Volume() *Volume {
	return &s.volume
}
func (s *Bar) Trades() *Trades {
	return &s.trades
}
func (s *Bar) Greeks() *Greeks {
	return &s.greeks
}
func (s *Bar) Liquidations() *Liquidations {
	return &s.liquidations
}

type BarMut struct {
	Bar
}

func (s *BarMut) Clone() *BarMut {
	v := &BarMut{}
	*v = *s
	return v
}
func (s *BarMut) Freeze() *Bar {
	return (*Bar)(unsafe.Pointer(s))
}
func (s *BarMut) Time() *TimeMut {
	return s.time.Mut()
}
func (s *BarMut) SetTime(v *Time) *BarMut {
	s.time = *v
	return s
}
func (s *BarMut) SetPrecision(v float64) *BarMut {
	s.precision = v
	return s
}
func (s *BarMut) Price() *CandleMut {
	return s.price.Mut()
}
func (s *BarMut) SetPrice(v *Candle) *BarMut {
	s.price = *v
	return s
}
func (s *BarMut) Bid() *CandleMut {
	return s.bid.Mut()
}
func (s *BarMut) SetBid(v *Candle) *BarMut {
	s.bid = *v
	return s
}
func (s *BarMut) Ask() *CandleMut {
	return s.ask.Mut()
}
func (s *BarMut) SetAsk(v *Candle) *BarMut {
	s.ask = *v
	return s
}
func (s *BarMut) Spread() *SpreadMut {
	return s.spread.Mut()
}
func (s *BarMut) SetSpread(v *Spread) *BarMut {
	s.spread = *v
	return s
}
func (s *BarMut) SetTicks(v int64) *BarMut {
	s.ticks = v
	return s
}
func (s *BarMut) Volume() *VolumeMut {
	return s.volume.Mut()
}
func (s *BarMut) SetVolume(v *Volume) *BarMut {
	s.volume = *v
	return s
}
func (s *BarMut) Trades() *TradesMut {
	return s.trades.Mut()
}
func (s *BarMut) SetTrades(v *Trades) *BarMut {
	s.trades = *v
	return s
}
func (s *BarMut) Greeks() *GreeksMut {
	return s.greeks.Mut()
}
func (s *BarMut) SetGreeks(v *Greeks) *BarMut {
	s.greeks = *v
	return s
}
func (s *BarMut) Liquidations() *LiquidationsMut {
	return s.liquidations.Mut()
}
func (s *BarMut) SetLiquidations(v *Liquidations) *BarMut {
	s.liquidations = *v
	return s
}

type Time struct {
	start    int64
	duration int64
	end      int64
}

func (s *Time) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Time) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["start"] = s.Start()
	m["duration"] = s.Duration()
	m["end"] = s.End()
	return m
}

func (s *Time) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Time) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Time) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Time) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Time) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*Time)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *Time) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*Time)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Time) Clone() *Time {
	v := &Time{}
	*v = *s
	return v
}
func (s *Time) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Time) Mut() *TimeMut {
	return (*TimeMut)(unsafe.Pointer(s))
}
func (s *Time) Start() int64 {
	return s.start
}
func (s *Time) Duration() int64 {
	return s.duration
}
func (s *Time) End() int64 {
	return s.end
}

type TimeMut struct {
	Time
}

func (s *TimeMut) Clone() *TimeMut {
	v := &TimeMut{}
	*v = *s
	return v
}
func (s *TimeMut) Freeze() *Time {
	return (*Time)(unsafe.Pointer(s))
}
func (s *TimeMut) SetStart(v int64) *TimeMut {
	s.start = v
	return s
}
func (s *TimeMut) SetDuration(v int64) *TimeMut {
	s.duration = v
	return s
}
func (s *TimeMut) SetEnd(v int64) *TimeMut {
	s.end = v
	return s
}

// Candlestick
type Candle struct {
	open  float64
	high  float64
	low   float64
	close float64
}

func (s *Candle) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Candle) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["open"] = s.Open()
	m["high"] = s.High()
	m["low"] = s.Low()
	m["close"] = s.Close()
	return m
}

func (s *Candle) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[32]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 32 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Candle) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[32]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Candle) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[32]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Candle) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[32]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Candle) Read(b []byte) (n int, err error) {
	if len(b) < 32 {
		return -1, io.ErrShortBuffer
	}
	v := (*Candle)(unsafe.Pointer(&b[0]))
	*v = *s
	return 32, nil
}
func (s *Candle) UnmarshalBinary(b []byte) error {
	if len(b) < 32 {
		return io.ErrShortBuffer
	}
	v := (*Candle)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Candle) Clone() *Candle {
	v := &Candle{}
	*v = *s
	return v
}
func (s *Candle) Bytes() []byte {
	return (*(*[32]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Candle) Mut() *CandleMut {
	return (*CandleMut)(unsafe.Pointer(s))
}
func (s *Candle) Open() float64 {
	return s.open
}
func (s *Candle) High() float64 {
	return s.high
}
func (s *Candle) Low() float64 {
	return s.low
}
func (s *Candle) Close() float64 {
	return s.close
}

// Candlestick
type CandleMut struct {
	Candle
}

func (c *CandleMut) Clone() *CandleMut {
	v := &CandleMut{}
	*v = *c
	return v
}
func (c *CandleMut) Freeze() *Candle {
	return (*Candle)(unsafe.Pointer(c))
}
func (c *CandleMut) SetOpen(v float64) *CandleMut {
	c.open = v
	return c
}
func (c *CandleMut) SetHigh(v float64) *CandleMut {
	c.high = v
	return c
}
func (c *CandleMut) SetLow(v float64) *CandleMut {
	c.low = v
	return c
}
func (c *CandleMut) SetClose(v float64) *CandleMut {
	c.close = v
	return c
}

type Spread struct {
	low  float64
	mid  float64
	high float64
}

func (s *Spread) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Spread) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["low"] = s.Low()
	m["mid"] = s.Mid()
	m["high"] = s.High()
	return m
}

func (s *Spread) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Spread) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Spread) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Spread) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Spread) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*Spread)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *Spread) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*Spread)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Spread) Clone() *Spread {
	v := &Spread{}
	*v = *s
	return v
}
func (s *Spread) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Spread) Mut() *SpreadMut {
	return (*SpreadMut)(unsafe.Pointer(s))
}
func (s *Spread) Low() float64 {
	return s.low
}
func (s *Spread) Mid() float64 {
	return s.mid
}
func (s *Spread) High() float64 {
	return s.high
}

type SpreadMut struct {
	Spread
}

func (s *SpreadMut) Clone() *SpreadMut {
	v := &SpreadMut{}
	*v = *s
	return v
}
func (s *SpreadMut) Freeze() *Spread {
	return (*Spread)(unsafe.Pointer(s))
}
func (s *SpreadMut) SetLow(v float64) *SpreadMut {
	s.low = v
	return s
}
func (s *SpreadMut) SetMid(v float64) *SpreadMut {
	s.mid = v
	return s
}
func (s *SpreadMut) SetHigh(v float64) *SpreadMut {
	s.high = v
	return s
}

type Trades struct {
	count int64
	min   int64
	max   int64
}

func (s *Trades) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Trades) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["count"] = s.Count()
	m["min"] = s.Min()
	m["max"] = s.Max()
	return m
}

func (s *Trades) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[24]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 24 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Trades) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[24]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Trades) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Trades) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[24]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Trades) Read(b []byte) (n int, err error) {
	if len(b) < 24 {
		return -1, io.ErrShortBuffer
	}
	v := (*Trades)(unsafe.Pointer(&b[0]))
	*v = *s
	return 24, nil
}
func (s *Trades) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return io.ErrShortBuffer
	}
	v := (*Trades)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Trades) Clone() *Trades {
	v := &Trades{}
	*v = *s
	return v
}
func (s *Trades) Bytes() []byte {
	return (*(*[24]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Trades) Mut() *TradesMut {
	return (*TradesMut)(unsafe.Pointer(s))
}
func (s *Trades) Count() int64 {
	return s.count
}
func (s *Trades) Min() int64 {
	return s.min
}
func (s *Trades) Max() int64 {
	return s.max
}

type TradesMut struct {
	Trades
}

func (s *TradesMut) Clone() *TradesMut {
	v := &TradesMut{}
	*v = *s
	return v
}
func (s *TradesMut) Freeze() *Trades {
	return (*Trades)(unsafe.Pointer(s))
}
func (s *TradesMut) SetCount(v int64) *TradesMut {
	s.count = v
	return s
}
func (s *TradesMut) SetMin(v int64) *TradesMut {
	s.min = v
	return s
}
func (s *TradesMut) SetMax(v int64) *TradesMut {
	s.max = v
	return s
}

// Greeks are financial measures of the sensitivity of an option’s price to its
// underlying determining parameters, such as volatility or the price of the underlying
// asset. The Greeks are utilized in the analysis of an options portfolio and in sensitivity
// analysis of an option or portfolio of options. The measures are considered essential by
// many investors for making informed decisions in options trading.
//
// Delta, Gamma, Vega, Theta, and Rho are the key option Greeks. However, there are many other
// option Greeks that can be derived from those mentioned above.
type Greeks struct {
	iv    float64
	delta float64
	gamma float64
	vega  float64
	theta float64
	rho   float64
}

func (s *Greeks) String() string {
	return fmt.Sprintf("%v", s.MarshalMap(nil))
}

func (s *Greeks) MarshalMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["iv"] = s.Iv()
	m["delta"] = s.Delta()
	m["gamma"] = s.Gamma()
	m["vega"] = s.Vega()
	m["theta"] = s.Theta()
	m["rho"] = s.Rho()
	return m
}

func (s *Greeks) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.ReadFull(r, (*(*[48]byte)(unsafe.Pointer(s)))[0:])
	if err != nil {
		return int64(n), err
	}
	if n != 48 {
		return int64(n), io.ErrShortBuffer
	}
	return int64(n), nil
}
func (s *Greeks) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write((*(*[48]byte)(unsafe.Pointer(s)))[0:])
	return int64(n), err
}
func (s *Greeks) MarshalBinaryTo(b []byte) []byte {
	return append(b, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...)
}
func (s *Greeks) MarshalBinary() ([]byte, error) {
	var v []byte
	return append(v, (*(*[48]byte)(unsafe.Pointer(s)))[0:]...), nil
}
func (s *Greeks) Read(b []byte) (n int, err error) {
	if len(b) < 48 {
		return -1, io.ErrShortBuffer
	}
	v := (*Greeks)(unsafe.Pointer(&b[0]))
	*v = *s
	return 48, nil
}
func (s *Greeks) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return io.ErrShortBuffer
	}
	v := (*Greeks)(unsafe.Pointer(&b[0]))
	*s = *v
	return nil
}
func (s *Greeks) Clone() *Greeks {
	v := &Greeks{}
	*v = *s
	return v
}
func (s *Greeks) Bytes() []byte {
	return (*(*[48]byte)(unsafe.Pointer(s)))[0:]
}
func (s *Greeks) Mut() *GreeksMut {
	return (*GreeksMut)(unsafe.Pointer(s))
}
func (s *Greeks) Iv() float64 {
	return s.iv
}
func (s *Greeks) Delta() float64 {
	return s.delta
}
func (s *Greeks) Gamma() float64 {
	return s.gamma
}
func (s *Greeks) Vega() float64 {
	return s.vega
}
func (s *Greeks) Theta() float64 {
	return s.theta
}
func (s *Greeks) Rho() float64 {
	return s.rho
}

// Greeks are financial measures of the sensitivity of an option’s price to its
// underlying determining parameters, such as volatility or the price of the underlying
// asset. The Greeks are utilized in the analysis of an options portfolio and in sensitivity
// analysis of an option or portfolio of options. The measures are considered essential by
// many investors for making informed decisions in options trading.
//
// Delta, Gamma, Vega, Theta, and Rho are the key option Greeks. However, there are many other
// option Greeks that can be derived from those mentioned above.
type GreeksMut struct {
	Greeks
}

func (s *GreeksMut) Clone() *GreeksMut {
	v := &GreeksMut{}
	*v = *s
	return v
}
func (s *GreeksMut) Freeze() *Greeks {
	return (*Greeks)(unsafe.Pointer(s))
}
func (s *GreeksMut) SetIv(v float64) *GreeksMut {
	s.iv = v
	return s
}
func (s *GreeksMut) SetDelta(v float64) *GreeksMut {
	s.delta = v
	return s
}
func (s *GreeksMut) SetGamma(v float64) *GreeksMut {
	s.gamma = v
	return s
}
func (s *GreeksMut) SetVega(v float64) *GreeksMut {
	s.vega = v
	return s
}
func (s *GreeksMut) SetTheta(v float64) *GreeksMut {
	s.theta = v
	return s
}
func (s *GreeksMut) SetRho(v float64) *GreeksMut {
	s.rho = v
	return s
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

	ss(VolumeSide{}, VolumeSideMut{}, 24, []sf{
		{"total", 0, 8},
		{"interest", 8, 8},
		{"percent", 16, 8},
	})
	ss(Volume{}, VolumeMut{}, 56, []sf{
		{"total", 0, 8},
		{"buy", 8, 24},
		{"sell", 32, 24},
	})
	ss(Liquidations{}, LiquidationsMut{}, 56, []sf{
		{"trades", 0, 8},
		{"min", 8, 8},
		{"avg", 16, 8},
		{"max", 24, 8},
		{"buys", 32, 8},
		{"sells", 40, 8},
		{"value", 48, 8},
	})
	ss(Bar{}, BarMut{}, 344, []sf{
		{"time", 0, 24},
		{"precision", 24, 8},
		{"price", 32, 32},
		{"bid", 64, 32},
		{"ask", 96, 32},
		{"spread", 128, 24},
		{"ticks", 152, 8},
		{"volume", 160, 56},
		{"trades", 216, 24},
		{"greeks", 240, 48},
		{"liquidations", 288, 56},
	})
	ss(Time{}, TimeMut{}, 24, []sf{
		{"start", 0, 8},
		{"duration", 8, 8},
		{"end", 16, 8},
	})
	ss(Candle{}, CandleMut{}, 32, []sf{
		{"open", 0, 8},
		{"high", 8, 8},
		{"low", 16, 8},
		{"close", 24, 8},
	})
	ss(Spread{}, SpreadMut{}, 24, []sf{
		{"low", 0, 8},
		{"mid", 8, 8},
		{"high", 16, 8},
	})
	ss(Trades{}, TradesMut{}, 24, []sf{
		{"count", 0, 8},
		{"min", 8, 8},
		{"max", 16, 8},
	})
	ss(Greeks{}, GreeksMut{}, 48, []sf{
		{"iv", 0, 8},
		{"delta", 8, 8},
		{"gamma", 16, 8},
		{"vega", 24, 8},
		{"theta", 32, 8},
		{"rho", 40, 8},
	})

}
