package npl

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"

	"github.com/gopherd/core/text/document"
	"github.com/gopherd/core/types"
	"github.com/gopherd/template/parse"

	"github.com/mkideal/next/cmd/nextls/internal/protocol"
)

func init() {
	protocol.Register(".npl", &nplHandler{})
	protocol.Register(".tmpl", &nplHandler{})
	protocol.Register(".gotmpl", &nplHandler{})
}

func b(s string) []byte { return []byte(s) }

const (
	keywordDefine  = "define"
	keywordEnd     = "end"
	keywordNext    = "next"
	leftDelimiter  = "{{"
	rightDelimiter = "}}"
)

var keywords = map[string]bool{
	"eq":  true,
	"ne":  true,
	"lt":  true,
	"le":  true,
	"gt":  true,
	"ge":  true,
	"and": true,
	"or":  true,
	"not": true,
}

// nplHandler handles LSP requests for the NPL file
type nplHandler struct {
}

// HandleSemanticTokens handles a "textDocument/semanticTokens/full" request.
func (h *nplHandler) HandleSemanticTokens(ctx context.Context, _ protocol.DocumentManager, doc *document.Document) (any, error) {
	filename := doc.URI()
	content := doc.Content()
	if len(content) == 0 {
		return nil, nil
	}

	root := parseTemplate(filename, content)

	var tokens []protocol.Token
	walkTemplateTokens(root, doc, &tokens)

	// sort tokens by position for adding new special tokens between them
	slices.SortFunc(tokens, protocol.CompareToken)

	var newTokens []protocol.Token
	begin := uint32(0)
	end := uint32(0)
	for i := -1; i < len(tokens); i++ {
		if i == -1 {
			if len(tokens) > 0 {
				end = tokens[0].Pos
			} else {
				end = uint32(len(content))
			}
		} else {
			begin = tokens[i].Pos + tokens[i].Length
			end = uint32(len(content))
			if i < len(tokens)-1 {
				end = tokens[i+1].Pos
			}
		}

		// Search special tokens in [begin, end): "(", ")", "|", "{{", "}}", "-", "end"
		for index := begin; index < end; index++ {
			if content[index] == '(' || content[index] == ')' {
				addToken(&newTokens, doc, string(content[index]), parse.Pos(index), protocol.TokenTypeRegexp)
			} else if content[index] == '|' {
				addToken(&newTokens, doc, string(content[index]), parse.Pos(index), protocol.TokenTypeRegexp)
			} else if content[index] == '{' && index+1 < end && content[index+1] == '{' {
				addToken(&newTokens, doc, leftDelimiter, parse.Pos(index), protocol.TokenTypeRegexp)
				index++
			} else if content[index] == '}' && index+1 < end && content[index+1] == '}' {
				addToken(&newTokens, doc, rightDelimiter, parse.Pos(index), protocol.TokenTypeRegexp)
				index++
			} else if content[index] == '-' {
				addToken(&newTokens, doc, "-", parse.Pos(index), protocol.TokenTypeRegexp)
			} else if content[index] == 'e' && index+3 < end && content[index+1] == 'n' && content[index+2] == 'd' && !isLetter(content[index+3]) {
				addToken(&newTokens, doc, keywordEnd, parse.Pos(index), protocol.TokenTypeKeyword)
				index += 2
			}
		}
	}

	return types.Object{"data": protocol.BuildTokens(append(tokens, newTokens...))}, nil
}

func parseTemplate(filename, text string) parse.Node {
	trees := make(map[string]*parse.Tree)
	tree := parse.New(filename)
	tree.Mode = parse.ParseComments | parse.SkipFuncCheck | parse.ContinueAfterError
	tree.Parse(text, "", "", trees)
	root := tree.Root
	slog.Debug("Parsed template", "filename", filename, "trees", len(trees), "nodes", len(root.Nodes))

	var nodes []*parse.Tree
	for _, t := range trees {
		nodes = append(nodes, t)
	}
	var seen = make(map[parse.Node]bool)
	for _, n := range root.Nodes {
		seen[n] = true
	}
	for _, n := range nodes {
		if seen[n.Root] {
			continue
		}
		if n.Root != root {
			root.Nodes = append(root.Nodes, n.Root)
		}
		if n.Define != nil && !seen[n.Define] {
			root.Nodes = append(root.Nodes, n.Define)
		}
	}
	sort.Slice(root.Nodes, func(i, j int) bool {
		return root.Nodes[i].Position() < root.Nodes[j].Position()
	})
	slog.Debug("Parsed template", "filename", filename, "nodes", len(root.Nodes))
	return root
}

const nopos parse.Pos = -1

func walkTemplateTokens(node parse.Node, doc *document.Document, tokens *[]protocol.Token) {
	line, column := protocol.PositionFor(int(node.Position()), doc)
	slog.Debug("Walking template node", "type", fmt.Sprintf("%T", node), "text", node.String(), "line", line, "column", column)
	content := doc.Content()
	switch n := node.(type) {
	case *parse.EndNode:
		addToken(tokens, doc, "end", n.Pos, protocol.TokenTypeKeyword)
	case *parse.ActionNode:
		walkTemplateTokens(n.Pipe, doc, tokens)

	case *parse.BoolNode:
		lit := "false"
		if n.True {
			lit = "true"
		}
		addToken(tokens, doc, lit, n.Pos, protocol.TokenTypeEnumMember)

	case *parse.ChainNode:
		handleChainNode(n, doc, tokens)

	case *parse.CommandNode:
		rendered := false
		if len(n.Args) >= 2 && n.Args[0].String() == "render" {
			// special case for render command: {{render "<name>" .}}
			if arg, ok := n.Args[1].(*parse.StringNode); ok {
				rendered = parseTemplateName(arg.Quoted, arg.Text).addTokens(tokens, doc, arg.Pos) > 0
				if rendered {
					walkTemplateTokens(n.Args[0], doc, tokens)
					for i := 2; i < len(n.Args); i++ {
						walkTemplateTokens(n.Args[i], doc, tokens)
					}
				}
			}
		}
		if !rendered {
			for i := range n.Args {
				walkTemplateTokens(n.Args[i], doc, tokens)
			}
		}

	case *parse.FieldNode:
		pos := n.Pos
		if n.Pos > 0 && (isLetter(content[n.Pos-1]) || isDigit(content[n.Pos-1]) || content[n.Pos-1] == '_') {
			pos = last(content, n.Pos, ".")
		}
		addToken(tokens, doc, n.String(), pos, protocol.TokenTypeProperty)

	case *parse.CommentNode:
		addToken(tokens, doc, n.Text, n.Pos, protocol.TokenTypeComment)

	case *parse.DotNode:
		addToken(tokens, doc, ".", n.Pos, protocol.TokenTypeKeyword)

	case *parse.IdentifierNode:
		if keywords[n.Ident] {
			addToken(tokens, doc, n.Ident, n.Pos, protocol.TokenTypeKeyword)
		} else {
			addToken(tokens, doc, n.Ident, n.Pos, protocol.TokenTypeFunction)
		}

	case *parse.IfNode:
		addToken(tokens, doc, "if", last(content, n.Pos, "if"), protocol.TokenTypeKeyword)
		walkTemplateTokens(n.Pipe, doc, tokens)
		walkTemplateTokens(n.List, doc, tokens)
		if n.ElseList != nil {
			// search for "else" keyword token in [n.ElseList]
			pos := last(content, n.ElseList.Pos, "else")
			if pos >= 0 {
				addToken(tokens, doc, "else", pos, protocol.TokenTypeKeyword)
			}
			walkTemplateTokens(n.ElseList, doc, tokens)
		}

	case *parse.ListNode:
		for _, child := range n.Nodes {
			walkTemplateTokens(child, doc, tokens)
		}

	case *parse.NilNode:
		addToken(tokens, doc, "nil", n.Pos, protocol.TokenTypeEnumMember)

	case *parse.NumberNode:
		addToken(tokens, doc, n.Text, n.Pos, protocol.TokenTypeNumber)

	case *parse.PipeNode:
		for _, decl := range n.Decl {
			walkTemplateTokens(decl, doc, tokens)
		}
		for _, cmd := range n.Cmds {
			walkTemplateTokens(cmd, doc, tokens)
		}

	case *parse.RangeNode:
		addToken(tokens, doc, "range", last(content, n.Pos, "range"), protocol.TokenTypeKeyword)
		walkTemplateTokens(n.Pipe, doc, tokens)
		walkTemplateTokens(n.List, doc, tokens)
		if n.ElseList != nil {
			if pos := last(content, n.ElseList.Pos, "else"); pos >= 0 {
				addToken(tokens, doc, "else", pos, protocol.TokenTypeKeyword)
			}
			walkTemplateTokens(n.ElseList, doc, tokens)
		}

	case *parse.BreakNode:
		addToken(tokens, doc, "break", n.Pos, protocol.TokenTypeKeyword)

	case *parse.ContinueNode:
		addToken(tokens, doc, "continue", n.Pos, protocol.TokenTypeKeyword)

	case *parse.StringNode:
		addToken(tokens, doc, n.Quoted, n.Pos, protocol.TokenTypeString)

	case *parse.TemplateNode:
		if pos := last(content, n.Pos, n.Keyword); pos >= 0 {
			addToken(tokens, doc, n.Keyword, pos, protocol.TokenTypeKeyword)
		}
		addToken(tokens, doc, n.Quoted, n.Pos, protocol.TokenTypeString)
		if n.Pipe != nil {
			walkTemplateTokens(n.Pipe, doc, tokens)
		}

	case *parse.DefineNode:
		if pos := last(content, n.Pos, keywordDefine); pos >= 0 {
			addToken(tokens, doc, keywordDefine, pos, protocol.TokenTypeKeyword)
		}
		if doc.Extension() == ".npl" {
			parseTemplateName(n.Quoted, n.Name).addTokens(tokens, doc, n.Pos)
		} else {
			addToken(tokens, doc, n.Quoted, n.Pos, protocol.TokenTypeString)
		}

	case *parse.TextNode:
		addToken(tokens, doc, string(n.Text), n.Pos, protocol.TokenTypeText)

	case *parse.VariableNode:
		s := n.String()
		pos := n.Pos
		for pos > 0 && (isLetter(content[pos-1]) || isDigit(content[pos-1]) || content[pos-1] == '_' || content[pos-1] == '.' || content[pos-1] == '$') {
			pos--
		}
		addToken(tokens, doc, s, pos, protocol.TokenTypeVariable)

	case *parse.WithNode:
		addToken(tokens, doc, "with", last(content, n.Pos, "with"), protocol.TokenTypeKeyword)
		walkTemplateTokens(n.Pipe, doc, tokens)
		walkTemplateTokens(n.List, doc, tokens)
		if n.ElseList != nil {
			if pos := last(content, n.ElseList.Pos, "else"); pos >= 0 {
				addToken(tokens, doc, "else", pos, protocol.TokenTypeKeyword)
			}
			walkTemplateTokens(n.ElseList, doc, tokens)
		}

	default:
		slog.Error("Unhandled node type", "type", fmt.Sprintf("%T", n))
	}
}

func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// last find last index of pattern in text before pos
func last(text string, pos parse.Pos, pattern string) parse.Pos {
	i := strings.LastIndex(text[:pos], pattern)
	if i == -1 {
		return -1
	}
	return parse.Pos(i)
}

func handleChainNode(n *parse.ChainNode, doc *document.Document, tokens *[]protocol.Token) {
	walkTemplateTokens(n.Node, doc, tokens)
	addToken(tokens, doc, "."+strings.Join(n.Field, "."), n.Pos, protocol.TokenTypeProperty)
}

// special case for ".X.Y.Z" chain
func addTokenSpecial(tokens *[]protocol.Token, doc *document.Document, lit string, pos int, tokenType protocol.TokenType) bool {
	if strings.Contains(lit, "\n") || lit[0] == '"' || lit[0] == '`' || lit[0] == '\'' || !strings.Contains(lit, ".") {
		return false
	}
	begin := 0
	currentPos := int(pos)
	line, column := protocol.PositionFor(currentPos, doc)
	runes := []rune(lit)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '.' {
			if begin < i {
				*tokens = append(*tokens, protocol.Token{
					Line:      uint32(line),
					Character: uint32(column + begin),
					Length:    uint32(i - begin),
					TokenType: tokenType,
					Pos:       uint32(currentPos + begin),
				})
			}
			begin = i + 1
			dotType := protocol.TokenTypeKeyword
			if i > 0 || tokenType == protocol.TokenTypeText {
				dotType = tokenType
			}
			*tokens = append(*tokens, protocol.Token{
				Line:      uint32(line),
				Character: uint32(column + i),
				Length:    1, // "."
				TokenType: dotType,
				Pos:       uint32(currentPos + i),
			})
		}
	}
	if begin < len(runes) {
		*tokens = append(*tokens, protocol.Token{
			Line:      uint32(line),
			Character: uint32(column + begin),
			Length:    uint32(len(runes) - begin),
			TokenType: tokenType,
			Pos:       uint32(currentPos + begin),
		})
	}
	return true
}

func addToken(tokens *[]protocol.Token, doc *document.Document, lit string, pos parse.Pos, tokenType protocol.TokenType) {
	if pos < 0 {
		return
	}
	protocol.AddToken(tokens, doc, lit, int(pos), tokenType, addTokenSpecial)
}

type templateName struct {
	quoted, name    string
	nextPos         parse.Pos // index of "next" in name
	langPos         parse.Pos // index of <lang> in name
	nodePos         parse.Pos // index of <node> in name
	funcPos         parse.Pos // index of <fun> in name
	lang, node, fun string
}

func (t templateName) addTokens(tokens *[]protocol.Token, doc *document.Document, pos parse.Pos) (n int) {
	if t.nextPos != -1 {
		addToken(tokens, doc, keywordNext, pos+t.nextPos, protocol.TokenTypeMacro)
		offset := parse.Pos(len(keywordNext))
		addToken(tokens, doc, "/", pos+t.nextPos+offset, protocol.TokenTypeRegexp)
		n++
	}
	if t.lang != "" {
		addToken(tokens, doc, t.lang, pos+t.langPos, protocol.TokenTypeEnumMember)
		offset := parse.Pos(len(t.lang))
		addToken(tokens, doc, "/", pos+t.langPos+offset, protocol.TokenTypeRegexp)
		n++
	}
	if t.node != "" {
		addToken(tokens, doc, t.node, pos+t.nodePos, protocol.TokenTypeType)
		n++
		if t.fun != "" {
			addToken(tokens, doc, ":", pos+t.nodePos+parse.Pos(len(t.node)), protocol.TokenTypeRegexp)
			addToken(tokens, doc, t.fun, pos+t.funcPos, protocol.TokenTypeNumber)
			n++
		}
	}
	if n == 0 {
		addToken(tokens, doc, t.quoted, pos, protocol.TokenTypeString)
	} else {
		addToken(tokens, doc, "\"", pos, protocol.TokenTypeRegexp)
		addToken(tokens, doc, "\"", pos+parse.Pos(len(t.quoted))-1, protocol.TokenTypeRegexp)
	}
	return
}

// template name format: [next/][<lang>/]<node>[:<fun>]
func parseTemplateName(quoted, name string) templateName {
	t := templateName{
		quoted:  quoted,
		name:    name,
		nextPos: -1,
	}
	const ns = "/"
	parts := strings.Split(name, ns)
	switch len(parts) {
	case 1:
		// <node>
		t.node = parts[0]
		t.nodePos = parse.Pos(1) // +1 for '"'
	case 2:
		if parts[0] == keywordNext {
			// next/<node>
			t.nextPos = parse.Pos(1) // +1 for '"'
			t.node = parts[1]
			t.nodePos = t.nextPos + parse.Pos(len(keywordNext)+1) // +1 for '/'
		} else {
			// <lang>/<node>
			t.lang = parts[0]
			t.langPos = parse.Pos(1) // +1 for '"'
			t.node = parts[1]
			t.nodePos = t.langPos + parse.Pos(len(t.lang)+1) // +1 for '/'
		}
	case 3:
		if parts[0] != keywordNext {
			return t
		}
		t.nextPos = parse.Pos(1) // +1 for '"'
		t.lang = parts[1]
		t.langPos = t.nextPos + parse.Pos(len(keywordNext)+1) // +1 for '/'
		t.node = parts[2]
		t.nodePos = t.langPos + parse.Pos(len(t.lang)+1) // +1 for '/'
	}
	if i := strings.Index(t.node, ":"); i >= 0 {
		t.fun = t.node[i+1:]
		t.node = t.node[:i]
		t.funcPos = t.nodePos + parse.Pos(i+1) // +1 for ':'
	}
	return t
}
