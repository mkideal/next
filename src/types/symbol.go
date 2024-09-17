package types

import (
	"fmt"
	"strings"
)

// -------------------------------------------------------------------------

// @api(Object/Common/Symbol) represents a Next symbol: value(const, enum member), type(enum, struct, interface).
type Symbol interface {
	LocatedObject

	symbolNode()
}

var _ Symbol = (*Value)(nil)
var _ Symbol = (*DeclType[Decl])(nil)

func (*Value) symbolNode()       {}
func (*DeclType[T]) symbolNode() {}

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
	// parent returns the parent scope of this scope.
	parent() Scope

	// LookupLocalSymbol looks up a symbol by name in the scope.
	LookupLocalSymbol(name string) Symbol
}

func (f *File) parent() Scope { return &fileParentScope{f} }
func (e *Enum) parent() Scope { return e.file }

func (f *File) LookupLocalSymbol(name string) Symbol { return f.symbols[name] }
func (e *Enum) LookupLocalSymbol(name string) Symbol {
	for _, m := range e.Members.List {
		if m.name == name {
			return m.value
		}
	}
	return nil
}

// LookupSymbol looks up a symbol by name in the given scope and its parent scopes.
func LookupSymbol(scope Scope, name string) Symbol {
	for s := scope; s != nil; s = s.parent() {
		if sym := s.LookupLocalSymbol(name); sym != nil {
			return sym
		}
	}
	return nil
}

// LookupType looks up a type by name in the given scope and its parent scopes.
func LookupType(scope Scope, name string) (Type, error) {
	return expectTypeSymbol(name, LookupSymbol(scope, name))
}

// LookupValue looks up a value by name in the given scope and its parent scopes.
func LookupValue(scope Scope, name string) (*Value, error) {
	return expectValueSymbol(name, LookupSymbol(scope, name))
}

func expectTypeSymbol(name string, s Symbol) (Type, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if t, ok := s.(Type); ok {
		return t, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "Type", Got: fmt.Sprintf("%T", s)}
}

func expectValueSymbol(name string, s Symbol) (*Value, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*Value); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "Value", Got: fmt.Sprintf("%T", s)}
}

type fileParentScope struct {
	f *File
}

func (s *fileParentScope) parent() Scope {
	return nil
}

func (s *fileParentScope) LookupLocalSymbol(name string) Symbol {
	var files []*File
	pkg, name := splitSymbolName(name)
	for i := range s.f.imports.List {
		target := s.f.imports.List[i].target
		if target != nil && target.pkg != nil && target.pkg.name == pkg {
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
