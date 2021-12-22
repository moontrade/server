package pricing

import "sort"

var HighResolution = SortDurations([]int64{
	//Tick,
	Second,
	Second * 3,
	Second * 5,
	Second * 10,
	Second * 15,
	Second * 30,
	Minute,
	Minute * 2,
	Minute * 3,
	Minute * 5,
	Minute * 10,
	Minute * 15,
	Minute * 30,
	Hour,
	Hour * 4,
	Hour * 6,
	Hour * 8,
	Hour * 12,
	//Day,
	//Day * 7,
	//Day * 30,
})

// durationSlice attaches the methods of Interface to []int, sorting in increasing order.
type durationSlice []int64

func (x durationSlice) Len() int           { return len(x) }
func (x durationSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x durationSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func SortDurations(durations []int64) []int64 {
	d := durationSlice(durations)
	sort.Sort(d)
	return d
}

type Duration int64

const (
	Tick   = int64(1000)
	Second = int64(1000)
	Minute = Second * 60
	Hour   = Minute * 60
	Day    = Hour * 24
	Week   = Day * 7
	Month  = Day * 30
	Year   = Day * 365
)

func IsTimeAligned(now int64, duration int64) bool {
	switch {
	case duration <= Second:
		return (now%Second)%duration == 0
	case duration <= Minute:
		return (now%Minute)%duration == 0
	case duration <= Hour:
		return (now%Hour)%duration == 0
	case duration <= Day:
		return (now%Day)%duration == 0
	case duration <= Week:
		return (now%Week)%duration == 0
	case duration <= Month:
		return (now%Month)%duration == 0
	case duration <= Year:
		return (now%Year)%duration == 0
	}
	return false
}
