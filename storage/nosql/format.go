package nosql

type Format byte

const (
	FormatRaw      Format = 0
	FormatJson     Format = 1
	FormatMsgpack  Format = 2
	FormatBeam     Format = 3
	FormatProtobuf Format = 4
	FormatIndex    Format = 5
)
