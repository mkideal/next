package types

import "strings"

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

func (f *File) ParentScope() Scope     { return &fileParentScope{f} }
func (e *EnumType) ParentScope() Scope { return e.decl.file }

func (f *File) LookupLocalSymbol(name string) Symbol { return f.symbols[name] }
func (e *EnumType) LookupLocalSymbol(name string) Symbol {
	for _, m := range e.Members {
		if m.name == name {
			return m
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

func lookupValue(scope Scope, name string) (*ValueSpec, error) {
	return expectValueSymbol(name, lookupSymbol(scope, name))
}

func expectTypeSymbol(name string, s Symbol) (Type, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if t, ok := s.(Type); ok {
		return t, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "type", Got: s.SymbolType()}
}

func expectValueSymbol(name string, s Symbol) (*ValueSpec, error) {
	if s == nil {
		return nil, &SymbolNotFoundError{Name: name}
	}
	if v, ok := s.(*ValueSpec); ok {
		return v, nil
	}
	return nil, &UnexpectedSymbolTypeError{Name: name, Want: "value", Got: s.SymbolType()}
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
	for i := range s.f.imports {
		if s.f.imports[i].importedFile.pkg == pkg {
			files = append(files, s.f.imports[i].importedFile)
		}
	}
	for _, file := range files {
		if s := file.symbols[name]; s != nil {
			return s
		}
	}
	return nil
}
