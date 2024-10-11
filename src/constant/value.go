// Package constant implements Values representing untyped
// Next constants and their corresponding operations.
package constant

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"sync"

	"github.com/mkideal/next/src/token"
)

//go:generate golang.org/x/tools/cmd/stringer@latest -type Kind

// Kind specifies the kind of value represented by a Value.
type Kind int

const (
	Unknown Kind = iota // unknown values
	Bool                // boolean values
	String              // string values
	Int                 // integer values
	Float               // floating-point values
)

// Value represents the value of a Go constant.
type Value interface {
	// Kind returns the value kind.
	Kind() Kind

	// String returns a short, quoted (human-readable) form of the value.
	// For numeric values, the result may be an approximation;
	// for String values the result may be a shortened string.
	// Use ExactString for a string representing a value exactly.
	String() string

	// ExactString returns an exact, quoted (human-readable) form of the value.
	// If the Value is of Kind String, use StringVal to obtain the unquoted string.
	ExactString() string

	// Prevent external implementations.
	implementsValue()
}

// Maximum supported mantissa precision.
// The spec requires at least 256 bits; typical implementations use 512 bits.
const prec = 512

type (
	unknownVal struct{}
	boolVal    bool
	stringVal  struct {
		// Lazy value: either a string (l,r==nil) or an addition (l,r!=nil).
		mu   sync.Mutex
		s    string
		l, r *stringVal
	}
	int64Val int64                    // Int values representable as an int64
	intVal   struct{ val *big.Int }   // Int values not representable as an int64
	ratVal   struct{ val *big.Rat }   // Float values representable as a fraction
	floatVal struct{ val *big.Float } // Float values not representable as a fraction
)

func (unknownVal) Kind() Kind { return Unknown }
func (boolVal) Kind() Kind    { return Bool }
func (*stringVal) Kind() Kind { return String }
func (int64Val) Kind() Kind   { return Int }
func (intVal) Kind() Kind     { return Int }
func (ratVal) Kind() Kind     { return Float }
func (floatVal) Kind() Kind   { return Float }

func (unknownVal) String() string { return "undefined" }
func (x boolVal) String() string  { return strconv.FormatBool(bool(x)) }

// String returns a possibly shortened quoted form of the String value.
func (x *stringVal) String() string {
	return strconv.Quote(x.string())
}

// string constructs and returns the actual string literal value.
// If x represents an addition, then it rewrites x to be a single
// string, to speed future calls.
func (x *stringVal) string() string {
	x.mu.Lock()
	if x.l != nil {
		x.s = strings.Join(reverse(x.appendReverse(nil)), "")
		x.l = nil
		x.r = nil
	}
	s := x.s
	x.mu.Unlock()

	return s
}

// reverse reverses x in place and returns it.
func reverse(x []string) []string {
	n := len(x)
	for i := 0; i+i < n; i++ {
		x[i], x[n-1-i] = x[n-1-i], x[i]
	}
	return x
}

// appendReverse appends to list all of x's subpieces, but in reverse,
// and returns the result.
func (x *stringVal) appendReverse(list []string) []string {
	y := x
	for y.r != nil {
		y.r.mu.Lock()
		list = y.r.appendReverse(list)
		y.r.mu.Unlock()

		l := y.l
		if y != x {
			y.mu.Unlock()
		}
		l.mu.Lock()
		y = l
	}
	s := y.s
	if y != x {
		y.mu.Unlock()
	}
	return append(list, s)
}

func (x int64Val) String() string { return strconv.FormatInt(int64(x), 10) }
func (x intVal) String() string   { return x.val.String() }
func (x ratVal) String() string   { return rtof(x).String() }

// String returns a decimal approximation of the Float value.
func (x floatVal) String() string {
	f := x.val

	// Handle special cases
	if f.IsInf() {
		if f.Sign() > 0 {
			return "+Inf"
		}
		return "-Inf"
	}

	// Convert to string with full precision
	s := f.Text('g', -1)

	// If the number is zero
	if f.Sign() == 0 {
		return "0.0"
	}

	// For numbers in scientific notation
	if strings.ContainsAny(s, "eE") {
		parts := strings.Split(strings.ToLower(s), "e")
		mantissa := parts[0]
		exponent := parts[1]

		// Remove trailing zeros from mantissa
		mantissa = strings.TrimRight(strings.TrimRight(mantissa, "0"), ".")

		// If mantissa is empty, it was 1 or -1, so restore it
		if mantissa == "" || mantissa == "-" {
			mantissa += "1"
		}

		return fmt.Sprintf("%se%s", mantissa, exponent)
	}

	// For regular numbers, remove trailing zeros
	if strings.Contains(s, ".") {
		s = strings.TrimSuffix(strings.TrimRight(s, "0"), ".")
	}

	// If the number is an integer, add a trailing ".0"
	if !strings.Contains(s, ".") {
		s += ".0"
	}

	return s
}

func (x unknownVal) ExactString() string { return x.String() }
func (x boolVal) ExactString() string    { return x.String() }
func (x *stringVal) ExactString() string { return strconv.Quote(x.string()) }
func (x int64Val) ExactString() string   { return x.String() }
func (x intVal) ExactString() string     { return x.String() }

func (x ratVal) ExactString() string {
	r := x.val
	if r.IsInt() {
		return r.Num().String()
	}
	return r.String()
}

func (x floatVal) ExactString() string { return x.val.Text('p', 0) }

func (unknownVal) implementsValue() {}
func (boolVal) implementsValue()    {}
func (*stringVal) implementsValue() {}
func (int64Val) implementsValue()   {}
func (ratVal) implementsValue()     {}
func (intVal) implementsValue()     {}
func (floatVal) implementsValue()   {}

func newInt() *big.Int     { return new(big.Int) }
func newRat() *big.Rat     { return new(big.Rat) }
func newFloat() *big.Float { return new(big.Float).SetPrec(prec) }

func i64toi(x int64Val) intVal   { return intVal{newInt().SetInt64(int64(x))} }
func i64tor(x int64Val) ratVal   { return ratVal{newRat().SetInt64(int64(x))} }
func i64tof(x int64Val) floatVal { return floatVal{newFloat().SetInt64(int64(x))} }
func itor(x intVal) ratVal       { return ratVal{newRat().SetInt(x.val)} }
func itof(x intVal) floatVal     { return floatVal{newFloat().SetInt(x.val)} }
func rtof(x ratVal) floatVal     { return floatVal{newFloat().SetRat(x.val)} }

func makeInt(x *big.Int) Value {
	if x.IsInt64() {
		return int64Val(x.Int64())
	}
	return intVal{x}
}

func makeRat(x *big.Rat) Value {
	a := x.Num()
	b := x.Denom()
	if smallInt(a) && smallInt(b) {
		return ratVal{x}
	}
	return floatVal{newFloat().SetRat(x)}
}

var floatVal0 = floatVal{newFloat()}

func makeFloat(x *big.Float) Value {
	if x.Sign() == 0 {
		return floatVal0
	}
	if x.IsInf() {
		return unknownVal{}
	}
	return floatVal{x}
}

// makeFloatFromLiteral creates a Value from a floating-point literal.
func makeFloatFromLiteral(lit string) Value {
	if f, ok := newFloat().SetString(lit); ok {
		if smallFloat(f) {
			if f.Sign() == 0 {
				lit = "0"
			}
			if r, ok := newRat().SetString(lit); ok {
				return ratVal{r}
			}
		}
		return makeFloat(f)
	}
	return nil
}

// Permit fractions with component sizes up to maxExp
// before switching to using floating-point numbers.
const maxExp = 4 << 10

// smallInt reports whether x would lead to a "reasonably"-sized fraction
// if converted to a *big.Rat.
func smallInt(x *big.Int) bool {
	return x.BitLen() < maxExp
}

// smallFloat64 reports whether x would lead to a "reasonably"-sized fraction
// if converted to a *big.Rat.
func smallFloat64(x float64) bool {
	if math.IsInf(x, 0) {
		return false
	}
	_, e := math.Frexp(x)
	return -maxExp < e && e < maxExp
}

// smallFloat reports whether x would lead to a "reasonably"-sized fraction
// if converted to a *big.Rat.
func smallFloat(x *big.Float) bool {
	if x.IsInf() {
		return false
	}
	e := x.MantExp(nil)
	return -maxExp < e && e < maxExp
}

// MakeUnknown returns the Unknown value.
func MakeUnknown() Value { return unknownVal{} }

// MakeBool returns the Bool value for b.
func MakeBool(b bool) Value { return boolVal(b) }

// MakeString returns the String value for s.
func MakeString(s string) Value {
	if s == "" {
		return &emptyString
	}
	return &stringVal{s: s}
}

var emptyString stringVal

// MakeInt64 returns the Int value for x.
func MakeInt64(x int64) Value { return int64Val(x) }

// MakeUint64 returns the Int value for x.
func MakeUint64(x uint64) Value {
	if x < 1<<63 {
		return int64Val(int64(x))
	}
	return intVal{newInt().SetUint64(x)}
}

// MakeFloat64 returns the Float value for x.
// If x is -0.0, the result is 0.0.
// If x is not finite, the result is Unknown.
func MakeFloat64(x float64) Value {
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return unknownVal{}
	}
	if smallFloat64(x) {
		return ratVal{newRat().SetFloat64(x + 0)} // convert -0 to 0
	}
	return floatVal{newFloat().SetFloat64(x + 0)}
}

// MakeFromLiteral returns the corresponding integer, floating-point,
// imaginary, character, or string value for a Go literal string.
// The tok value must be one of token.INT, token.FLOAT, token.IMAG,
// token.CHAR, or token.STRING. The final argument must be zero.
// If the literal string syntax is invalid, the result is Unknown.
func MakeFromLiteral(lit string, tok token.Token, zero uint) Value {
	if zero != 0 {
		panic("MakeFromLiteral called with non-zero last argument")
	}

	switch tok {
	case token.INT:
		if x, err := strconv.ParseInt(lit, 0, 64); err == nil {
			return int64Val(x)
		}
		if x, ok := newInt().SetString(lit, 0); ok {
			return intVal{x}
		}

	case token.FLOAT:
		if x := makeFloatFromLiteral(lit); x != nil {
			return x
		}

	case token.CHAR:
		if n := len(lit); n >= 2 {
			if code, _, _, err := strconv.UnquoteChar(lit[1:n-1], '\''); err == nil {
				return MakeInt64(int64(code))
			}
		}

	case token.STRING:
		if s, err := strconv.Unquote(lit); err == nil {
			return MakeString(s)
		}

	default:
		panic(fmt.Sprintf("%v is not a valid token", tok))
	}

	return unknownVal{}
}

// BoolVal returns the Go boolean value of x, which must be a Bool or an Unknown.
// If x is Unknown, the result is false.
func BoolVal(x Value) bool {
	switch x := x.(type) {
	case boolVal:
		return bool(x)
	case unknownVal:
		return false
	default:
		panic(fmt.Sprintf("%v not a Bool", x))
	}
}

// StringVal returns the Go string value of x, which must be a String or an Unknown.
// If x is Unknown, the result is "".
func StringVal(x Value) string {
	switch x := x.(type) {
	case *stringVal:
		return x.string()
	case unknownVal:
		return ""
	default:
		panic(fmt.Sprintf("%v not a String", x))
	}
}

// Int64Val returns the Go int64 value of x and whether the result is exact;
// x must be an Int or an Unknown. If the result is not exact, its value is undefined.
// If x is Unknown, the result is (0, false).
func Int64Val(x Value) (int64, bool) {
	switch x := x.(type) {
	case int64Val:
		return int64(x), true
	case intVal:
		return x.val.Int64(), false // not an int64Val and thus not exact
	case unknownVal:
		return 0, false
	default:
		panic(fmt.Sprintf("%v not an Int", x))
	}
}

// Uint64Val returns the Go uint64 value of x and whether the result is exact;
// x must be an Int or an Unknown. If the result is not exact, its value is undefined.
// If x is Unknown, the result is (0, false).
func Uint64Val(x Value) (uint64, bool) {
	switch x := x.(type) {
	case int64Val:
		return uint64(x), x >= 0
	case intVal:
		return x.val.Uint64(), x.val.IsUint64()
	case unknownVal:
		return 0, false
	default:
		panic(fmt.Sprintf("%v not an Int", x))
	}
}

// Float32Val returns the nearest Go float32 value of x and whether the result is exact;
// x must be numeric or Unknown. For values too small (too close to 0) to represent
// as float32, Float32Val silently underflows to 0.
// If x is Unknown, the result is (0, false).
func Float32Val(x Value) (float32, bool) {
	switch x := x.(type) {
	case int64Val:
		f := float32(x)
		return f, int64Val(f) == x
	case intVal:
		f, acc := newFloat().SetInt(x.val).Float32()
		return f, acc == big.Exact
	case ratVal:
		return x.val.Float32()
	case floatVal:
		f, acc := x.val.Float32()
		return f, acc == big.Exact
	case unknownVal:
		return 0, false
	default:
		panic(fmt.Sprintf("%v not a Float", x))
	}
}

// Float64Val returns the nearest Go float64 value of x and whether the result is exact;
// x must be numeric or Unknown. For values too small (too close to 0) to represent
// as float64, Float64Val silently underflows to 0. The result sign always
// matches the sign of x, even for 0.
// If x is Unknown, the result is (0, false).
func Float64Val(x Value) (float64, bool) {
	switch x := x.(type) {
	case int64Val:
		f := float64(int64(x))
		return f, int64Val(f) == x
	case intVal:
		f, acc := newFloat().SetInt(x.val).Float64()
		return f, acc == big.Exact
	case ratVal:
		return x.val.Float64()
	case floatVal:
		f, acc := x.val.Float64()
		return f, acc == big.Exact
	case unknownVal:
		return 0, false
	default:
		panic(fmt.Sprintf("%v not a Float", x))
	}
}

// Val returns the underlying value for a given constant. Since it returns an
// interface, it is up to the caller to type assert the result to the expected
// type.
func Val(x Value) any {
	switch x := x.(type) {
	case boolVal:
		return bool(x)
	case *stringVal:
		return x.string()
	case int64Val:
		return int64(x)
	case intVal:
		return x.val
	case ratVal:
		return x.val
	case floatVal:
		return x.val
	default:
		return nil
	}
}

// Make returns the Value for x.
func Make(x any) Value {
	switch x := x.(type) {
	case bool:
		return boolVal(x)
	case string:
		return &stringVal{s: x}
	case int:
		return int64Val(x)
	case int64:
		return int64Val(x)
	case uint:
		return MakeUint64(uint64(x))
	case uint64:
		return MakeUint64(x)
	case float32:
		return makeFloat(newFloat().SetFloat64(float64(x)))
	case float64:
		return makeFloat(newFloat().SetFloat64(x))
	case *big.Int:
		return makeInt(x)
	case *big.Rat:
		return makeRat(x)
	case *big.Float:
		return makeFloat(x)
	default:
		return unknownVal{}
	}
}

// BitLen returns the number of bits required to represent
// the absolute value x in binary representation; x must be an Int or an Unknown.
// If x is Unknown, the result is 0.
func BitLen(x Value) int {
	switch x := x.(type) {
	case int64Val:
		return bits.Len64(uint64(abs(int64(x))))
	case intVal:
		return x.val.BitLen()
	case unknownVal:
		return 0
	default:
		panic(fmt.Sprintf("%v not an Int", x))
	}
}

// Sign returns -1, 0, or 1 depending on whether x < 0, x == 0, or x > 0;
// x must be numeric or Unknown. If x is Unknown, the result is 1.
func Sign(x Value) int {
	switch x := x.(type) {
	case int64Val:
		switch {
		case x < 0:
			return -1
		case x > 0:
			return 1
		}
		return 0
	case intVal:
		return x.val.Sign()
	case ratVal:
		return x.val.Sign()
	case floatVal:
		return x.val.Sign()
	case unknownVal:
		return 1 // avoid spurious division by zero errors
	default:
		panic(fmt.Sprintf("%v not numeric", x))
	}
}

// Bytes returns the bytes for the absolute value of x in little-
// endian binary representation; x must be an Int.
func Bytes(x Value) []byte {
	var t intVal
	switch x := x.(type) {
	case int64Val:
		t = i64toi(x)
	case intVal:
		t = x
	default:
		panic(fmt.Sprintf("%v not an Int", x))
	}

	words := t.val.Bits()
	bytes := make([]byte, len(words)*wordSize)

	i := 0
	for _, w := range words {
		for j := 0; j < wordSize; j++ {
			bytes[i] = byte(w)
			w >>= 8
			i++
		}
	}
	// remove leading 0's
	for i > 0 && bytes[i-1] == 0 {
		i--
	}

	return bytes[:i]
}

// MakeFromBytes returns the Int value given the bytes of its little-endian
// binary representation. An empty byte slice argument represents 0.
func MakeFromBytes(bytes []byte) Value {
	words := make([]big.Word, (len(bytes)+(wordSize-1))/wordSize)

	i := 0
	var w big.Word
	var s uint
	for _, b := range bytes {
		w |= big.Word(b) << s
		if s += 8; s == wordSize*8 {
			words[i] = w
			i++
			w = 0
			s = 0
		}
	}
	// store last word
	if i < len(words) {
		words[i] = w
		i++
	}
	// remove leading 0's
	for i > 0 && words[i-1] == 0 {
		i--
	}

	return makeInt(newInt().SetBits(words[:i]))
}

// Num returns the numerator of x; x must be Int, Float, or Unknown.
// If x is Unknown, or if it is too large or small to represent as a
// fraction, the result is Unknown. Otherwise the result is an Int
// with the same sign as x.
func Num(x Value) Value {
	switch x := x.(type) {
	case int64Val, intVal:
		return x
	case ratVal:
		return makeInt(x.val.Num())
	case floatVal:
		if smallFloat(x.val) {
			r, _ := x.val.Rat(nil)
			return makeInt(r.Num())
		}
	case unknownVal:
		break
	default:
		panic(fmt.Sprintf("%v not Int or Float", x))
	}
	return unknownVal{}
}

// Denom returns the denominator of x; x must be Int, Float, or Unknown.
// If x is Unknown, or if it is too large or small to represent as a
// fraction, the result is Unknown. Otherwise the result is an Int >= 1.
func Denom(x Value) Value {
	switch x := x.(type) {
	case int64Val, intVal:
		return int64Val(1)
	case ratVal:
		return makeInt(x.val.Denom())
	case floatVal:
		if smallFloat(x.val) {
			r, _ := x.val.Rat(nil)
			return makeInt(r.Denom())
		}
	case unknownVal:
		break
	default:
		panic(fmt.Sprintf("%v not Int or Float", x))
	}
	return unknownVal{}
}

// ToInt converts x to an Int value if x is representable as an Int.
// Otherwise it returns Unknown.
func ToInt(x Value) Value {
	switch x := x.(type) {
	case int64Val, intVal:
		return x

	case ratVal:
		if x.val.IsInt() {
			return makeInt(x.val.Num())
		}

	case floatVal:
		if smallFloat(x.val) {
			i := newInt()
			if _, acc := x.val.Int(i); acc == big.Exact {
				return makeInt(i)
			}

			const delta = 4 // a small number of bits > 0
			var t big.Float
			t.SetPrec(prec - delta)

			// try rounding down a little
			t.SetMode(big.ToZero)
			t.Set(x.val)
			if _, acc := t.Int(i); acc == big.Exact {
				return makeInt(i)
			}

			// try rounding up a little
			t.SetMode(big.AwayFromZero)
			t.Set(x.val)
			if _, acc := t.Int(i); acc == big.Exact {
				return makeInt(i)
			}
		}
	}

	return unknownVal{}
}

// ToFloat converts x to a Float value if x is representable as a Float.
// Otherwise it returns Unknown.
func ToFloat(x Value) Value {
	switch x := x.(type) {
	case int64Val:
		return i64tor(x) // x is always a small int
	case intVal:
		if smallInt(x.val) {
			return itor(x)
		}
		return itof(x)
	case ratVal, floatVal:
		return x
	}
	return unknownVal{}
}

// is32bit reports whether x can be represented using 32 bits.
func is32bit(x int64) bool {
	const s = 32
	return -1<<(s-1) <= x && x <= 1<<(s-1)-1
}

// is63bit reports whether x can be represented using 63 bits.
func is63bit(x int64) bool {
	const s = 63
	return -1<<(s-1) <= x && x <= 1<<(s-1)-1
}

// UnaryOp returns the result of the unary expression op y.
// The operation must be defined for the operand.
// If prec > 0 it specifies the ^ (xor) result size in bits.
// If y is Unknown, the result is Unknown.
func UnaryOp(op token.Token, y Value, prec uint) Value {
	switch op {
	case token.ADD:
		switch y.(type) {
		case unknownVal, int64Val, intVal, ratVal, floatVal:
			return y
		}

	case token.SUB:
		switch y := y.(type) {
		case unknownVal:
			return y
		case int64Val:
			if z := -y; z != y {
				return z // no overflow
			}
			return makeInt(newInt().Neg(big.NewInt(int64(y))))
		case intVal:
			return makeInt(newInt().Neg(y.val))
		case ratVal:
			return makeRat(newRat().Neg(y.val))
		case floatVal:
			return makeFloat(newFloat().Neg(y.val))
		}

	case token.XOR:
		z := newInt()
		switch y := y.(type) {
		case unknownVal:
			return y
		case int64Val:
			z.Not(big.NewInt(int64(y)))
		case intVal:
			z.Not(y.val)
		default:
			goto Error
		}
		// For unsigned types, the result will be negative and
		// thus "too large": We must limit the result precision
		// to the type's precision.
		if prec > 0 {
			z.AndNot(z, newInt().Lsh(big.NewInt(-1), prec)) // z &^= (-1)<<prec
		}
		return makeInt(z)

	case token.NOT:
		switch y := y.(type) {
		case unknownVal:
			return y
		case boolVal:
			return !y
		}
	}

Error:
	panic(fmt.Sprintf("invalid unary operation %s%v", op, y))
}

// BinaryOp returns the result of the binary expression x op y.
// The operation must be defined for the operands. If one of the
// operands is Unknown, the result is Unknown.
// BinaryOp doesn't handle comparisons or shifts; use Compare
// or Shift instead.
//
// To force integer division of Int operands, use op == token.QUO_ASSIGN
// instead of token.QUO; the result is guaranteed to be Int in this case.
// Division by zero leads to a run-time panic.
func BinaryOp(x_ Value, op token.Token, y_ Value) Value {
	if op == token.LOR {
		return Or(x_, y_)
	}
	x, y := match(x_, y_)

	switch x := x.(type) {
	case unknownVal:
		return x

	case boolVal:
		y := y.(boolVal)
		switch op {
		case token.LAND:
			return x && y
		case token.LOR:
			return x || y
		}

	case int64Val:
		a := int64(x)
		b := int64(y.(int64Val))
		var c int64
		switch op {
		case token.ADD:
			if !is63bit(a) || !is63bit(b) {
				return makeInt(newInt().Add(big.NewInt(a), big.NewInt(b)))
			}
			c = a + b
		case token.SUB:
			if !is63bit(a) || !is63bit(b) {
				return makeInt(newInt().Sub(big.NewInt(a), big.NewInt(b)))
			}
			c = a - b
		case token.MUL:
			if !is32bit(a) || !is32bit(b) {
				return makeInt(newInt().Mul(big.NewInt(a), big.NewInt(b)))
			}
			c = a * b
		case token.QUO:
			return makeInt(newInt().Quo(big.NewInt(a), big.NewInt(b)))
		case token.REM:
			return makeInt(newInt().Rem(big.NewInt(a), big.NewInt(b)))
		case token.AND:
			c = a & b
		case token.OR:
			c = a | b
		case token.XOR:
			c = a ^ b
		case token.AND_NOT:
			c = a &^ b
		default:
			goto Error
		}
		return int64Val(c)

	case intVal:
		a := x.val
		b := y.(intVal).val
		c := newInt()
		switch op {
		case token.ADD:
			c.Add(a, b)
		case token.SUB:
			c.Sub(a, b)
		case token.MUL:
			c.Mul(a, b)
		case token.QUO:
			c.Quo(a, b)
		case token.REM:
			c.Rem(a, b)
		case token.AND:
			c.And(a, b)
		case token.OR:
			c.Or(a, b)
		case token.XOR:
			c.Xor(a, b)
		case token.AND_NOT:
			c.AndNot(a, b)
		default:
			goto Error
		}
		return makeInt(c)

	case ratVal:
		a := x.val
		b := y.(ratVal).val
		c := newRat()
		switch op {
		case token.ADD:
			c.Add(a, b)
		case token.SUB:
			c.Sub(a, b)
		case token.MUL:
			c.Mul(a, b)
		case token.QUO:
			c.Quo(a, b)
		default:
			goto Error
		}
		return makeRat(c)

	case floatVal:
		a := x.val
		b := y.(floatVal).val
		c := newFloat()
		switch op {
		case token.ADD:
			c.Add(a, b)
		case token.SUB:
			c.Sub(a, b)
		case token.MUL:
			c.Mul(a, b)
		case token.QUO:
			c.Quo(a, b)
		default:
			goto Error
		}
		return makeFloat(c)

	case *stringVal:
		if op == token.ADD {
			return &stringVal{l: x, r: y.(*stringVal)}
		}
	}

Error:
	panic(fmt.Sprintf("invalid binary operation %v %s %v", x_, op, y_))
}

// Shift returns the result of the shift expression x op s
// with op == token.SHL or token.SHR (<< or >>). x must be
// an Int or an Unknown. If x is Unknown, the result is x.
func Shift(x Value, op token.Token, s uint) Value {
	switch x := x.(type) {
	case unknownVal:
		return x

	case int64Val:
		if s == 0 {
			return x
		}
		switch op {
		case token.SHL:
			z := i64toi(x).val
			return makeInt(z.Lsh(z, s))
		case token.SHR:
			return x >> s
		}

	case intVal:
		if s == 0 {
			return x
		}
		z := newInt()
		switch op {
		case token.SHL:
			return makeInt(z.Lsh(x.val, s))
		case token.SHR:
			return makeInt(z.Rsh(x.val, s))
		}
	}

	panic(fmt.Sprintf("invalid shift %v %s %d", x, op, s))
}

// Or returns the result of the value of x or y like in JavaScript.
//
// - If y is unknown, the result is x.
// - If x is false, the result is y.
// - If x is 0, the result is y.
// - If x is "", the result is y.
// - Otherwise, the result is x.
func Or(x, y Value) Value {
	if y.Kind() == Unknown {
		return x
	}
	switch x := x.(type) {
	case unknownVal:
		return y
	case boolVal:
		if !x {
			return y
		}
	case int64Val:
		if x == 0 {
			return y
		}
	case intVal:
		if x.val.Sign() == 0 {
			return y
		}
	case ratVal:
		if x.val.Sign() == 0 {
			return y
		}
	case floatVal:
		if x.val.Sign() == 0 {
			return y
		}
	case *stringVal:
		if x.string() == "" {
			return y
		}
	}
	return x
}

// Compare returns the result of the comparison x op y.
// The comparison must be defined for the operands.
// If one of the operands is Unknown, the result is false.
func Compare(x_ Value, op token.Token, y_ Value) bool {
	x, y := match(x_, y_)

	switch x := x.(type) {
	case unknownVal:
		return false

	case boolVal:
		y := y.(boolVal)
		switch op {
		case token.EQL:
			return x == y
		case token.NEQ:
			return x != y
		}

	case int64Val:
		y := y.(int64Val)
		switch op {
		case token.EQL:
			return x == y
		case token.NEQ:
			return x != y
		case token.LSS:
			return x < y
		case token.LEQ:
			return x <= y
		case token.GTR:
			return x > y
		case token.GEQ:
			return x >= y
		}

	case intVal:
		return cmpZero(x.val.Cmp(y.(intVal).val), op)

	case ratVal:
		return cmpZero(x.val.Cmp(y.(ratVal).val), op)

	case floatVal:
		return cmpZero(x.val.Cmp(y.(floatVal).val), op)

	case *stringVal:
		xs := x.string()
		ys := y.(*stringVal).string()
		switch op {
		case token.EQL:
			return xs == ys
		case token.NEQ:
			return xs != ys
		case token.LSS:
			return xs < ys
		case token.LEQ:
			return xs <= ys
		case token.GTR:
			return xs > ys
		case token.GEQ:
			return xs >= ys
		}
	}

	panic(fmt.Sprintf("invalid comparison %v %s %v", x_, op, y_))
}

// cmpZero returns the result of comparing x to 0.
func cmpZero(x int, op token.Token) bool {
	switch op {
	case token.EQL:
		return x == 0
	case token.NEQ:
		return x != 0
	case token.LSS:
		return x < 0
	case token.LEQ:
		return x <= 0
	case token.GTR:
		return x > 0
	case token.GEQ:
		return x >= 0
	}
	panic(fmt.Sprintf("invalid comparison %v %s 0", x, op))
}

// match returns the matching representation (same type) with the
// smallest complexity for two values x and y. If one of them is
// numeric, both of them must be numeric. If one of them is Unknown
// or invalid (say, nil) both results are that value.
func match(x, y Value) (_, _ Value) {
	switch ox, oy := ord(x), ord(y); {
	case ox < oy:
		x, y = match0(x, y)
	case ox > oy:
		y, x = match0(y, x)
	}
	return x, y
}

// match0 must only be called by match.
// Invariant: ord(x) < ord(y)
func match0(x, y Value) (_, _ Value) {
	switch y.(type) {
	case intVal:
		switch x1 := x.(type) {
		case int64Val:
			return i64toi(x1), y
		}
	case ratVal:
		switch x1 := x.(type) {
		case int64Val:
			return i64tor(x1), y
		case intVal:
			return itor(x1), y
		}
	case floatVal:
		switch x1 := x.(type) {
		case int64Val:
			return i64tof(x1), y
		case intVal:
			return itof(x1), y
		case ratVal:
			return rtof(x1), y
		}
	}

	return x, x
}

// ord returns the order of the given value.
func ord(x Value) int {
	switch x.(type) {
	default:
		return -1
	case unknownVal:
		return 0
	case boolVal, *stringVal:
		return 1
	case int64Val:
		return 2
	case intVal:
		return 3
	case ratVal:
		return 4
	case floatVal:
		return 5
	}
}

// Compute the size of a Word in bytes.
const (
	_m       = ^big.Word(0)
	_log     = _m>>8&1 + _m>>16&1 + _m>>32&1
	wordSize = 1 << _log
)

// abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// Underlying returns the underlying value of a Value, if it is a simple type.:w
func Underlying(value Value) any {
	if value == nil {
		return nil
	}
	switch value.Kind() {
	case String:
		return StringVal(value)
	case Int:
		if i, exactly := Int64Val(value); exactly {
			return i
		}
		u, _ := Uint64Val(value)
		return u
	case Float:
		if f, exactly := Float32Val(value); exactly {
			return f
		}
		f, _ := Float64Val(value)
		return f
	case Bool:
		return BoolVal(value)
	default:
		return nil
	}
}
