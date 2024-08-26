package types

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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

func parseLangTypes[M ~map[string]string](m M, lang string, r io.Reader) error {
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

func resolveLangType[M ~map[string]string](m M, lang string, t Type) (string, error) {
	switch t := t.(type) {
	case *BasicType:
		p, ok := m[lang+"."+t.name]
		if !ok {
			return "", fmt.Errorf("type %q not found", t.name)
		}
		return p, nil

	case *MapType:
		p, ok := m[lang+".map<%K%,%V%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "map<%K%,%V%>")
		}
		if strings.Contains(p, "%K%") {
			k, err := resolveLangType(m, lang, t.KeyType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%K%", k)
		}
		if strings.Contains(p, "%V%") {
			v, err := resolveLangType(m, lang, t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%V%", v)
		}
		return p, nil

	case *VectorType:
		p, ok := m[lang+".vector<%T%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "vector<%T%>")
		}
		if strings.Contains(p, "%T%") {
			e, err := resolveLangType(m, lang, t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T%", e)
		}
		return p, nil

	case *ArrayType:
		p, ok := m[lang+".array<%T%,%N%>"]
		if !ok {
			return "", fmt.Errorf("type %q not found", "array<%T%,%N%>")
		}
		if strings.Contains(p, "%T%") {
			e, err := resolveLangType(m, lang, t.ElemType)
			if err != nil {
				return "", err
			}
			p = strings.ReplaceAll(p, "%T%", e)
		}
		if strings.Contains(p, "%N%") {
			p = strings.ReplaceAll(p, "%N%", strconv.FormatUint(t.N, 10))
		}
		return p, nil

	default:
		name := t.String()
		p, ok := m[lang+"."+name]
		if ok {
			return p, nil
		}
		return name, nil
	}
}