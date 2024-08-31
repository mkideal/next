package types

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

func removeAllSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func parseLangMap[M ~map[string]string](m M, lang string, r io.Reader) error {
	s := bufio.NewScanner(r)
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
