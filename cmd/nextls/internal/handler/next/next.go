package next

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gopherd/core/text/document"
	"github.com/gopherd/core/types"

	"github.com/mkideal/next/api"
	"github.com/mkideal/next/cmd/nextls/internal/protocol"
	"github.com/mkideal/next/src/ast"
	"github.com/mkideal/next/src/compile"
	"github.com/mkideal/next/src/parser"
	"github.com/mkideal/next/src/token"
)

func init() {
	protocol.Register(".next", &nextHandler{})
}

type nextHandler struct {
	nextCompiler string
	env          api.NextEnv
}

func jsonCommand[T any](ctx context.Context, command []string) (*T, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run command %v: %v\n%s", command, err, stderr.String())
	}
	var result T
	if err := json.NewDecoder(&stdout).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v\n%s", err, stdout.String())
	}
	return &result, nil
}

func (h *nextHandler) Init() {
	h.nextCompiler = "next"
	if x := os.Getenv("NEXT_COMPILER"); x != "" {
		h.nextCompiler = x
	}
	env, err := jsonCommand[api.NextEnv](context.Background(), []string{h.nextCompiler, "env", "-json"})
	if err != nil {
		h.env.NextPath = filepath.Dir(h.nextCompiler)
		slog.Error("failed to get next environment", "error", err)
	} else {
		h.env = *env
	}
}

type documentsFS struct {
	documents protocol.DocumentManager
}

type documentsFile struct {
	content *strings.Reader
}

func (f *documentsFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (f *documentsFile) Read(p []byte) (int, error) { return f.content.Read(p) }
func (f *documentsFile) Close() error               { return nil }

func (fs documentsFS) Open(name string) (fs.File, error) {
	doc := fs.documents.GetDocument(name)
	if doc == nil {
		return os.Open(name)
	}
	return &documentsFile{content: strings.NewReader(doc.Content())}, nil
}

func (documentsFS) Abs(name string) (string, error) {
	return name, nil
}

func (h *nextHandler) HandleSemanticTokens(ctx context.Context, documents protocol.DocumentManager, doc *document.Document) (any, error) {
	context, _ := parseFile(documents, doc)
	file := context.GetFile(doc.Path())
	if file == nil {
		return nil, nil
	}
	node := compile.FileNode(file)
	visitor := &tokenVisiter{compiler: context, file: file, doc: doc}

	addToken(&visitor.tokens, doc, token.PACKAGE.String(), node.Package, protocol.TokenTypeKeyword)
	if node.Name != nil {
		addToken(&visitor.tokens, doc, node.Name.Name, node.Name.Pos(), protocol.TokenTypeNamespace)
	}
	ast.Walk(visitor, node.Annotations)
	for _, imp := range node.Imports {
		ast.Walk(visitor, imp)
	}
	for _, decl := range node.Decls {
		ast.Walk(visitor, decl)
	}
	for _, stmt := range node.Stmts {
		ast.Walk(visitor, stmt)
	}
	for i := range node.Comments {
		for _, c := range node.Comments[i].List {
			addToken(&visitor.tokens, doc, c.Text, c.Pos(), protocol.TokenTypeComment)
		}
	}

	tokens := protocol.BuildTokens(visitor.tokens)
	slog.Debug("semantic tokens", "tokens", tokens)
	return types.Object{"data": tokens}, nil
}

func parseFile(documents protocol.DocumentManager, doc *document.Document) (*compile.Compiler, error) {
	compiler := compile.NewCompiler(compile.StandardPlatform(), documentsFS{documents: documents})
	if f, _ := parser.ParseFile(compiler.FileSet(), doc.Path(), doc.Content(), parser.ParseComments); f != nil {
		compiler.AddFile(f)
	}
	err := compiler.Resolve()
	return compiler, err
}

type tokenVisiter struct {
	compiler *compile.Compiler
	file     *compile.File
	doc      *document.Document
	tokens   []protocol.Token
}

// isNil returns whether the underlying value is nil.
// This is only implemented for pointer, interface, slice, and map types.
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (v *tokenVisiter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if isNil(reflect.ValueOf(node)) {
		return nil
	}

	switch node := node.(type) {
	case *ast.Annotation:
		addToken(&v.tokens, v.doc, token.AT.String(), node.At, protocol.TokenTypeAnnotation)
		addToken(&v.tokens, v.doc, node.Name.Name, node.Name.Pos(), protocol.TokenTypeAnnotation)
		for _, p := range node.Params {
			ast.Walk(v, p)
		}
		return nil
	case *ast.AnnotationParam:
		addToken(&v.tokens, v.doc, node.Name.Name, node.Name.Pos(), protocol.TokenTypeAnnotationParameter)
		if node.AssignPos.IsValid() {
			addToken(&v.tokens, v.doc, token.ASSIGN.String(), node.AssignPos, protocol.TokenTypeOperator)
			if n := compile.LookupFileObject(v.file, node.Value); n != nil {
				if t, ok := n.(compile.Type); ok {
					addTypeTokens(v, t)
					return nil
				}
			}
			ast.Walk(v, node.Value)
		}
		return nil
	case *ast.ImportDecl:
		addToken(&v.tokens, v.doc, token.IMPORT.String(), node.ImportPos, protocol.TokenTypeKeyword)
		addToken(&v.tokens, v.doc, node.Path.Value, node.Path.Pos(), protocol.TokenTypeString)
		return nil
	case *ast.ConstDecl:
		addGenDeclTokens(v, node, protocol.TokenTypeConst)
		return nil
	case *ast.EnumDecl:
		addGenDeclTokens(v, node, protocol.TokenTypeEnum)
		return nil
	case *ast.StructDecl:
		addGenDeclTokens(v, node, protocol.TokenTypeStruct)
		return nil
	case *ast.InterfaceDecl:
		addGenDeclTokens(v, node, protocol.TokenTypeInterface)
		return nil
	case *ast.EnumType:
		for _, m := range node.Members.List {
			ast.Walk(v, m.Annotations)
			ast.Walk(v, m.Value)
			addToken(&v.tokens, v.doc, m.Name.Name, m.Name.Pos(), protocol.TokenTypeEnumMember)
		}
		return nil
	case *ast.StructType:
		for _, f := range node.Fields.List {
			ast.Walk(v, f.Annotations)
			if f.Type != nil {
				x := compile.LookupFileObject(v.file, f.Type)
				if x == nil {
					slog.Warn("type not found", "type", f.Type, "addr", fmt.Sprintf("%p", f.Type))
				} else if t, ok := x.(compile.Type); ok {
					addTypeTokens(v, t)
				} else {
					slog.Warn("unknown type", "type", f.Type, "addr", fmt.Sprintf("%p", f.Type), "node", fmt.Sprintf("%T", x))
				}
			}
			addToken(&v.tokens, v.doc, f.Name.Name, f.Name.Pos(), protocol.TokenTypeProperty)
		}
		return nil
	case *ast.InterfaceType:
		for _, m := range node.Methods.List {
			ast.Walk(v, m.Annotations)
			for _, p := range m.Params.List {
				ast.Walk(v, p.Annotations)
				if p.Type != nil {
					if t, ok := compile.LookupFileObject(v.file, p.Type).(compile.Type); ok {
						addTypeTokens(v, t)
					}
				}
				addToken(&v.tokens, v.doc, p.Name.Name, p.Name.Pos(), protocol.TokenTypeParameter)
			}
			if m.Result != nil {
				if t, ok := compile.LookupFileObject(v.file, m.Result).(compile.Type); ok {
					addTypeTokens(v, t)
				}
			}
			addToken(&v.tokens, v.doc, m.Name.Name, m.Name.Pos(), protocol.TokenTypeMethod)
		}
		return nil
	default:
		expr, ok := node.(ast.Expr)
		if !ok {
			return v
		}
		expr = ast.Unparen(expr)
		switch expr := expr.(type) {
		case *ast.Ident:
			if expr.Name == "iota" {
				addToken(&v.tokens, v.doc, expr.Name, expr.Pos(), protocol.TokenTypeKeyword)
			} else if expr.Name == "true" || expr.Name == "false" {
				addToken(&v.tokens, v.doc, expr.Name, expr.Pos(), protocol.TokenTypeConst)
			} else if s := compile.LookupSymbol(v.file, expr.Name); s != nil {
				if t, ok := s.(compile.Type); ok {
					addTypeTokens(v, t)
				} else {
					addToken(&v.tokens, v.doc, expr.Name, expr.Pos(), protocol.TokenTypeConst)
				}
			} else {
				addToken(&v.tokens, v.doc, expr.Name, expr.Pos(), protocol.TokenTypeConst)
			}
		case *ast.BasicLit:
			switch expr.Kind {
			case token.STRING:
				addToken(&v.tokens, v.doc, expr.Value, expr.Pos(), protocol.TokenTypeString)
			case token.INT, token.FLOAT:
				addToken(&v.tokens, v.doc, expr.Value, expr.Pos(), protocol.TokenTypeNumber)
			}
		case *ast.SelectorExpr:
			addSelectExprTokens(v, expr)
			return nil
		case *ast.CallExpr:
			fun := ast.Unparen(expr.Fun)
			addDefaultToken(v, fun, protocol.TokenTypeFunction)
			for _, arg := range expr.Args {
				ast.Walk(v, arg)
			}
			return nil
		case *ast.UnaryExpr:
		case *ast.BinaryExpr:
		}
		return v
	}
}

func addTypeTokens(v *tokenVisiter, x compile.Type) {
	t, ok := x.(*compile.UsedType)
	if !ok {
		slog.Warn("expected used type", "type", x)
		return
	}
	node := compile.UsedTypeNode(t)
	switch t2 := t.Type.(type) {
	case *compile.PrimitiveType:
		addToken(&v.tokens, v.doc, t.String(), node.Pos(), protocol.TokenTypeType)
	case *compile.ArrayType:
		addToken(&v.tokens, v.doc, token.ARRAY.String(), node.Pos(), protocol.TokenTypeType)
		addTypeTokens(v, t2.ElemType)
	case *compile.VectorType:
		addToken(&v.tokens, v.doc, token.VECTOR.String(), node.Pos(), protocol.TokenTypeType)
		addTypeTokens(v, t2.ElemType)
	case *compile.MapType:
		addToken(&v.tokens, v.doc, token.MAP.String(), node.Pos(), protocol.TokenTypeType)
		addTypeTokens(v, t2.KeyType)
		addTypeTokens(v, t2.ElemType)
	case *compile.DeclType[*compile.Enum]:
		addDeclTypeToken(v, t2, node, protocol.TokenTypeEnum)
	case *compile.DeclType[*compile.Struct]:
		addDeclTypeToken(v, t2, node, protocol.TokenTypeStruct)
	case *compile.DeclType[*compile.Interface]:
		addDeclTypeToken(v, t2, node, protocol.TokenTypeInterface)
	}
}

func addDeclTypeToken[T compile.Decl](v *tokenVisiter, _ *compile.DeclType[T], node ast.Type, tokenType protocol.TokenType) {
	if t, ok := node.(*ast.SelectorExpr); ok {
		addSelectExprTokens(v, t)
		return
	}
	addDefaultToken(v, node, tokenType)
}

func contentOf(doc *document.Document, node ast.Node) string {
	if node == nil {
		return ""
	}
	content := doc.Content()
	begin, end := int(node.Pos())-1, int(node.End())-1
	if begin > 0 && begin < end && end < len(content) {
		return content[begin:end]
	}
	return ""
}

func addSelectExprTokens(v *tokenVisiter, x *ast.SelectorExpr) {
	if val, _ := compile.LookupValue(v.file, contentOf(v.doc, x)); val != nil {
		if val.Enum() != nil {
			addDefaultToken(v, x.X, protocol.TokenTypeEnum)
			addDefaultToken(v, x.Sel, protocol.TokenTypeEnumMember)
		} else {
			addDefaultToken(v, x.X, protocol.TokenTypeVariable)
			addDefaultToken(v, x.Sel, protocol.TokenTypeConst)
		}
		return
	}
	if typ, _ := compile.LookupType(v.file, contentOf(v.doc, x)); typ != nil {
		addDefaultToken(v, x.X, protocol.TokenTypeVariable)
		addDefaultToken(v, x.Sel, protocol.TokenTypeType)
		return
	}

	addDefaultToken(v, x.X, protocol.TokenTypeVariable)
	addDefaultToken(v, x.Sel, protocol.TokenTypeConst)
}

func addDefaultToken(v *tokenVisiter, node ast.Node, tokenType protocol.TokenType) {
	if node == nil {
		return
	}
	if selector, ok := node.(*ast.SelectorExpr); ok {
		addSelectExprTokens(v, selector)
		return
	}
	if content := contentOf(v.doc, node); content != "" {
		addToken(&v.tokens, v.doc, content, node.Pos(), tokenType)
	}
}

func addGenDeclTokens[T ast.Node](v *tokenVisiter, decl *ast.GenDecl[T], tokenType protocol.TokenType) {
	ast.Walk(v, decl.Annotations)
	ast.Walk(v, decl.Spec)
	addToken(&v.tokens, v.doc, decl.Tok.String(), decl.TokPos, protocol.TokenTypeKeyword)
	addToken(&v.tokens, v.doc, decl.Name.Name, decl.Name.Pos(), tokenType)
}

func addToken(tokens *[]protocol.Token, doc *document.Document, lit string, pos token.Pos, tokenType protocol.TokenType) {
	if pos <= token.NoPos {
		return
	}
	newTokens := protocol.AddToken(tokens, doc, lit, pos-1, tokenType)
	if tokenType == protocol.TokenTypeKeyword {
		mod := keywordMod(lit)
		for i := range newTokens {
			newTokens[i].TokenMod = uint32(mod)
		}
	}
}

func keywordMod(keyword string) protocol.TokenModifier {
	switch keyword {
	case "package", "const", "enum", "struct", "interface":
		return 1 << protocol.TokenModifierDeclaration
	default:
		return 0
	}
}
