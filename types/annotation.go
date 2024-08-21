package types

import (
	"errors"
	"strconv"

	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

var (
	ErrParamNotFound        = errors.New("param not found")
	ErrUnpexpectedParamType = errors.New("unexpected param type")
)

// AnnotationParam represents an annotation parameter.
// If Name is empty, the Value is not nil.
//
// Example:
//
//	@json(omitempty)
//	@type(100)
//	@event(name="Login")
//	@message("Login", type=100)
type AnnotationParam struct {
	Name  string
	Value constant.Value
}

func (a *AnnotationParam) IsNamed() bool {
	if a == nil {
		return false
	}
	return a.Name != ""
}

func (a *AnnotationParam) IsValue() bool {
	if a == nil {
		return false
	}
	return a.Value != nil
}

func (a *AnnotationParam) String() string {
	if a == nil {
		return ""
	}
	if a.IsNamed() {
		return a.Name + "=" + a.Value.String()
	}
	return a.Value.String()
}

func (a *AnnotationParam) Str() (string, error) {
	if a == nil {
		return "", ErrParamNotFound
	}
	if a.Value.Kind() == constant.String {
		return strconv.Unquote(a.Value.String())
	}
	return "", ErrUnpexpectedParamType
}

func (a *AnnotationParam) Int() (int64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Int {
		return strconv.ParseInt(a.Value.String(), 10, 64)
	}
	return 0, ErrUnpexpectedParamType
}

func (a *AnnotationParam) Float() (float64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Float {
		return strconv.ParseFloat(a.Value.String(), 64)
	}
	return 0, ErrUnpexpectedParamType
}

func (a *AnnotationParam) Bool() (bool, error) {
	if a == nil {
		return false, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Bool {
		return strconv.ParseBool(a.Value.String())
	}
	return false, ErrUnpexpectedParamType
}

// Annotation represents an annotation.
type Annotation struct {
	pos    token.Position
	Name   string
	Params []*AnnotationParam
}

func (a *Annotation) Contains(name string) bool {
	for _, p := range a.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (a *Annotation) Get(i int) *AnnotationParam {
	if i < 0 || i >= len(a.Params) {
		return nil
	}
	return a.Params[i]
}

func (a *Annotation) Find(name string) *AnnotationParam {
	for _, p := range a.Params {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// AnnotationGroup represents a group of annotations.
type AnnotationGroup struct {
	List []Annotation
}

func (ag *AnnotationGroup) Contains(name string) bool {
	for _, a := range ag.List {
		if a.Name == name {
			return true
		}
	}
	return false
}

func (ag *AnnotationGroup) Find(name string) *Annotation {
	for _, a := range ag.List {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (ag *AnnotationGroup) FindParam(name, param string) *AnnotationParam {
	for _, a := range ag.List {
		if a.Name == name {
			return a.Find(param)
		}
	}
	return nil
}
