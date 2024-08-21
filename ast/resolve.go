package ast

import (
	"fmt"
	"strconv"

	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/token"
)

type pkgBuilder struct {
	fset   *token.FileSet
	errors scanner.ErrorList
}

func (p *pkgBuilder) error(pos token.Pos, msg string) {
	p.errors.Add(p.fset.Position(pos), msg)
}

func (p *pkgBuilder) errorf(pos token.Pos, format string, args ...any) {
	p.error(pos, fmt.Sprintf(format, args...))
}

// An Importer resolves import paths to package Objects.
// The imports map records the packages already imported,
// indexed by package id (canonical import path).
// An Importer must determine the canonical import path and
// check the map to see if it is already present in the imports map.
// If so, the Importer can return the map entry. Otherwise, the
// Importer should load the package data for the given path into
// a new *[Object] (pkg), record pkg in the imports map, and then
// return pkg.
//
// Deprecated: use the type checker [go/types] instead; see [Object].
type Importer func(imports map[string]*Object, path string) (pkg *Object, err error)

// NewPackage creates a new [Package] node from a set of [File] nodes. It resolves
// unresolved identifiers across files and updates each file's Unresolved list
// accordingly. If a non-nil importer and universe scope are provided, they are
// used to resolve identifiers not declared in any of the package files. Any
// remaining unresolved identifiers are reported as undeclared. If the files
// belong to different packages, one package name is selected and files with
// different package names are reported and then ignored.
// The result is a package node and a [scanner.ErrorList] if there were errors.
//
// Deprecated: use the type checker [go/types] instead; see [Object].
func NewPackage(fset *token.FileSet, files map[string]*File, importer Importer) (*Package, error) {
	var p pkgBuilder
	p.fset = fset

	// complete package scope
	pkgName := ""
	for _, file := range files {
		// package names must match
		switch name := file.Name.Name; {
		case pkgName == "":
			pkgName = name
		case name != pkgName:
			p.errorf(file.Package, "package %s; expected %s", name, pkgName)
			continue // ignore this file
		}
	}

	// package global mapping of imported package ids to package objects
	imports := make(map[string]*Object)

	// complete file scopes with imports and resolve identifiers
	for _, file := range files {
		// ignore file if it belongs to a different package
		// (error has already been reported)
		if file.Name.Name != pkgName {
			continue
		}

		// build file scope by processing all imports
		for _, spec := range file.Imports {
			if importer == nil {
				continue
			}
			path, _ := strconv.Unquote(spec.Path.Value)
			_, err := importer(imports, path)
			if err != nil {
				p.errorf(spec.Path.Pos(), "could not import %s (%s)", path, err)
				continue
			}
		}
	}

	p.errors.Sort()
	return &Package{pkgName, imports, files}, p.errors.Err()
}
