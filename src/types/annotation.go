package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"

	"github.com/mattn/go-shellwords"

	"github.com/next/next/api"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// Annotations represents a group of annotations.
// @api(template/annotation) Annotations
type Annotations map[string]Annotation

func (a Annotations) get(name string) Annotation {
	if a == nil {
		return nil
	}
	return a[name]
}

// Annotation represents an annotation.
//
// Example:
//
// ```
// @json(omitempty)
// @event(name="Login")
// @message(name="Login", type=100)
// ```
// @api(template/annotation) Annotation
type Annotation map[string]any

func (a Annotation) get(name string) any {
	if a == nil {
		return nil
	}
	return a[name]
}

// linkedAnnotation represents an annotation linked to a declaration.
type linkedAnnotation struct {
	name       string
	annotation Annotation

	// obj is declared as type `Object` to avoid circular dependency.
	// Actually, it's a `Node`.
	obj Object
}

func (c *Context) solveAnnotations() error {
	parser := shellwords.NewParser()
	parser.ParseEnv = true
	programs := make(map[string][][]string)
	keys := make([]string, 0, len(c.flags.solvers))
	for name, solvers := range c.flags.solvers {
		if name == "next" {
			return fmt.Errorf("'next' is a reserved annotation for the next compiler, please don't use solvers for it")
		}
		keys = append(keys, name)
		for _, solver := range solvers {
			words, err := parser.Parse(solver)
			if err != nil {
				return err
			}
			if len(words) == 0 {
				return fmt.Errorf("empty solver for %q", name)
			}
			programs[name] = append(programs[name], words)
		}
	}
	sort.Strings(keys)
	for _, name := range keys {
		for _, words := range programs[name] {
			req := c.createAnnotationSolverRequest(name)
			if len(req.Annotations) == 0 {
				continue
			}
			c.Printf("run solver %q", words)
			cmd := exec.Command(words[0], words[1:]...)
			var stdin bytes.Buffer
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdin = &stdin
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := json.NewEncoder(&stdin).Encode(req); err != nil {
				return fmt.Errorf("failed to encode request for solver %q: %v", name, err)
			}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run solver %q: %v\n%s", name, err, stderr.String())
			}
			var res api.AnnotationSolverResponse
			if err := json.NewDecoder(&stdout).Decode(&res); err != nil {
				return fmt.Errorf("failed to decode response for solver %q: %v", name, err)
			}
			for id, a := range res.Annotations {
				la, ok := c.annotations[token.Pos(id)]
				if !ok {
					return fmt.Errorf("solver %q: annotation %d not found", words, id)
				}
				for name, param := range a.Params {
					var v constant.Value
					if param.Value != nil {
						v = constant.Make(param.Value)
						if v.Kind() == constant.Unknown {
							return fmt.Errorf("solver %q: invalid value for parameter %q in annotation %q: %v", words, name, la.name, param.Value)
						}
						c.Printf("solver %q: set %q.%q to %v", words, la.name, name, param.Value)
					}
					_, ok := la.annotation[name]
					if !ok {
						if v == nil {
							return fmt.Errorf("solver %q: add an invalid value for parameter %q in annotation %q: %v", words, name, la.name, param.Value)
						}
						la.annotation[name] = constant.Underlying(v)
						continue
					}
					if v != nil {
						la.annotation[name] = constant.Underlying(v)
					} else {
						c.Printf("solver %q: remove %q.%q", words, la.name, name)
						delete(la.annotation, name)
					}
				}
			}
		}
	}
	return nil
}

func (c *Context) createAnnotationSolverRequest(name string) *api.AnnotationSolverRequest {
	req := &api.AnnotationSolverRequest{
		Objects:     make(map[api.ID]*api.Object),
		Annotations: make(map[api.ID]*api.Annotation),
	}
	var objects = make(map[token.Pos]Node)
	for pos, a := range c.annotations {
		if a.name != name {
			continue
		}
		var params = make(map[string]api.Parameter)
		for name, param := range a.annotation {
			params[name] = api.Parameter{
				Name:  name,
				Value: param,
			}
		}
		// a.obj is declared as type `Object` to avoid circular dependency.
		// Actually, it's a `Node`.
		obj := a.obj.(Node)
		req.Annotations[api.ID(pos)] = &api.Annotation{
			ID:     api.ID(pos),
			Object: api.ID(obj.getPos()),
			Params: params,
		}
		objects[obj.getPos()] = obj
	}
	for pos, obj := range objects {
		req.Objects[api.ID(pos)] = &api.Object{
			ID:   api.ID(pos),
			Type: obj.getType(),
			Name: obj.getName(),
			Pkg:  obj.File().pkg.name,
			File: obj.File().Path,
		}
	}
	return req
}
