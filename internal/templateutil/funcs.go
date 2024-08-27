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

	// @api(template/funcs) quote (s: string)
	// `quote` returns a double-quoted string literal representing s.
	// Special characters are escaped with backslashes.
	//
	// Example:
	//
	// ```
	// {{quote "Hello, World!"}}
	// ```
	// or
	// ```
	// {{"Hello, World!" | quote}}
	// ```
	//
	// Output:
	// ```
	// "Hello, World!"
	// ```
	"quote": strconv.Quote,

	// @api(template/funcs) unquote (s: string)
	// `unquote` interprets s as a double-quoted string literal and returns the string value that s represents.
	// Special characters are unescaped.
	//
	// Example:
	//
	// ```
	// {{unquote "\"Hello, World!\""}}
	// ```
	// or
	// ```
	// {{"\"Hello, World!\"" | unquote}}
	// ```
	//
	// Output:
	// ```
	// Hello, World!
	// ```
	"unquote": strconv.Unquote,

	// @api(template/funcs) capitalize (s: string)
	// `capitalize` capitalizes the first character of the given string.
	//
	// Example:
	//
	// ```
	// {{capitalize "hello world"}}
	// ```
	// or
	// ```
	// {{"hello world" | capitalize}}
	// ```
	//
	// Output:
	// ```
	// Hello world
	// ```
	"capitalize": capitalize,

	// @api(template/funcs) lower (s: string)
	// `lower` converts the entire string to lowercase.
	//
	// Example:
	//
	// ```
	// {{lower "Hello World"}}
	// ```
	// or
	// ```
	// {{"Hello World" | lower}}
	// ```
	//
	// Output:
	// ```
	// hello world
	// ```
	"lower": strings.ToLower,

	// @api(template/funcs) upper (s: string)
	// `upper` converts the entire string to uppercase.
	//
	// Example:
	//
	// ```
	// {{upper "hello world"}}
	// ```
	// or
	// ```
	// {{"hello world" | upper}}
	// ```
	//
	// Output:
	// ```
	// HELLO WORLD
	// ```
	"upper": strings.ToUpper,

	// @api(template/funcs) replace (s: string, old: string, new: string[, n: int])
	// `replace` returns a copy of the string s with the first n non-overlapping instances of old replaced by new.
	//
	// Example:
	//
	// ```
	// {{replace "oink oink oink" "oink" "moo" 2}}
	// {{replace "oink oink oink" "oink" "moo"}}
	// ```
	//
	// Output:
	// ```
	// moo moo oink
	// moo moo moo
	// ```
	"replace": replace,

	// @api(template/funcs) trim (s: string)
	// `trim` returns a slice of the string s with all leading and trailing white space removed.
	//
	// Example:
	//
	// ```
	// {{trim "  hello world  "}}
	// ```
	// or
	// ```
	// {{"  hello world  " | trim}}
	// ```
	//
	// Output:
	// ```
	// hello world
	// hello world
	// ```
	"trim": strings.TrimSpace,

	// @api(template/funcs) trimPrefix (prefix: string, s: string)
	// `trimPrefix` returns s without the provided leading prefix string.
	//
	// Example:
	//
	// ```
	// {{trimPrefix "Hello" "Hello, World!"}}
	// ```
	// or
	// ```
	// {{"Hello, World!" | trimPrefix "Hello"}}
	// ```
	//
	// Output:
	// ```
	// , World!
	// ```
	"trimPrefix": trimPrefix,

	// @api(template/funcs) trimSuffix (suffix: string, s: string)
	// `trimSuffix` returns s without the provided trailing suffix string.
	//
	// Example:
	//
	// ```
	// {{trimSuffix "!" "Hello, World!"}}
	// ```
	// or
	// ```
	// {{"Hello, World!" | trimSuffix "!"}}
	// ```
	//
	// Output:
	// ```
	// Hello, World
	// ```
	"trimSuffix": trimSuffix,

	// @api(template/funcs) split (sep: string, s: string)
	// `split` slices s into all substrings separated by sep and returns a slice of the substrings between those separators.
	//
	// Example:
	//
	// ```
	// {{split "," "a,b,c"}}
	// ```
	// or
	// ```
	// {{"a,b,c" | split ","}}
	// ```
	//
	// Output:
	// ```
	// [a b c]
	// ```
	"split": split,

	// @api(template/funcs) join (sep: string, v: []string)
	// `join` concatenates the elements of v to create a single string. The separator string sep is placed between elements in the resulting string.
	//
	// Example:
	//
	// ```
	// {{join ", " (slice "apple" "banana" "cherry")}}
	// ```
	//
	// Output:
	// ```
	// apple, banana, cherry
	// ```
	"join": join,

	// @api(template/funcs) striptags (s: string)
	// `striptags` removes all HTML tags from the given string.
	//
	// Example:
	//
	// ```
	// {{striptags "<p>Hello <b>World</b>!</p>"}}
	// ```
	// or
	// ```
	// {{"<p>Hello <b>World</b>!</p>" | striptags}}
	// ```
	//
	// Output:
	// ```
	// Hello World!
	// ```
	"striptags": striptags,

	// @api(template/funcs) substr (start: int, length: int, s: string)
	// `substr` returns a substring of s starting at index start with the given length.
	//
	// Example:
	//
	// ```
	// {{substr 7 5 "Hello, World!"}}
	// ```
	// or
	// ```
	// {{"Hello, World!" | substr 7 5}}
	// ```
	//
	// Output:
	// ```
	// World
	// ```
	"substr": substr,

	// @api(template/funcs) repeat (count: int, s: string)
	// `repeat` returns a new string consisting of count copies of the string s.
	//
	// Example:
	//
	// ```
	// {{repeat 3 "na"}}
	// ```
	// or
	// ```
	// {{"na" | repeat 3}}
	// ```
	//
	// Output:
	// ```
	// nanana
	// ```
	"repeat": strings.Repeat,

	// @api(template/funcs) camelCase (s: string)
	// `camelCase` converts the given string to camel case.
	//
	// Example:
	//
	// ```
	// {{camelCase "hello world"}}
	// ```
	// or
	// ```
	// {{"hello world" | camelCase}}
	// ```
	//
	// Output:
	// ```
	// helloWorld
	// ```
	"camelCase": camelCase,

	// @api(template/funcs) snakeCase (s: string)
	// `snakeCase` converts the given string to snake case.
	//
	// Example:
	//
	// ```
	// {{snakeCase "helloWorld"}}
	// ```
	// or
	// ```
	// {{"helloWorld" | snakeCase}}
	// ```
	//
	// Output:
	// ```
	// hello_world
	// ```
	"snakeCase": snakeCase,

	// @api(template/funcs) kebabCase (s: string)
	// `kebabCase` converts the given string to kebab case.
	//
	// Example:
	//
	// ```
	// {{kebabCase "helloWorld"}}
	// ```
	// or
	// ```
	// {{"helloWorld" | kebabCase}}
	// ```
	//
	// Output:
	// ```
	// hello-world
	// ```
	"kebabCase": kebabCase,

	// @api(template/funcs) truncate (s: string, length: int, suffix: string)
	// `truncate` truncates the given string to the specified length and adds the suffix if truncation occurred.
	//
	// Example:
	//
	// ```
	// {{truncate 10 "..." "This is a long sentence."}}
	// ```
	// or
	// ```
	// {{"This is a long sentence." | truncate 10 "..."}}
	// ```
	//
	// Output:
	// ```
	// This is...
	// ```
	"truncate": truncate,

	// @api(template/funcs) wordwrap (s: string, width: int)
	// `wordwrap` wraps the given string to a maximum width.
	//
	// Example:
	//
	// ```
	// {{wordwrap 20 "This is a long sentence that needs wrapping."}}
	// ```
	// or
	// ```
	// {{"This is a long sentence that needs wrapping." | wordwrap 20}}
	// ```
	//
	// Output:
	// ```
	// This is a long
	// sentence that needs
	// wrapping.
	// ```
	"wordwrap": wordwrap,

	// @api(template/funcs) center (s: string, width: int)
	// `center` centers the string in a field of the specified width.
	//
	// Example:
	//
	// ```
	// {{center 11 "hello"}}
	// ```
	// or
	// ```
	// {{"hello" | center 11}}
	//
	// Output:
	// ```
	//    hello
	// ```
	"center": center,

	// @api(template/funcs) matchRegex (pattern: string, s: string)
	// `matchRegex` checks if the string matches the given regular expression pattern.
	//
	// Example:
	//
	// ```
	// {{matchRegex "^[a-z]+$" "hello"}}
	// ```
	// or
	// ```
	// {{"hello" | matchRegex "^[a-z]+$"}}
	// ```
	//
	// Output:
	// ```
	// true
	// ```
	"matchRegex": matchRegex,

	// Escaping functions

	// @api(template/funcs) html (s: string)
	// `html` escapes special characters like "<" to become "&lt;". It escapes only five such characters: <, >, &, ' and ".
	//
	// Example:
	//
	// ```
	// {{html "<script>alert('XSS')</script>"}}
	// ```
	// or
	// ```
	// {{"<script>alert('XSS')</script>" | html}}
	// ```
	//
	// Output:
	// ```
	// &lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;
	// ```
	"html": html.EscapeString,

	// @api(template/funcs) urlquery (s: string)
	// `urlquery` escapes the string so it can be safely placed inside a URL query.
	//
	// Example:
	//
	// ```
	// {{urlquery "hello world"}}
	// ```
	// or
	// ```
	// {{"hello world" | urlquery}}
	// ```
	//
	// Output:
	// ```
	// hello+world
	// ```
	"urlquery": url.QueryEscape,

	// @api(template/funcs) urlUnescape (s: string)
	// `urlUnescape` does the inverse transformation of urlquery, converting each 3-byte encoded substring of the form "%AB" into the hex-decoded byte 0xAB.
	//
	// Example:
	//
	// ```
	// {{urlUnescape "hello+world"}}
	// ```
	// or
	// ```
	// {{"hello+world" | urlUnescape}}
	// ```
	//
	// Output:
	// ```
	// hello world
	// ```
	"urlUnescape": urlUnescape,

	// Encoding functions

	// @api(template/funcs) b64enc (v: string)
	// `b64enc` encodes the given string to base64.
	//
	// Example:
	//
	// ```
	// {{b64enc "hello world"}}
	// ```
	// or
	// ```
	// {{"hello world" | b64enc}}
	// ```
	//
	// Output:
	// ```
	// aGVsbG8gd29ybGQ=
	// ```
	"b64enc": base64.StdEncoding.EncodeToString,

	// @api(template/funcs) b64dec (s: string)
	// `b64dec` decodes the given base64 string.
	//
	// Example:
	//
	// ```
	// {{b64dec "aGVsbG8gd29ybGQ="}}
	// ```
	// or
	// ```
	// {{"aGVsbG8gd29ybGQ=" | b64dec}}
	// ```
	//
	// Output:
	// ```
	// hello world
	// ```
	"b64dec": b64dec,

	// List functions

	// @api(template/funcs) first (list: []any)
	// `first` returns the first element of a list.
	//
	// Example:
	//
	// ```
	// {{first (slice 1 2 3)}}
	// ```
	//
	// Output:
	// ```
	// 1
	// ```
	"first": first,

	// @api(template/funcs) last (list: []any)
	// `last` returns the last element of a list.
	//
	// Example:
	//
	// ```
	// {{last (slice 1 2 3)}}
	// ```
	//
	// Output:
	// ```
	// 3
	// ```
	"last": last,

	// @api(template/funcs) rest (list: []any)
	// `rest` returns all elements of a list except the first.
	//
	// Example:
	//
	// ```
	// {{rest (slice 1 2 3)}}
	// ```
	//
	// Output:
	// ```
	// [2 3]
	// ```
	"rest": rest,

	// @api(template/funcs) reverse (list: []any)
	// `reverse` reverses the order of elements in a list.
	//
	// Example:
	//
	// ```
	// {{reverse (slice 1 2 3)}}
	// ```
	//
	// Output:
	// ```
	// [3 2 1]
	// ```
	"reverse": reverse,

	// @api(template/funcs) sort (list: []string)
	// `sort` sorts a list of strings in ascending order.
	//
	// Example:
	//
	// ```
	// {{sort (slice "banana" "apple" "cherry")}}
	// ```
	//
	// Output:
	// ```
	// [apple banana cherry]
	// ```
	"sort": sortStrings,

	// @api(template/funcs) uniq (list: []any)
	// `uniq` removes duplicate elements from a list.
	//
	// Example:
	//
	// ```
	// {{uniq (slice 1 2 2 3 3 3)}}
	// ```
	//
	// Output:
	// ```
	// [1 2 3]
	// ```
	"uniq": uniq,

	// @api(template/funcs) in (item: any, list: []any)
	// `in` checks if an item is present in a list.
	//
	// Example:
	//
	// ```
	// {{in "b" (slice "a" "b" "c")}}
	// ```
	//
	// Output:
	// ```
	// true
	// ```
	"in": in,

	// Math functions

	// @api(template/funcs) add (a: number, b: number)
	// `add` adds two numbers.
	//
	// Example:
	//
	// ```
	// {{add 1 2}}
	// ```
	//
	// Output:
	// ```
	// 3
	// ```
	"add": add,

	// @api(template/funcs) sub (a: number, b: number)
	// `sub` subtracts the second number from the first.
	//
	// Example:
	//
	// ```
	// {{sub 5 3}}
	// ```
	//
	// Output:
	// ```
	// 2
	// ```
	"sub": sub,

	// @api(template/funcs) mul (a: number, b: number)
	// `mul` multiplies two numbers.
	//
	// Example:
	//
	// ```
	// {{mul 2 3}}
	// ```
	//
	// Output:
	// ```
	// 6
	// ```
	"mul": mul,

	// @api(template/funcs) quo (a: number, b: number)
	// `quo` divides the first number by the second.
	//
	// Example:
	//
	// ```
	// {{quo 6 3}}
	// ```
	//
	// Output:
	// ```
	// 2
	// ```
	"quo": quo,

	// @api(template/funcs) rem (a: number, b: number)
	// `rem` returns the remainder of dividing the first number by the second.
	//
	// Example:
	//
	// ```
	// {{rem 7 3}}
	// ```
	//
	// Output:
	// ```
	// 1
	// ```
	"rem": rem,

	// @api(template/funcs) mod (a: number, b: number)
	// `mod` returns the modulus of dividing the first number by the second.
	//
	// Example:
	//
	// ```
	// {{mod -7 3}}
	// ```
	//
	// Output:
	// ```
	// 2
	// ```
	"mod": mod,

	// @api(template/funcs) max (numbers: ...number)
	// `max` returns the largest of the given numbers.
	//
	// Example:
	//
	// ```
	// {{max 1 5 3 2}}
	// ```
	//
	// Output:
	// ```
	// 5
	// ```
	"max": max,

	// @api(template/funcs) min (numbers: ...number)
	// `min` returns the smallest of the given numbers.
	//
	// Example:
	//
	// ```
	// {{min 1 5 3 2}}
	// ```
	//
	// Output:
	// ```
	// 1
	// ```
	"min": min,

	// @api(template/funcs) ceil (x: number)
	// `ceil` returns the least integer value greater than or equal to x.
	//
	// Example:
	//
	// ```
	// {{ceil 1.5}}
	// ```
	//
	// Output:
	// ```
	// 2
	// ```
	"ceil": ceil,

	// @api(template/funcs) floor (x: number)
	// `floor` returns the greatest integer value less than or equal to x.
	//
	// Example:
	//
	// ```
	// {{floor 1.5}}
	// ```
	//
	// Output:
	// ```
	// 1
	// ```
	"floor": floor,

	// @api(template/funcs) round (x: number, precision: int)
	// `round` returns x rounded to the specified precision.
	//
	// Example:
	//
	// ```
	// {{round 1.234 2}}
	// ```
	//
	// Output:
	// ```
	// 1.23
	// ```
	"round": round,

	// Type conversion functions

	// @api(template/funcs) int (v: any)
	// `int` converts the given value to an integer.
	//
	// Example:
	//
	// ```
	// {{int "123"}}
	// ```
	//
	// Output:
	// ```
	// 123
	// ```
	"int": toInt,

	// @api(template/funcs) float (v: any)
	// `float` converts the given value to a float.
	//
	// Example:
	//
	// ```
	// {{float "1.23"}}
	// ```
	//
	// Output:
	// ```
	// 1.23
	// ```
	"float": toFloat,

	// @api(template/funcs) string (v: any)
	// `string` converts the given value to a string.
	//
	// Example:
	//
	// ```
	// {{string 123}}
	// ```
	//
	// Output:
	// ```
	// 123
	// ```
	"string": toString,

	// @api(template/funcs) bool (v: any)
	// `bool` converts the given value to a boolean.
	//
	// Example:
	//
	// ```
	// {{bool 1}}
	// ```
	//
	// Output:
	// ```
	// true
	// ```
	"bool": toBool,

	// Date functions

	// @api(template/funcs) now ()
	// `now` returns the current local time.
	//
	// Example:
	//
	// ```
	// {{now}}
	// {{now.Format "2006-01-02 15:04:05"}}
	// ```
	//
	// Output:
	// ```
	// 2023-05-17 14:30:45.123456789 +0000 UTC
	// 2023-05-17 14:30:45
	// ```
	"now": time.Now,

	// @api(template/funcs) parse (layout: string, value: string)
	// `parse` parses a formatted string and returns the time value it represents.
	//
	// Example:
	//
	// ```
	// {{parseTime "2006-01-02" "2023-05-17"}}
	// ```
	//
	// Output:
	// ```
	// 2023-05-17 00:00:00 +0000 UTC
	// ```
	"parseTime": parseTime,

	// Conditional functions

	// @api(template/funcs) ternary (cond: bool, trueVal: any, falseVal: any)
	// `ternary` returns trueVal if the condition is true, falseVal otherwise.
	//
	// Example:
	//
	// ```
	// {{ternary (eq 1 1) "yes" "no"}}
	// ```
	//
	// Output:
	// ```
	// yes
	// ```
	"ternary": ternary,
}

// String functions

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func replace(s, old, new string, n ...reflect.Value) (string, error) {
	if len(n) == 0 {
		return strings.Replace(s, old, new, -1), nil
	}
	if len(n) > 1 {
		return "", fmt.Errorf("replace: too many arguments")
	}
	switch n[0].Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strings.Replace(s, old, new, int(n[0].Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strings.Replace(s, old, new, int(n[0].Uint())), nil
	default:
		return "", fmt.Errorf("replace: unsupported type %s", n[0].Type())
	}
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

func truncate(length int, suffix, s string) string {
	if length <= 0 {
		return ""
	}
	if len(s) <= length {
		return s
	}
	return s[:length-len(suffix)] + suffix
}

func wordwrap(width int, s string) string {
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

func center(width int, s string) string {
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
