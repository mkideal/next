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

// @api(template/annotation) AnnotationParam
// AnnotationParam represents an annotation parameter.
// If Name is empty, the Value is not nil.
//
// Example:
//
// ```
// @json(omitempty)
// @type(100)
// @event(name="Login")
// @message("Login", type=100)
// ```
type AnnotationParam struct {
	Name  string
	Value constant.Value
}

func (a *AnnotationParam) String() string {
	if a == nil {
		return ""
	}
	if a.Name != "" {
		if a.Value == nil {
			return a.Name
		}
		return a.Name + "=" + a.Value.String()
	}
	return a.Value.String()
}

// @api(template/annotation) AnnotationParam.IsNamed
func (a *AnnotationParam) IsNamed() bool {
	if a == nil {
		return false
	}
	return a.Name != ""
}

// @api(template/annotation) AnnotationParam.HasValue
func (a *AnnotationParam) HasValue() bool {
	if a == nil {
		return false
	}
	return a.Value != nil
}

// @api(template/annotation) AnnotationParam.GetString
func (a *AnnotationParam) GetString() (string, error) {
	if a == nil {
		return "", ErrParamNotFound
	}
	if a.Value.Kind() == constant.String {
		return strconv.Unquote(a.Value.String())
	}
	return "", ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.GetInt
func (a *AnnotationParam) GetInt() (int64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Int {
		return strconv.ParseInt(a.Value.String(), 10, 64)
	}
	return 0, ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.GetFloat
func (a *AnnotationParam) GetFloat() (float64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Float {
		return strconv.ParseFloat(a.Value.String(), 64)
	}
	return 0, ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.GetBool
func (a *AnnotationParam) GetBool() (bool, error) {
	if a == nil {
		return false, ErrParamNotFound
	}
	if a.Value.Kind() == constant.Bool {
		return strconv.ParseBool(a.Value.String())
	}
	return false, ErrUnpexpectedParamType
}

// @api(template/annotation) Annotation
// Annotation represents an annotation.
type Annotation struct {
	pos    token.Position
	Name   string
	Params []*AnnotationParam
}

// @api(template/annotation) Annotation.Contains
func (a *Annotation) Contains(name string) bool {
	for _, p := range a.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

// @api(template/annotation) Annotation.At
func (a *Annotation) At(i int) *AnnotationParam {
	if i < 0 || i >= len(a.Params) {
		return nil
	}
	return a.Params[i]
}

// @api(template/annotation) Annotation.Param
func (a *Annotation) Param(name string) *AnnotationParam {
	for _, p := range a.Params {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// @api(template/annotation) AnnotationGroup
// AnnotationGroup represents a group of annotations.
type AnnotationGroup struct {
	list []Annotation
}

// @api(template/annotation) AnnotationGroup.Contains
func (ag *AnnotationGroup) Contains(name string) bool {
	if ag == nil {
		return false
	}
	for _, a := range ag.list {
		if a.Name == name {
			return true
		}
	}
	return false
}

// @api(template/annotation) AnnotationGroup.Lookup
func (ag *AnnotationGroup) Lookup(name string) *Annotation {
	if ag == nil {
		return nil
	}
	for _, a := range ag.list {
		if a.Name == name {
			return &a
		}
	}
	return nil
}
