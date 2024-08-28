package types

import (
	"strconv"

	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

// @api(template/annotation) AnnotationParam
// AnnotationParam represents an annotation parameter.
// If Name is empty, the Value is not nil.
//
// Example:
//
// ```
// @json(omitempty)
// @event(name="Login")
// @message(name="Login", type=100)
// ```
type AnnotationParam struct {
	name  string
	value constant.Value
}

func (a *AnnotationParam) Name() string {
	return a.name
}

func (a *AnnotationParam) Value() constant.Value {
	if a == nil {
		return nil
	}
	return a.value
}

func (a *AnnotationParam) String() string {
	if a == nil {
		return ""
	}
	if a.name != "" {
		if a.value == nil {
			return a.name
		}
		return a.name + "=" + a.value.String()
	}
	return a.value.String()
}

// @api(template/annotation) AnnotationParam.AsString
func (a *AnnotationParam) AsString() (string, error) {
	if a == nil {
		return "", ErrParamNotFound
	}
	if a.value.Kind() == constant.String {
		return strconv.Unquote(a.value.String())
	}
	return "", ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.AsInt
func (a *AnnotationParam) AsInt() (int64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.value.Kind() == constant.Int {
		return strconv.ParseInt(a.value.String(), 10, 64)
	}
	return 0, ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.AsFloat
func (a *AnnotationParam) AsFloat() (float64, error) {
	if a == nil {
		return 0, ErrParamNotFound
	}
	if a.value.Kind() == constant.Float {
		return strconv.ParseFloat(a.value.String(), 64)
	}
	return 0, ErrUnpexpectedParamType
}

// @api(template/annotation) AnnotationParam.AsBool
func (a *AnnotationParam) AsBool() (bool, error) {
	if a == nil {
		return false, ErrParamNotFound
	}
	if a.value == nil {
		// Default value is true if the named parameter has no value.
		return true, nil
	}
	if a.value.Kind() == constant.Bool {
		return constant.BoolVal(a.value), nil
	}
	return false, ErrUnpexpectedParamType
}

// @api(template/annotation) Annotation
// Annotation represents an annotation.
type Annotation struct {
	pos    token.Position
	name   string
	params []*AnnotationParam
}

func (a *Annotation) Name() string {
	return a.name
}

func (a *Annotation) Params() []*AnnotationParam {
	if a == nil {
		return nil
	}
	return a.params
}

// @api(template/annotation) Annotation.Contains
func (a *Annotation) Contains(name string) bool {
	if a == nil {
		return false
	}
	for _, p := range a.params {
		if p.name == name {
			return true
		}
	}
	return false
}

// @api(template/annotation) Annotation.Param
func (a *Annotation) Param(name string) *AnnotationParam {
	if a == nil {
		return nil
	}
	for _, p := range a.params {
		if p.name == name {
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
		if a.name == name {
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
		if a.name == name {
			return &a
		}
	}
	return nil
}
