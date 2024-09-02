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
)

// Funcs is a map of utility functions for use in templates
var Funcs = map[string]any{
	// _ is a no-op function that returns an empty string.
	// It's useful to place a newline in the template.
	"_": func() string { return "" },

	// map maps a list of values using the given converter.
	"map": mapFunc,

	// String functions

	"quote":       convs(noError(strconv.Quote)),
	"unquote":     convs(strconv.Unquote),
	"capitalize":  convs(noError(capitalize)),
	"lower":       convs(noError(strings.ToLower)),
	"upper":       convs(noError(strings.ToUpper)),
	"replace":     replace,
	"trim":        convs(noError(strings.TrimSpace)),
	"trimPrefix":  conv2s(trimPrefix),
	"hasPrefix":   hasPrefix,
	"trimSuffix":  conv2s(trimSuffix),
	"hasSuffix":   hasSuffix,
	"split":       conv2(split),
	"join":        conv2(join),
	"striptags":   convs(striptags),
	"substr":      conv3(substr),
	"repeat":      conv2(repeat),
	"camelCase":   convs(noError(camelCase)),
	"pascalCase":  convs(noError(pascalCase)),
	"snakeCase":   convs(noError(snakeCase)),
	"kebabCase":   convs(noError(kebabCase)),
	"truncate":    conv3(truncate),
	"wordwrap":    conv2(wordwrap),
	"center":      conv2(center),
	"matchRegex":  matchRegex,
	"html":        convs(noError(html.EscapeString)),
	"urlquery":    convs(noError(url.QueryEscape)),
	"urlUnescape": convs(url.QueryUnescape),

	// Encoding functions

	"b64enc": convs(noError(b64enc)),
	"b64dec": convs(b64dec),

	// List functions

	"list":     list,
	"first":    conv(first),
	"last":     conv(last),
	"reverse":  conv(reverse),
	"sort":     conv(sortStrings),
	"uniq":     conv(uniq),
	"includes": includes,

	// Math functions

	"add":   conv2(add),
	"sub":   conv2(sub),
	"mul":   conv2(mul),
	"quo":   conv2(quo),
	"rem":   conv2(rem),
	"mod":   conv2(mod),
	"ceil":  conv(ceil),
	"floor": conv(floor),
	"round": conv2(round),
	"min":   min,
	"max":   max,

	// Type conversion functions

	"int":    conv(toInt64),
	"float":  conv(toFloat64),
	"string": conv(toString),
	"bool":   conv(toBool),

	// Date functions

	"now":       time.Now,
	"parseTime": parseTime,

	// Conditional functions

	"ternary": ternary,
}

// String functions

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

func replace(old, new string, args ...reflect.Value) (reflect.Value, error) {
	switch len(args) {
	case 0:
		return convs(noError(func(s string) string {
			return strings.Replace(s, old, new, -1)
		})).value(), nil
	case 1:
		if s, ok := asString(args[0]); ok {
			return reflect.ValueOf(strings.Replace(s, old, new, -1)), nil
		} else if n, err := toInt64(args[0]); err == nil {
			return convs(noError(func(s string) string {
				return strings.Replace(s, old, new, int(n.Int()))
			})).value(), nil
		} else if c, err := asConverterFunc(args[0]); err != nil {
			return reflect.Value{}, err
		} else if c != nil {
			return reflect.ValueOf(c.then(func(s string) (string, error) {
				return strings.Replace(s, old, new, -1), nil
			})), nil
		} else {
			return reflect.Value{}, fmt.Errorf("replace: expected string, int or ConverterFunc, got %s", args[0].Type())
		}
	case 2:
		n, err := toInt64(args[0])
		if err != nil {
			return reflect.Value{}, err
		}
		s, ok := asString(args[1])
		if ok {
			return reflect.ValueOf(strings.Replace(s, old, new, int(n.Int()))), nil
		}
		if c, err := asConverterFunc(args[1]); err != nil {
			return reflect.Value{}, err
		} else if c != nil {
			return reflect.ValueOf(c.then(func(s string) (string, error) {
				return strings.Replace(s, old, new, int(n.Int())), nil
			})), nil
		} else {
			return reflect.Value{}, fmt.Errorf("replace: expected string or ConverterFunc, got %s", args[1].Type())
		}
	default:
		return reflect.Value{}, fmt.Errorf("replace: expected 0, 1 or 2 arguments, got %d", len(args))
	}
}

func trimSpace(v reflect.Value) (string, error) {
	s, ok := asString(v)
	if !ok {
		return "", fmt.Errorf("trim: expected string, got %s", v.Type())
	}
	return strings.TrimSpace(s), nil
}

func trimPrefix(prefix, v string) (string, error) {
	return strings.TrimPrefix(v, prefix), nil
}

func hasPrefix(prefix, v reflect.Value) (bool, error) {
	sp, ok := asString(prefix)
	if !ok {
		return false, fmt.Errorf("hasPrefix: expected string as first argument, got %s", prefix.Type())
	}
	sv, ok := asString(v)
	if !ok {
		return false, fmt.Errorf("hasPrefix: expected string as second argument, got %s", v.Type())
	}
	return strings.HasPrefix(sv, sp), nil
}

func trimSuffix(suffix, v string) (string, error) {
	return strings.TrimSuffix(v, suffix), nil
}

func hasSuffix(suffix, v reflect.Value) (bool, error) {
	ss, ok := asString(suffix)
	if !ok {
		return false, fmt.Errorf("hasSuffix: expected string as first argument, got %s", suffix.Type())
	}
	sv, ok := asString(v)
	if !ok {
		return false, fmt.Errorf("hasSuffix: expected string as second argument, got %s", v.Type())
	}
	return strings.HasSuffix(sv, ss), nil
}

func split(sep string, v reflect.Value) (reflect.Value, error) {
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("split: expected string as second argument, got %s", v.Type())
	}
	return reflect.ValueOf(strings.Split(s, sep)), nil
}

func join(sep string, v reflect.Value) (reflect.Value, error) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return reflect.Value{}, fmt.Errorf("join: expected slice or array as second argument, got %s", v.Type())
	}

	length := v.Len()
	parts := make([]string, length)

	for i := 0; i < length; i++ {
		parts[i] = fmt.Sprint(v.Index(i).Interface())
	}

	return reflect.ValueOf(strings.Join(parts, sep)), nil
}

func striptags(s string) (string, error) {
	return regexp.MustCompile("<[^>]*>").ReplaceAllString(s, ""), nil
}

func substr(start, length int, v reflect.Value) (reflect.Value, error) {
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("substr: expected string as third argument, got %s", v.Type())
	}
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
	if start > end {
		start = end
	}
	return reflect.ValueOf(s[start:end]), nil
}

func repeat(count int, v reflect.Value) (reflect.Value, error) {
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("repeat: expected string as second argument, got %s", v.Type())
	}
	if count <= 0 {
		return reflect.ValueOf(""), nil
	}
	return reflect.ValueOf(strings.Repeat(s, count)), nil
}

func camelCase(s string) string {
	var result strings.Builder
	capNext := false
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if i == 0 {
				result.WriteRune(unicode.ToLower(r))
			} else if capNext {
				result.WriteRune(unicode.ToUpper(r))
				capNext = false
			} else {
				result.WriteRune(r)
			}
		} else {
			capNext = true
		}
	}
	return result.String()
}

func pascalCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	capNext := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if capNext {
				result.WriteRune(unicode.ToUpper(r))
				capNext = false
			} else {
				result.WriteRune(r)
			}
		} else {
			capNext = true
		}
	}
	return result.String()
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

func truncate(length int, suffix, v reflect.Value) (reflect.Value, error) {
	ss, ok := asString(suffix)
	if !ok {
		return reflect.Value{}, fmt.Errorf("truncate: expected string as first argument, got %s", suffix.Type())
	}
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("truncate: expected string as second argument, got %s", v.Type())
	}
	if length <= 0 {
		return reflect.ValueOf(""), nil
	}
	if len(s) <= length {
		return reflect.ValueOf(s), nil
	}
	return reflect.ValueOf(s[:length-len(ss)] + ss), nil
}

func wordwrap(width int, v reflect.Value) (reflect.Value, error) {
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("wordwrap: expected string, got %s", v.Type())
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return reflect.ValueOf(s), nil
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
	return reflect.ValueOf(strings.Join(lines, "\n")), nil
}

func center(width int, v reflect.Value) (reflect.Value, error) {
	s, ok := asString(v)
	if !ok {
		return reflect.Value{}, fmt.Errorf("center: expected string, got %s", v.Type())
	}
	if width <= len(s) {
		return reflect.ValueOf(s), nil
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return reflect.ValueOf(strings.Repeat(" ", left) + s + strings.Repeat(" ", right)), nil
}

func matchRegex(pattern, v reflect.Value) (bool, error) {
	p, ok := asString(pattern)
	if !ok {
		return false, fmt.Errorf("matchRegex: expected string as first argument, got %s", pattern.Type())
	}
	s, ok := asString(v)
	if !ok {
		return false, fmt.Errorf("matchRegex: expected string as second argument, got %s", v.Type())
	}
	return regexp.MatchString(p, s)
}

// Encoding functions

func b64enc(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func b64dec(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// List functions

func list(values ...reflect.Value) (reflect.Value, error) {
	if len(values) == 0 {
		return reflect.ValueOf([]string{}), nil
	}
	result := reflect.MakeSlice(reflect.SliceOf(values[0].Type()), len(values), len(values))
	for i, v := range values {
		result.Index(i).Set(v)
	}
	return result, nil
}

func first(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v.Index(0), nil
	default:
		if s, ok := asString(v); ok {
			if len(s) == 0 {
				return reflect.Value{}, nil
			}
			return reflect.ValueOf(s[0]), nil
		}
		return reflect.Value{}, fmt.Errorf("first: unsupported type %s", v.Type())
	}
}

func last(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v.Index(v.Len() - 1), nil
	default:
		if s, ok := asString(v); ok {
			if len(s) == 0 {
				return reflect.Value{}, nil
			}
			return reflect.ValueOf(s[len(s)-1]), nil
		}
		return reflect.Value{}, fmt.Errorf("last: unsupported type %s", v.Type())
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reverse(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		length := v.Len()
		reversed := reflect.MakeSlice(v.Type(), length, length)
		for i := 0; i < length; i++ {
			reversed.Index(i).Set(v.Index(length - 1 - i))
		}
		return reversed, nil
	default:
		if s, ok := asString(v); ok {
			return reflect.ValueOf(reverseString(s)), nil
		}
		if c, err := asConverterFunc(v); err != nil {
			return reflect.Value{}, err
		} else if c != nil {
			return reflect.ValueOf(c.then(func(s string) (string, error) {
				return reverseString(s), nil
			})), nil
		}
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

func ContainsWord(s, word string) bool {
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(word))
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

func includes(item, collection reflect.Value) (bool, error) {
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
	default:
		if s, ok := asString(collection); ok {
			if i, ok := asString(item); ok {
				return ContainsWord(s, i), nil
			}
		}
		return false, fmt.Errorf("includes: unsupported collection type %s", collection.Type())
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
	f, err := toFloat64(x)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(math.Ceil(f.Float())), nil
}

func floor(x reflect.Value) (reflect.Value, error) {
	f, err := toFloat64(x)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(math.Floor(f.Float())), nil
}

func round(precision int, x reflect.Value) (reflect.Value, error) {
	f, err := toFloat64(x)
	if err != nil {
		return reflect.Value{}, err
	}
	shift := math.Pow10(precision)
	return reflect.ValueOf(math.Round(f.Float()*shift) / shift), nil
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

// toStrings converts a slice or array of reflect.Value to a slice of strings.
func toStrings(v reflect.Value) ([]string, error) {
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("toStrings: expected slice or array, got %s", v.Type())
	}
	if v.Len() == 0 {
		return nil, nil
	}
	s := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		if str, ok := asString(v.Index(i)); ok {
			s[i] = str
		} else {
			return nil, fmt.Errorf("toStrings: expected string, got %s", v.Index(i).Type())
		}
	}
	return s, nil
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

func toInt64(v reflect.Value) (reflect.Value, error) {
	if v.Kind() == reflect.Int64 {
		return v, nil
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(int64(v.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(int64(v.Float())), nil
	case reflect.String:
		i, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i), nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf(int64(1)), nil
		}
		return reflect.ValueOf(int64(0)), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %s to int", v.Type())
	}
}

func toFloat64(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(float64(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(float64(v.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(v.Float()), nil
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(f), nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf(float64(1)), nil
		}
		return reflect.ValueOf(float64(0)), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %s to float", v.Type())
	}
}

func toString(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.String:
		return v, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Float32:
		return reflect.ValueOf(strconv.FormatFloat(v.Float(), 'f', -1, 32)), nil
	case reflect.Float64:
		return reflect.ValueOf(strconv.FormatFloat(v.Float(), 'f', -1, 64)), nil
	case reflect.Bool:
		return reflect.ValueOf(strconv.FormatBool(v.Bool())), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %s to string", v.Type())
	}
}

func toBool(v reflect.Value) (reflect.Value, error) {
	if v.Kind() == reflect.Bool {
		return v, nil
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(v.Int() != 0), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(v.Uint() != 0), nil
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		return reflect.ValueOf(f != 0 && !math.IsNaN(f)), nil
	case reflect.String:
		x, err := strconv.ParseBool(v.String())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(x), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %s to bool", v.Type())
	}
}

func asString(v reflect.Value) (string, bool) {
	if v.Kind() == reflect.String {
		return v.String(), true
	}
	// If v implements fmt.Stringer, use its String method
	if v.CanInterface() {
		if s, ok := v.Interface().(fmt.Stringer); ok {
			return s.String(), true
		}
	}
	return "", false
}

// Date functions

func timeFormat(t time.Time, layout string) string {
	return t.Format(layout)
}

func parseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// Conditional functions

func ternary(cond, trueVal, falseVal reflect.Value) (reflect.Value, error) {
	if cond.Kind() != reflect.Bool {
		return reflect.Value{}, fmt.Errorf("ternary: condition must be bool, got %s", cond.Type())
	}
	if cond.Bool() {
		return trueVal, nil
	}
	return falseVal, nil
}
