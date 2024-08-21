package types

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
