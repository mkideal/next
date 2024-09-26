package compile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrParamNotFound          = errors.New("param not found")
	ErrUnpexpectedParamType   = errors.New("unexpected param type")
	ErrUnexpectedConstantType = errors.New("unexpected constant type")
)

type TemplateNotFoundError struct {
	Name string
}

func (e *TemplateNotFoundError) Error() string {
	return "template " + e.Name + " not found"
}

func IsTemplateNotFoundError(err error) bool {
	var e *TemplateNotFoundError
	return errors.As(err, &e)
}

type SymbolNotFoundError struct {
	Name string
}

func (e *SymbolNotFoundError) Error() string {
	return "symbol " + e.Name + " not found"
}

type UnexpectedSymbolTypeError struct {
	Name string
	Want string
	Got  string
}

func (e *UnexpectedSymbolTypeError) Error() string {
	return "symbol " + e.Name + " is not of type " + e.Want + " but " + e.Got
}

type SymbolRedefinedError struct {
	Name string
	Prev Symbol
}

func (e *SymbolRedefinedError) Error() string {
	return "symbol " + e.Name + " redefined"
}

func removeAllSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func parseLangMap[M ~map[string]string](m M, lang string, content []byte) error {
	s := bufio.NewScanner(bytes.NewReader(content))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		index := strings.Index(line, "=")
		if index < 1 {
			return fmt.Errorf("invalid format: %q, expected key=value", line)
		}
		k, v := removeAllSpaces(line[:index]), strings.TrimSpace(line[index+1:])
		if len(k) == 0 || len(v) == 0 {
			return fmt.Errorf("invalid format: %q, expected key=value", line)
		}
		m[lang+"."+k] = v
	}
	return s.Err()
}

var boxRegexp = regexp.MustCompile(`box\(([^)]*)\)`)

func removeBox(input string) string {
	return boxRegexp.ReplaceAllString(input, "$1")
}
