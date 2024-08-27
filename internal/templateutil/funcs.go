package templateutil

import (
	"encoding/base64"
	"fmt"
	"html"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Funcs is a map of utility functions for use in templates
var Funcs = map[string]any{
	// String functions
	"capitalize": capitalize,
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"replace":    strings.Replace,
	"trim":       strings.TrimSpace,
	"trimPrefix": trimPrefix,
	"trimSuffix": trimSuffix,
	"split":      split,
	"join":       join,
	"striptags":  striptags,
	"substr":     substr,
	"repeat":     strings.Repeat,
	"camelCase":  camelCase,
	"snakeCase":  snakeCase,
	"kebabCase":  kebabCase,
	"truncate":   truncate,
	"wordwrap":   wordwrap,
	"center":     center,
	"matchRegex": matchRegex,

	// Escaping functions
	"html":        html.EscapeString,
	"urlquery":    url.QueryEscape,
	"urlUnescape": urlUnescape,

	// Encoding functions
	"b64enc": base64.StdEncoding.EncodeToString,
	"b64dec": b64dec,

	// List functions
	"first":   first,
	"last":    last,
	"rest":    rest,
	"reverse": reverse,
	"sort":    sortStrings,
	"uniq":    uniq,
	"in":      in,

	// Math functions
	"add":   add,
	"sub":   sub,
	"mul":   mul,
	"quo":   quo,
	"rem":   rem,
	"mod":   mod,
	"max":   max,
	"min":   min,
	"ceil":  ceil,
	"floor": floor,
	"round": round,

	// Type conversion functions
	"int":    toInt,
	"float":  toFloat,
	"string": toString,
	"bool":   toBool,

	// Date functions
	"now":    time.Now,
	"format": timeFormat,
	"parse":  timeParse,

	// Conditional functions
	"default":  dfault,
	"ternary":  ternary,
	"coalesce": coalesce,
}

// String functions

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func trimPrefix(prefix, s string) string {
	return strings.TrimPrefix(s, prefix)
}

func trimSuffix(suffix, s string) string {
	return strings.TrimSuffix(s, suffix)
}

func split(sep, s string) []string {
	return strings.Split(s, sep)
}

func join(sep string, v reflect.Value) (string, error) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return "", fmt.Errorf("join: unsupported type %s", v.Type())
	}

	length := v.Len()
	parts := make([]string, length)

	for i := 0; i < length; i++ {
		parts[i] = fmt.Sprint(v.Index(i).Interface())
	}

	return strings.Join(parts, sep), nil
}

func striptags(s string) string {
	return regexp.MustCompile("<[^>]*>").ReplaceAllString(s, "")
}

func substr(start, length int, s string) string {
	if start < 0 {
		start = 0
	}
	if length < 0 {
		length = 0
	}
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

func camelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = capitalize(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

func snakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsNumber(r) && !unicode.IsNumber(rune(s[i-1]))) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func kebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsNumber(r) && !unicode.IsNumber(rune(s[i-1]))) {
			result.WriteRune('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func truncate(s string, length int, suffix string) string {
	if length <= 0 {
		return ""
	}
	if len(s) <= length {
		return s
	}
	return s[:length-len(suffix)] + suffix
}

func wordwrap(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var lines []string
	var currentLine string
	for _, word := range words {
		if len(currentLine)+len(word) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return strings.Join(lines, "\n")
}

func center(s string, width int) string {
	if width <= len(s) {
		return s
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func matchRegex(pattern, s string) (bool, error) {
	return regexp.MatchString(pattern, s)
}

func urlUnescape(s string) (string, error) {
	return url.QueryUnescape(s)
}

// Encoding functions

func b64dec(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// List functions

func first(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v.Index(0), nil
	default:
		return reflect.Value{}, fmt.Errorf("first: unsupported type %s", v.Type())
	}
}

func last(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v.Index(v.Len() - 1), nil
	default:
		return reflect.Value{}, fmt.Errorf("last: unsupported type %s", v.Type())
	}
}

func rest(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if v.Len() <= 1 {
			return reflect.MakeSlice(v.Type(), 0, 0), nil
		}
		return v.Slice(1, v.Len()), nil
	default:
		return reflect.Value{}, fmt.Errorf("rest: unsupported type %s", v.Type())
	}
}

func reverse(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.String:
		runes := []rune(v.String())
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return reflect.ValueOf(string(runes)), nil
	case reflect.Slice, reflect.Array:
		length := v.Len()
		reversed := reflect.MakeSlice(v.Type(), length, length)
		for i := 0; i < length; i++ {
			reversed.Index(i).Set(v.Index(length - 1 - i))
		}
		return reversed, nil
	default:
		return reflect.Value{}, fmt.Errorf("reverse: unsupported type %s", v.Type())
	}
}

func sortStrings(v reflect.Value) (reflect.Value, error) {
	if v.Kind() != reflect.Slice || v.Type().Elem().Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("sortStrings: expected []string, got %s", v.Type())
	}
	sorted := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		sorted[i] = v.Index(i).String()
	}
	sort.Strings(sorted)
	return reflect.ValueOf(sorted), nil
}

func uniq(v reflect.Value) (reflect.Value, error) {
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return reflect.Value{}, fmt.Errorf("uniq: unsupported type %s", v.Type())
	}

	length := v.Len()
	seen := make(map[any]bool)
	uniqueSlice := reflect.MakeSlice(v.Type(), 0, length)

	for i := 0; i < length; i++ {
		elem := v.Index(i)
		if !seen[elem.Interface()] {
			seen[elem.Interface()] = true
			uniqueSlice = reflect.Append(uniqueSlice, elem)
		}
	}

	return uniqueSlice, nil
}

func in(item, collection reflect.Value) (bool, error) {
	switch collection.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < collection.Len(); i++ {
			if reflect.DeepEqual(item.Interface(), collection.Index(i).Interface()) {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		return collection.MapIndex(item).IsValid(), nil
	case reflect.String:
		return strings.Contains(collection.String(), fmt.Sprint(item.Interface())), nil
	default:
		return false, fmt.Errorf("in: unsupported collection type %s", collection.Type())
	}
}

// Math functions

func add(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		return new(big.Float).Add(x, y)
	})
}

func sub(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		return new(big.Float).Sub(x, y)
	})
}

func mul(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		return new(big.Float).Mul(x, y)
	})
}

func quo(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		if y.Sign() == 0 {
			return nil // Division by zero
		}
		return new(big.Float).Quo(x, y)
	})
}

func rem(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		if y.Sign() == 0 {
			return nil // Division by zero
		}
		return new(big.Float).Quo(x, y).SetMode(big.ToZero)
	})
}

func mod(a, b reflect.Value) (reflect.Value, error) {
	return numericBinaryOp(a, b, func(x, y *big.Float) *big.Float {
		if y.Sign() == 0 {
			return nil // Division by zero
		}
		q := new(big.Float).Quo(x, y)
		q.SetMode(big.ToZero)
		q.SetPrec(0)
		return new(big.Float).Sub(x, new(big.Float).Mul(q, y))
	})
}

func max(args ...reflect.Value) (reflect.Value, error) {
	if len(args) == 0 {
		return reflect.Value{}, fmt.Errorf("max: at least one argument is required")
	}

	maxVal := args[0]
	for _, arg := range args[1:] {
		result, err := numericCompare(maxVal, arg)
		if err != nil {
			return reflect.Value{}, err
		}
		if result < 0 {
			maxVal = arg
		}
	}

	return maxVal, nil
}

func min(args ...reflect.Value) (reflect.Value, error) {
	if len(args) == 0 {
		return reflect.Value{}, fmt.Errorf("min: at least one argument is required")
	}

	minVal := args[0]
	for _, arg := range args[1:] {
		result, err := numericCompare(minVal, arg)
		if err != nil {
			return reflect.Value{}, err
		}
		if result > 0 {
			minVal = arg
		}
	}

	return minVal, nil
}

func ceil(x reflect.Value) (reflect.Value, error) {
	f, err := toFloat(x)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(math.Ceil(f)), nil
}

func floor(x reflect.Value) (reflect.Value, error) {
	f, err := toFloat(x)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(math.Floor(f)), nil
}

func round(x reflect.Value, precision int) (reflect.Value, error) {
	f, err := toFloat(x)
	if err != nil {
		return reflect.Value{}, err
	}
	shift := math.Pow10(precision)
	return reflect.ValueOf(math.Round(f*shift) / shift), nil
}

// Helper functions for numeric operations

func numericBinaryOp(a, b reflect.Value, op func(*big.Float, *big.Float) *big.Float) (reflect.Value, error) {
	x, err := toBigFloat(a)
	if err != nil {
		return reflect.Value{}, err
	}
	y, err := toBigFloat(b)
	if err != nil {
		return reflect.Value{}, err
	}

	result := op(x, y)
	if result == nil {
		return reflect.Value{}, fmt.Errorf("operation error (possibly division by zero)")
	}

	// Try to convert back to original type if possible
	switch {
	case isInt(a) && isInt(b):
		if i, acc := result.Int64(); acc == big.Exact {
			return reflect.ValueOf(i), nil
		}
	case isUint(a) && isUint(b):
		if u, acc := result.Uint64(); acc == big.Exact {
			return reflect.ValueOf(u), nil
		}
	case isFloat(a) || isFloat(b):
		f, _ := result.Float64()
		return reflect.ValueOf(f), nil
	}

	// If conversion is not possible, return as big.Float
	return reflect.ValueOf(result), nil
}

func numericCompare(a, b reflect.Value) (int, error) {
	x, err := toBigFloat(a)
	if err != nil {
		return 0, err
	}
	y, err := toBigFloat(b)
	if err != nil {
		return 0, err
	}
	return x.Cmp(y), nil
}

func toBigFloat(v reflect.Value) (*big.Float, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(big.Float).SetInt64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(big.Float).SetUint64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return new(big.Float).SetFloat64(v.Float()), nil
	default:
		return nil, fmt.Errorf("unsupported type for numeric operation: %s", v.Type())
	}
}

// Type checking functions

func isInt(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUint(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloat(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// Type conversion functions

func toInt(v reflect.Value) (int64, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.String:
		return strconv.ParseInt(v.String(), 10, 64)
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %s to int", v.Type())
	}
}

func toFloat(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(v.String(), 64)
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %s to float", v.Type())
	}
}

func toString(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return "", fmt.Errorf("cannot convert %s to string", v.Type())
	}
}

func toBool(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0, nil
	case reflect.String:
		return strconv.ParseBool(v.String())
	default:
		return false, fmt.Errorf("cannot convert %s to bool", v.Type())
	}
}

// Date functions

func timeFormat(t time.Time, layout string) string {
	return t.Format(layout)
}

func timeParse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// Conditional functions

func dfault(def, val reflect.Value) (reflect.Value, error) {
	if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return def, nil
	}
	return val, nil
}

func ternary(cond, trueVal, falseVal reflect.Value) (reflect.Value, error) {
	if cond.Kind() != reflect.Bool {
		return reflect.Value{}, fmt.Errorf("ternary: condition must be bool, got %s", cond.Type())
	}
	if cond.Bool() {
		return trueVal, nil
	}
	return falseVal, nil
}

func coalesce(args ...reflect.Value) (reflect.Value, error) {
	for _, arg := range args {
		if arg.IsValid() && !arg.IsZero() {
			return arg, nil
		}
	}
	return reflect.Value{}, nil
}

// Additional utility functions

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}
