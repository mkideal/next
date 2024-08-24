package templateutil

import (
	"encoding/base64"
	"fmt"
	"html"
	"math"
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
var Funcs = map[string]interface{}{
	// String functions
	"capitalize": capitalize, // Capitalizes the first character of a string
	"lower":      lower,      // Converts a string to lowercase
	"upper":      upper,      // Converts a string to uppercase
	"title":      title,      // Converts the first character of each word to uppercase
	"replace":    replace,    // Replaces occurrences of a substring
	"trim":       trim,       // Removes leading and trailing whitespace
	"trimPrefix": trimPrefix, // Removes specified prefix from a string
	"trimSuffix": trimSuffix, // Removes specified suffix from a string
	"striptags":  striptags,  // Removes HTML tags from a string
	"escape":     escape,     // Escapes HTML special characters
	"unescape":   unescape,   // Unescapes HTML special characters
	"safe":       safe,       // Marks a string as safe HTML (no escaping)
	"truncate":   truncate,   // Truncates a string to a specified length
	"wordwrap":   wordwrap,   // Wraps words at specified line length
	"center":     center,     // Centers the string within a specified width
	"format":     format,     // Formats a string according to a format specifier
	"substr":     substr,     // Returns a substring

	// List functions
	"join":    join,        // Joins a list of strings with a separator
	"split":   split,       // Splits a string into a list of substrings
	"length":  length,      // Returns the length of a string, slice, or map
	"first":   first,       // Returns the first element of a list or string
	"last":    last,        // Returns the last element of a list or string
	"reverse": reverse,     // Reverses a string or list
	"sort":    sortStrings, // Sorts a list of strings
	"unique":  unique,      // Returns a list of unique elements
	"in":      in,          // Checks if an element is in a list or string

	// Math functions
	"abs":   abs,   // Returns the absolute value of a number
	"ceil":  ceil,  // Returns the ceiling of a number
	"floor": floor, // Returns the floor of a number
	"round": round, // Rounds a number to a specified precision
	"max":   max,   // Returns the maximum of two or more numbers
	"min":   min,   // Returns the minimum of two or more numbers
	"sum":   sum,   // Returns the sum of a list of numbers
	"add":   add,   // Adds two numbers
	"sub":   sub,   // Subtracts two numbers
	"mul":   mul,   // Multiplies two numbers
	"quo":   quo,   // Divides two numbers (quotient)
	"mod":   mod,   // Returns the remainder of division
	"pow":   pow,   // Raises a number to a power
	"sqrt":  sqrt,  // Returns the square root of a number

	// Encoding functions
	"b64encode": b64encode, // Base64 encodes a string
	"b64decode": b64decode, // Base64 decodes a string
	"urlencode": urlencode, // URL-encodes a string
	"urldecode": urldecode, // URL-decodes a string

	// Date functions
	"now":            now,            // Returns the current time
	"datetimeformat": datetimeformat, // Formats a date according to the specified format
	"dateAdd":        dateAdd,        // Adds a duration to a date
	"dateSub":        dateSub,        // Subtracts a duration from a date

	// Logic functions
	"default":  defaultValue, // Returns a default value if the input is nil
	"coalesce": coalesce,     // Returns the first non-nil value
	"ternary":  ternary,      // Conditional (if-then-else) logic

	// Type conversion functions
	"toString": toString, // Converts a value to a string
	"toInt":    toInt,    // Converts a value to an int
	"toFloat":  toFloat,  // Converts a value to a float64
	"toBool":   toBool,   // Converts a value to a bool

	// Additional useful functions
	"contains":   contains,   // Checks if a string contains a substring or a slice contains an element
	"hasPrefix":  hasPrefix,  // Checks if a string starts with a prefix
	"hasSuffix":  hasSuffix,  // Checks if a string ends with a suffix
	"repeat":     repeat,     // Repeats a string a specified number of times
	"matchRegex": matchRegex, // Checks if a string matches a regular expression
	"camelCase":  camelCase,  // Converts a string to camelCase
	"snakeCase":  snakeCase,  // Converts a string to snake_case
	"kebabCase":  kebabCase,  // Converts a string to kebab-case
}

// String functions

func capitalize(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("capitalize function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("capitalize function requires a string argument")
	}
	if s == "" {
		return "", nil
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:], nil
}

func lower(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("lower function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("lower function requires a string argument")
	}
	return strings.ToLower(s), nil
}

func upper(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("upper function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("upper function requires a string argument")
	}
	return strings.ToUpper(s), nil
}

func title(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("title function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("title function requires a string argument")
	}
	return strings.Title(s), nil
}

func replace(args ...interface{}) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("replace function requires exactly four arguments")
	}
	s, ok1 := args[0].(string)
	old, ok2 := args[1].(string)
	new, ok3 := args[2].(string)
	n, ok4 := args[3].(int)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return "", fmt.Errorf("replace function requires (string, string, string, int) arguments")
	}
	return strings.Replace(s, old, new, n), nil
}

func trim(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("trim function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("trim function requires a string argument")
	}
	return strings.TrimSpace(s), nil
}

func trimPrefix(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("trimPrefix function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	prefix, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("trimPrefix function requires (string, string) arguments")
	}
	return strings.TrimPrefix(s, prefix), nil
}

func trimSuffix(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("trimSuffix function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	suffix, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("trimSuffix function requires (string, string) arguments")
	}
	return strings.TrimSuffix(s, suffix), nil
}

func striptags(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("striptags function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("striptags function requires a string argument")
	}
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(s, ""), nil
}

func escape(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("escape function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("escape function requires a string argument")
	}
	return html.EscapeString(s), nil
}

func unescape(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("unescape function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("unescape function requires a string argument")
	}
	return html.UnescapeString(s), nil
}

func safe(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("safe function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("safe function requires a string argument")
	}
	return s, nil
}

func truncate(args ...interface{}) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("truncate function requires exactly three arguments")
	}
	s, ok1 := args[0].(string)
	length, ok2 := args[1].(int)
	end, ok3 := args[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return "", fmt.Errorf("truncate function requires (string, int, string) arguments")
	}
	if len(s) <= length {
		return s, nil
	}
	return s[:length-len(end)] + end, nil
}

func wordwrap(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("wordwrap function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	width, ok2 := args[1].(int)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("wordwrap function requires (string, int) arguments")
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return "", nil
	}
	var lines []string
	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) <= width {
			line += " " + word
		} else {
			lines = append(lines, line)
			line = word
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n"), nil
}

func center(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("center function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	width, ok2 := args[1].(int)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("center function requires (string, int) arguments")
	}
	if width <= len(s) {
		return s, nil
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right), nil
}

func format(args ...interface{}) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("format function requires at least one argument")
	}
	format, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("format function requires a string as the first argument")
	}
	return fmt.Sprintf(format, args[1:]...), nil
}

func substr(args ...interface{}) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("substr function requires exactly three arguments")
	}
	s, ok1 := args[0].(string)
	start, ok2 := args[1].(int)
	length, ok3 := args[2].(int)
	if !ok1 || !ok2 || !ok3 {
		return "", fmt.Errorf("substr function requires (string, int, int) arguments")
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
	return s[start:end], nil
}

// List functions

func join(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("join function requires exactly two arguments")
	}
	elems, ok1 := args[0].([]string)
	sep, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("join function requires ([]string, string) arguments")
	}
	return strings.Join(elems, sep), nil
}

func split(args ...interface{}) ([]string, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("split function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	sep, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("split function requires (string, string) arguments")
	}
	return strings.Split(s, sep), nil
}

func length(args ...interface{}) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("length function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case string:
		return utf8.RuneCountInString(v), nil
	case []interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		return 0, fmt.Errorf("length function requires a string, slice, or map argument")
	}
}

func first(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case string:
		if len(v) == 0 {
			return "", nil
		}
		return string(v[0]), nil
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		return v[0], nil
	default:
		return nil, fmt.Errorf("first function requires a string or slice argument")
	}
}

func last(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("last function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case string:
		if len(v) == 0 {
			return "", nil
		}
		return string(v[len(v)-1]), nil
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		return v[len(v)-1], nil
	default:
		return nil, fmt.Errorf("last function requires a string or slice argument")
	}
}

func reverse(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("reverse function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case string:
		runes := []rune(v)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case []interface{}:
		reversed := make([]interface{}, len(v))
		for i, j := 0, len(v)-1; i < len(v); i, j = i+1, j-1 {
			reversed[i] = v[j]
		}
		return reversed, nil
	default:
		return nil, fmt.Errorf("reverse function requires a string or slice argument")
	}
}

func sortStrings(args ...interface{}) ([]string, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sort function requires exactly one argument")
	}
	slice, ok := args[0].([]string)
	if !ok {
		return nil, fmt.Errorf("sort function requires a []string argument")
	}
	sort.Strings(slice)
	return slice, nil
}

func unique(args ...interface{}) ([]interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unique function requires exactly one argument")
	}
	slice, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unique function requires a slice argument")
	}
	seen := make(map[interface{}]bool)
	var result []interface{}
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result, nil
}

func in(args ...interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("in function requires exactly two arguments")
	}
	item := args[0]
	collection := args[1]
	switch v := collection.(type) {
	case string:
		s, ok := item.(string)
		if !ok {
			return false, fmt.Errorf("for string collection, item must be a string")
		}
		return strings.Contains(v, s), nil
	case []interface{}:
		for _, elem := range v {
			if reflect.DeepEqual(elem, item) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("in function requires a string or slice as the second argument")
	}
}

// Math functions

func abs(args ...interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("abs function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case int:
		return math.Abs(float64(v)), nil
	case float64:
		return math.Abs(v), nil
	default:
		return 0, fmt.Errorf("abs function requires a numeric argument")
	}
}

func ceil(args ...interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("ceil function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case int:
		return math.Ceil(float64(v)), nil
	case float64:
		return math.Ceil(v), nil
	default:
		return 0, fmt.Errorf("ceil function requires a numeric argument")
	}
}

func floor(args ...interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("floor function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case int:
		return math.Floor(float64(v)), nil
	case float64:
		return math.Floor(v), nil
	default:
		return 0, fmt.Errorf("floor function requires a numeric argument")
	}
}

func round(args ...interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("round function requires exactly two arguments")
	}
	n, ok1 := args[0].(float64)
	precision, ok2 := args[1].(int)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("round function requires (float64, int) arguments")
	}
	shift := math.Pow(10, float64(precision))
	return math.Round(n*shift) / shift, nil
}

func max(args ...interface{}) (float64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("max function requires at least one argument")
	}
	max := math.Inf(-1)
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			if float64(v) > max {
				max = float64(v)
			}
		case float64:
			if v > max {
				max = v
			}
		default:
			return 0, fmt.Errorf("max function requires numeric arguments")
		}
	}
	return max, nil
}

func min(args ...interface{}) (float64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("min function requires at least one argument")
	}
	min := math.Inf(1)
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			if float64(v) < min {
				min = float64(v)
			}
		case float64:
			if v < min {
				min = v
			}
		default:
			return 0, fmt.Errorf("min function requires numeric arguments")
		}
	}
	return min, nil
}

func sum(args ...interface{}) (float64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("sum function requires at least one argument")
	}
	var sum float64
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			sum += float64(v)
		case float64:
			sum += v
		default:
			return 0, fmt.Errorf("sum function requires numeric arguments")
		}
	}
	return sum, nil
}

func add(args ...interface{}) (float64, error) {
	return sum(args...)
}

func sub(args ...interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("sub function requires exactly two arguments")
	}
	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("sub function requires numeric arguments")
	}
	return a - b, nil
}

func mul(args ...interface{}) (float64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("mul function requires at least one argument")
	}
	product := 1.0
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			product *= float64(v)
		case float64:
			product *= v
		default:
			return 0, fmt.Errorf("mul function requires numeric arguments")
		}
	}
	return product, nil
}

func quo(args ...interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("quo function requires exactly two arguments")
	}
	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("quo function requires numeric arguments")
	}
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func mod(args ...interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("mod function requires exactly two arguments")
	}
	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("mod function requires numeric arguments")
	}
	if b == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return math.Mod(a, b), nil
}

func pow(args ...interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("pow function requires exactly two arguments")
	}
	base, ok1 := args[0].(float64)
	exp, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("pow function requires numeric arguments")
	}
	return math.Pow(base, exp), nil
}

func sqrt(args ...interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("sqrt function requires exactly one argument")
	}
	n, ok := args[0].(float64)
	if !ok {
		return 0, fmt.Errorf("sqrt function requires a numeric argument")
	}
	if n < 0 {
		return 0, fmt.Errorf("sqrt of negative number")
	}
	return math.Sqrt(n), nil
}

// Encoding functions

func b64encode(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("b64encode function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("b64encode function requires a string argument")
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func b64decode(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("b64decode function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("b64decode function requires a string argument")
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %v", err)
	}
	return string(decoded), nil
}

func urlencode(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("urlencode function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("urlencode function requires a string argument")
	}
	return url.QueryEscape(s), nil
}

func urldecode(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("urldecode function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("urldecode function requires a string argument")
	}
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return "", fmt.Errorf("failed to decode URL-encoded string: %v", err)
	}
	return decoded, nil
}

// Date functions

func now(args ...interface{}) (time.Time, error) {
	if len(args) != 0 {
		return time.Time{}, fmt.Errorf("now function takes no arguments")
	}
	return time.Now(), nil
}

func datetimeformat(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("datetimeformat function requires exactly two arguments")
	}
	t, ok1 := args[0].(time.Time)
	format, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("datetimeformat function requires (time.Time, string) arguments")
	}
	return t.Format(format), nil
}

func dateAdd(args ...interface{}) (time.Time, error) {
	if len(args) != 2 {
		return time.Time{}, fmt.Errorf("dateAdd function requires exactly two arguments")
	}
	t, ok1 := args[0].(time.Time)
	duration, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return time.Time{}, fmt.Errorf("dateAdd function requires (time.Time, string) arguments")
	}
	d, err := time.ParseDuration(duration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration: %v", err)
	}
	return t.Add(d), nil
}

func dateSub(args ...interface{}) (time.Time, error) {
	if len(args) != 2 {
		return time.Time{}, fmt.Errorf("dateSub function requires exactly two arguments")
	}
	t, ok1 := args[0].(time.Time)
	duration, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return time.Time{}, fmt.Errorf("dateSub function requires (time.Time, string) arguments")
	}
	d, err := time.ParseDuration(duration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration: %v", err)
	}
	return t.Add(-d), nil
}

// Logic functions
func defaultValue(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("default function requires exactly two arguments")
	}
	if args[0] != nil {
		return args[0], nil
	}
	return args[1], nil
}

func coalesce(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("coalesce function requires at least one argument")
	}
	for _, arg := range args {
		if arg != nil {
			return arg, nil
		}
	}
	return nil, nil
}

func ternary(args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("ternary function requires exactly three arguments")
	}
	condition, ok := args[0].(bool)
	if !ok {
		return nil, fmt.Errorf("ternary function's first argument must be a boolean")
	}
	if condition {
		return args[1], nil
	}
	return args[2], nil
}

// Type conversion functions

func toString(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("toString function requires exactly one argument")
	}
	return fmt.Sprintf("%v", args[0]), nil
}

func toInt(args ...interface{}) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("toInt function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %v to int", args[0])
	}
}

func toFloat(args ...interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("toFloat function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case int:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %v to float64", args[0])
	}
}

func toBool(args ...interface{}) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("toBool function requires exactly one argument")
	}
	switch v := args[0].(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %v to bool", args[0])
	}
}

// Additional useful functions

func contains(args ...interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("contains function requires exactly two arguments")
	}
	switch v := args[0].(type) {
	case string:
		substr, ok := args[1].(string)
		if !ok {
			return false, fmt.Errorf("for string, second argument must be a string")
		}
		return strings.Contains(v, substr), nil
	case []interface{}:
		for _, item := range v {
			if reflect.DeepEqual(item, args[1]) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("contains function requires a string or slice as the first argument")
	}
}

func hasPrefix(args ...interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("hasPrefix function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	prefix, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("hasPrefix function requires (string, string) arguments")
	}
	return strings.HasPrefix(s, prefix), nil
}

func hasSuffix(args ...interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("hasSuffix function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	suffix, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("hasSuffix function requires (string, string) arguments")
	}
	return strings.HasSuffix(s, suffix), nil
}

func repeat(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("repeat function requires exactly two arguments")
	}
	s, ok1 := args[0].(string)
	count, ok2 := args[1].(int)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("repeat function requires (string, int) arguments")
	}
	return strings.Repeat(s, count), nil
}

func matchRegex(args ...interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("matchRegex function requires exactly two arguments")
	}
	pattern, ok1 := args[0].(string)
	s, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("matchRegex function requires (string, string) arguments")
	}
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %v", err)
	}
	return matched, nil
}

func camelCase(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("camelCase function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("camelCase function requires a string argument")
	}

	var result strings.Builder
	nextUpper := false

	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if i == 0 {
				result.WriteRune(unicode.ToLower(r))
			} else if nextUpper {
				result.WriteRune(unicode.ToUpper(r))
				nextUpper = false
			} else {
				result.WriteRune(r)
			}
		} else {
			nextUpper = true
		}
	}

	return result.String(), nil
}

func snakeCase(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("snakeCase function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("snakeCase function requires a string argument")
	}
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String(), nil
}

func kebabCase(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("kebabCase function requires exactly one argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("kebabCase function requires a string argument")
	}
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String(), nil
}
