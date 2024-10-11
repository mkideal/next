package protocol

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gopherd/core/constraints"
	"github.com/gopherd/core/text/document"
)

//go:generate stringer -type=TokenType -linecomment

// TokenType constants aligned with VSCode's standard semantic token types
type TokenType uint32

const (
	TokenTypeText          TokenType = iota // text
	TokenTypeNamespace                      // namespace
	TokenTypeType                           // type
	TokenTypeClass                          // class
	TokenTypeEnum                           // enum
	TokenTypeInterface                      // interface
	TokenTypeStruct                         // struct
	TokenTypeTypeParameter                  // typeParameter
	TokenTypeParameter                      // parameter
	TokenTypeVariable                       // variable
	TokenTypeProperty                       // property
	TokenTypeEnumMember                     // enumMember
	TokenTypeEvent                          // event
	TokenTypeFunction                       // function
	TokenTypeMethod                         // method
	TokenTypeMacro                          // macro
	TokenTypeKeyword                        // keyword
	TokenTypeModifier                       // modifier
	TokenTypeComment                        // comment
	TokenTypeString                         // string
	TokenTypeNumber                         // number
	TokenTypeRegexp                         // regexp
	TokenTypeOperator                       // operator
	TokenTypeDecorator                      // decorator

	tokenTypeCount // -count-

	// token aliases
	TokenTypeConst               = TokenTypeEnumMember // const
	TokenTypeAnnotation          = TokenTypeRegexp     // annotation
	TokenTypeAnnotationParameter = TokenTypeMacro      // annotationParameter
)

// TokenTypes returns a list of token types.
func TokenTypes() []string {
	var types []string
	for i := TokenType(0); i < tokenTypeCount; i++ {
		types = append(types, i.String())
	}
	return types
}

//go:generate stringer -type=TokenModifier -linecomment
type TokenModifier int

const (
	TokenModifierDeclaration TokenModifier = iota // declaration

	tokenModifierCount // -count-
)

func TokenModifiers() []string {
	var modifiers []string
	for i := TokenModifier(0); i < tokenModifierCount; i++ {
		modifiers = append(modifiers, i.String())
	}
	return modifiers
}

// Token represents a token in a document.
type Token struct {
	Pos       uint32    // 0-based character offset
	Line      uint32    // 0-based line number
	Character uint32    // 0-based character offset
	Length    uint32    // length of the token
	TokenType TokenType // token type
	TokenMod  uint32    // token modifier
}

// String returns a string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("Token{l: %d, c: %d, s: %d, t: %d, m: %d, p: %d}",
		t.Line, t.Character, t.Length, t.TokenType, t.TokenMod, t.Pos)
}

// CompareToken compares two tokens.
func CompareToken(a, b Token) int {
	if a.Line != b.Line {
		return int(a.Line) - int(b.Line)
	}
	return int(a.Character) - int(b.Character)
}

// AddTokenHook is a hook that is called for each token added. If the hook returns true, the token is not added.
type AddTokenHook func(tokens *[]Token, doc *document.Document, lit string, pos int, tokenType TokenType) bool

// AddToken adds a token to the list of tokens. It splits the token by newlines and adds each line as a separate token.
// The hook is called for each token added. If the hook returns true, the token is not added.
func AddToken[Pos constraints.Integer](tokens *[]Token, doc *document.Document, lit string, position Pos, tokenType TokenType, hooks ...AddTokenHook) []Token {
	pos := int(position)
	trimmed := strings.TrimLeft(lit, " \t\r\n")
	pos += utf8.RuneCountInString(lit) - utf8.RuneCountInString(trimmed)
	lit = strings.TrimRight(trimmed, " \t\r\n")
	if strings.TrimSpace(lit) == "" {
		return nil
	}

	for _, hook := range hooks {
		if hook(tokens, doc, lit, pos, tokenType) {
			return nil
		}
	}

	lines := strings.Split(lit, "\n")
	currentPos := int(pos)

	line, column := PositionFor(currentPos, doc)

	start := len(*tokens)
	for i, lineContent := range lines {
		if strings.TrimSpace(lineContent) == "" {
			line++
			column = 0
			currentPos += utf8.RuneCountInString(lineContent) + 1
			continue
		}

		length := utf8.RuneCountInString(lineContent)

		slog.Debug(
			"Adding token",
			"type", tokenType,
			"line", line,
			"column", column,
			"text", lineContent,
		)

		*tokens = append(*tokens, Token{
			Line:      uint32(line),
			Character: uint32(column),
			Length:    uint32(length),
			TokenType: tokenType,
			Pos:       uint32(currentPos),
		})

		if i < len(lines)-1 {
			currentPos += length + 1
			line, column = PositionFor(currentPos, doc)
		}
	}
	return (*tokens)[start:]
}

// BuildTokens builds a list of sorted tokens from a list of tokens.
func BuildTokens(tokens []Token) []uint32 {
	slices.SortFunc(tokens, CompareToken)
	sortedTokens := make([]uint32, 0, len(tokens)*5)
	for i, t := range tokens {
		if i == 0 {
			sortedTokens = append(sortedTokens, t.Line, t.Character, t.Length, uint32(t.TokenType), t.TokenMod)
		} else if t.Line == tokens[i-1].Line {
			sortedTokens = append(sortedTokens, 0, t.Character-tokens[i-1].Character, t.Length, uint32(t.TokenType), t.TokenMod)
		} else {
			sortedTokens = append(sortedTokens, t.Line-tokens[i-1].Line, t.Character, t.Length, uint32(t.TokenType), t.TokenMod)
		}
	}
	return sortedTokens
}
