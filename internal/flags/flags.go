package flags

import (
	"fmt"
	"strings"
)

type Map map[string]string

func (m *Map) Set(s string) error {
	if *m == nil {
		*m = make(Map)
	}
	var k, v string
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 1 {
		k = parts[0]
		v = "true"
	} else if len(parts) == 2 {
		k, v = parts[0], parts[1]
	} else {
		return fmt.Errorf("invalid format: %q, expected key=value", s)
	}
	if _, dup := (*m)[k]; dup {
		return fmt.Errorf("already set: %q", k)
	}
	(*m)[k] = v
	return nil
}

func (m Map) String() string {
	if m == nil {
		return ""
	}
	var sb strings.Builder
	for k, v := range m {
		if sb.Len() > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "%s=%s", k, v)
	}
	return sb.String()
}

type Slice []string

func (s *Slice) Set(v string) error {
	values := strings.Split(v, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		*s = append(*s, v)
	}
	return nil
}

func (s Slice) String() string {
	return strings.Join(s, ",")
}
