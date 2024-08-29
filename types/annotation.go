package types

import (
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

// @api(template/annotation) Annotation
// Annotation represents an annotation.
// If Name is empty, the Value is not nil.
//
// Example:
//
// ```
// @json(omitempty)
// @event(name="Login")
// @message(name="Login", type=100)
// ```
type Annotation map[string]*AnnotationParam

// @api(template/annotation) AnnotationGroup
// AnnotationGroup represents a group of annotations.
type AnnotationGroup map[string]Annotation

type AnnotationParam struct {
	pos   token.Pos
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
	return a.value.String()
}
