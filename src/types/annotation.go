package types

import (
	"github.com/next/next/api"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// @api(template/annotation) AnnotationGroup
// AnnotationGroup represents a group of annotations.
type AnnotationGroup map[string]Annotation

// @api(template/annotation) Annotation
// Annotation represents an annotation.
//
// Example:
//
// ```
// @json(omitempty)
// @event(name="Login")
// @message(name="Login", type=100)
// ```
type Annotation map[string]*AnnotationParam

// @api(template/annotation) AnnotationParam
// AnnotationParam represents an annotation parameter.
type AnnotationParam struct {
	pos   token.Pos
	name  string
	value constant.Value
}

// @api(template/annotation) Name
// Name returns the name of the annotation parameter.
func (a *AnnotationParam) Name() string {
	return a.name
}

// @api(template/annotation) Value
// Value returns the value of the annotation parameter.
func (a *AnnotationParam) Value() any {
	if a == nil {
		return nil
	}
	switch a.value.Kind() {
	case constant.String:
		return constant.StringVal(a.value)
	case constant.Int:
		if i, exactly := constant.Int64Val(a.value); exactly {
			return i
		}
		u, _ := constant.Uint64Val(a.value)
		return u
	case constant.Float:
		if f, exactly := constant.Float32Val(a.value); exactly {
			return f
		}
		f, _ := constant.Float64Val(a.value)
		return f
	case constant.Bool:
		return constant.BoolVal(a.value)
	default:
		return nil
	}
}

// @api(template/annotation) String
// String returns the string representation of the annotation parameter.
func (a *AnnotationParam) String() string {
	if a == nil {
		return ""
	}
	return a.value.String()
}

type linkedAnnotation struct {
	name       string
	object     Object
	annotation Annotation
}

func (c *Context) solveAnnotations() error {
	// TODO
	//for name, solvers := range c.flags.solvers {
	//	for _, solver := range solvers {
	//		req := c.creatAnnotationSolverRequest(name)
	//		if err := solver(c); err != nil {
	//			return err
	//		}
	//	}
	//}
	return nil
}

func (c *Context) creatAnnotationSolverRequest(name string) *api.AnnotationSolverRequest {
	req := &api.AnnotationSolverRequest{
		Objects:     make(map[api.ID]*api.Object),
		Annotations: make(map[api.ID]*api.Annotation),
	}
	var objects = make(map[token.Pos]Object)
	for pos, a := range c.annotations {
		if a.name != name {
			continue
		}
		var params = make(map[string]api.Parameter)
		for name, param := range a.annotation {
			params[name] = api.Parameter{
				Name:  param.name,
				Value: param.Value(),
			}
		}
		req.Annotations[api.ID(pos)] = &api.Annotation{
			ID:     api.ID(pos),
			Object: api.ID(a.object.getPos()),
			Params: params,
		}
		objects[a.object.getPos()] = a.object
	}
	for pos, obj := range objects {
		req.Objects[api.ID(pos)] = &api.Object{
			ID:   api.ID(pos),
			Type: obj.getType(),
			Name: obj.getName(),
			//Pkg:  obj.File().pkg.name,
			//File: obj.File().Path,
		}
	}
	return req
}
