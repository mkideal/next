package api

import (
	"encoding/json"
	"os"

	"github.com/gopherd/core/errkit"
)

const (
	StatusBadRequest    = 100
	StatusInternalError = 101
)

// ID represents an identifier.
type ID = int

type ObjectType = string

const (
	ObjectFile            ObjectType = "file"
	ObjectConstDecl       ObjectType = "const.decl"
	ObjectConst           ObjectType = "const"
	ObjectEnumDecl        ObjectType = "enum.decl"
	ObjectEnum            ObjectType = "enum"
	ObjectEnumMember      ObjectType = "enum.member"
	ObjectStructDecl      ObjectType = "struct.decl"
	ObjectStruct          ObjectType = "struct"
	ObjectStructField     ObjectType = "struct.field"
	ObjectInterfaceDecl   ObjectType = "interface.decl"
	ObjectInterface       ObjectType = "interface"
	ObjectInterfaceMethod ObjectType = "interface.method"
)

// Object represents an object which may be annotated.
type Object struct {
	ID   ID         `json:"id"`
	Name string     `json:"name"`
	Type ObjectType `json:"type"`
	Pkg  string     `json:"pkg"`
	File string     `json:"file"`
}

// Annotation represents an annotation.
type Annotation struct {
	ID     ID                   `json:"id"`
	Object ID                   `json:"object"`
	Params map[string]Parameter `json:"params"`
}

// Parameter represents an annotation parameter.
type Parameter struct {
	// Name of the parameter.
	Name string `json:"name"`
	// Type of the value may be: string, int64, uint64, float32, float64, bool, or nil.
	Value any `json:"value"`
}

// AnnotationSolverRequest represents a request to solve annotations.
type AnnotationSolverRequest struct {
	Objects     map[ID]*Object     `json:"objects"`
	Annotations map[ID]*Annotation `json:"annotations"`
}

// AnnotationSolverResponse represents a response to solve annotations.
type AnnotationSolverResponse struct {
	Annotations map[ID]*Annotation `json:"annotations"`
}

// SolveAnnotations solves annotations and writes the result to stdout and exits.
func SolveAnnotations(solver func(*AnnotationSolverRequest) (*AnnotationSolverResponse, error)) {
	var req AnnotationSolverRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(StatusBadRequest)
	}
	res, err := solver(&req)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		if code, ok := errkit.ExitCode(err); ok {
			os.Exit(code)
		}
		os.Exit(StatusInternalError)
	}
	if err := json.NewEncoder(os.Stdout).Encode(res); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(StatusInternalError)
	}
	os.Exit(0)
}

// NextEnv represents the environment of the next compiler.
type NextEnv struct {
	NextPath   string   `json:"next_path"`
	ImportDirs []string `json:"import_dirs"`
}
