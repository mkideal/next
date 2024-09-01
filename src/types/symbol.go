package types

import (
	"strings"

	"github.com/next/next/src/token"
)

// Symbol represents a Next symbol: value(constant, enum member), type(struct, enum)
type Symbol interface {
	Node
	symbolPos() token.Pos
	symbolType() string
}

const (
	ValueSymbol = "value"
	TypeSymbol  = "type"
)

func (x *Value) symbolPos() token.Pos       { return x.namePos }
func (t *DeclType[T]) symbolPos() token.Pos { return t.pos }

func (*Value) symbolType() string         { return ValueSymbol }
func (t *DeclType[T]) symbolType() string { return TypeSymbol }

func splitSymbolName(name string) (ns, sym string) {
	if i := strings.Index(name, "."); i >= 0 {
		return name[:i], name[i+1:]
	}
	return "", name
}

func joinSymbolName(syms ...string) string {
	return strings.Join(syms, ".")
}

// Scope represents a symbol scope.
type Scope interface {
	ParentScope() Scope
	LookupLocalSymbol(name string) Symbol
}

func (f *File) ParentScope() Scope { return &fileParentScope{f} }
func (e *Enum) ParentScope() Scope { return e.file }

func (f *File) LookupLocalSymbol(name string) Symbol { return f.symbols[name] }
func (e *Enum) LookupLocalSymbol(name string) Symbol {
	for _, m := range e.Members.List {
		if string(m.name) == name {
			return m.value
		}
	}
	return nil
}

func lookupSymbol(scope Scope, name string) Symbol {
	for s := scope; s != nil; s = s.ParentScope() {
		if sym := s.LookupLocalSymbol(name); sym != nil {
			return sym
		}
	}
	return nil
}

func lookupType(scope Scope, name string) (Type, error) {
	return expectTypeSymbol(name, lookupSymbol(scope, name))
}

func lookupValue(scope Scope, name string) (*Value, error) {
	return expectValueSymbol(name, lookupSymbol(scope, name))
}

func expectTypeSymbol(name string, s Symbol) (Type, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if t, ok := s.(Type); ok {
		return t, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "type", Got: s.symbolType()}
}

func expectValueSymbol(name string, s Symbol) (*Value, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*Value); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "value", Got: s.symbolType()}
}

type fileParentScope struct {
	f *File
}

func (s *fileParentScope) ParentScope() Scope {
	return nil
}

func (s *fileParentScope) LookupLocalSymbol(name string) Symbol {
	var files []*File
	pkg, name := splitSymbolName(name)
	for i := range s.f.imports.List {
		if s.f.imports.List[i].target.pkg.name == pkg {
			files = append(files, s.f.imports.List[i].target)
		}
	}
	for _, file := range files {
		if s := file.symbols[name]; s != nil {
			return s
		}
	}
	return nil
}
