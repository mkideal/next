package types

import "errors"

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
