package token

//go:generate stringer -type=Kind -linecomment
type Kind int

const (
	KindInvalid   Kind = iota // invalid
	KindBool                  // bool
	KindInt                   // int
	KindInt8                  // int8
	KindInt16                 // int16
	KindInt32                 // int32
	KindInt64                 // int64
	KindFloat32               // float32
	KindFloat64               // float64
	KindByte                  // byte
	KindBytes                 // bytes
	KindString                // string
	KindAny                   // any
	KindMap                   // map
	KindVector                // vector
	KindArray                 // array
	KindEnum                  // enum
	KindStruct                // struct
	KindInterface             // interface
)

var PrimitiveKinds = []Kind{
	KindBool,
	KindInt,
	KindInt8,
	KindInt16,
	KindInt32,
	KindInt64,
	KindFloat32,
	KindFloat64,
	KindByte,
	KindBytes,
	KindString,
	KindAny,
}
