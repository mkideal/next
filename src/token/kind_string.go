// Code generated by "stringer -type=Kind -linecomment"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[KindInvalid-0]
	_ = x[KindBool-1]
	_ = x[KindInt-2]
	_ = x[KindInt8-3]
	_ = x[KindInt16-4]
	_ = x[KindInt32-5]
	_ = x[KindInt64-6]
	_ = x[KindFloat32-7]
	_ = x[KindFloat64-8]
	_ = x[KindByte-9]
	_ = x[KindBytes-10]
	_ = x[KindString-11]
	_ = x[KindAny-12]
	_ = x[KindMap-13]
	_ = x[KindVector-14]
	_ = x[KindArray-15]
	_ = x[KindEnum-16]
	_ = x[KindStruct-17]
	_ = x[KindInterface-18]
}

const _Kind_name = "invalidboolintint8int16int32int64float32float64bytebytesstringanymapvectorarrayenumstructinterface"

var _Kind_index = [...]uint8{0, 7, 11, 14, 18, 23, 28, 33, 40, 47, 51, 56, 62, 65, 68, 74, 79, 83, 89, 98}

func (i Kind) String() string {
	if i < 0 || i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
