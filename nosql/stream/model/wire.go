package model

// WireNumber represents the field number.
type WireNumber int32

const (
	MinValidNumber      WireNumber = 1
	FirstReservedNumber WireNumber = 19000
	LastReservedNumber  WireNumber = 19999
	MaxValidNumber      WireNumber = 1<<29 - 1
)

// IsValid reports whether the field number is semantically valid.
//
// Note that while numbers within the reserved range are semantically invalid,
// they are syntactically valid in the wire format.
// Implementations may treat records with reserved field numbers as unknown.
func (n WireNumber) IsValid() bool {
	return MinValidNumber <= n && n < FirstReservedNumber || LastReservedNumber < n && n <= MaxValidNumber
}

// WireType represents the wire type.
type WireType int8

const (
	VarintType     WireType = 0
	Fixed32Type    WireType = 5
	Fixed64Type    WireType = 1
	BytesType      WireType = 2
	StartGroupType WireType = 3
	EndGroupType   WireType = 4
)
