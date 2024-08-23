package types

//go:generate stringer -type=Kind
type Kind int

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Byte
	Bytes
	String
	Map
	Vector
	Array
	Struct
	Enum
)
