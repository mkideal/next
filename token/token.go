// Package token defines constants representing the lexical tokens of the Next
// programming language and basic operations on tokens (printing, predicates).
package token

import (
	"strconv"
	"strings"
	"unicode"
)

// Token is the set of lexical tokens of the Next programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT
	FILE // special token to represent a file

	literal_beg
	// Identifiers and basic type literals
	IDENT  // main
	INT    // 12345
	FLOAT  // 123.45
	STRING // "abc"
	CHAR   // 'a'
	literal_end

	operator_beg
	// Operators and delimiters
	AT  // @
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND  // &
	OR   // |
	XOR  // ^
	SHL  // <<
	SHR  // >>
	LAND // &&
	LOR  // ||

	EQL     // ==
	LSS     // <
	GTR     // >
	ASSIGN  // =
	NOT     // !
	AND_NOT // &^

	NEQ // !=
	LEQ // <=
	GEQ // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	operator_end

	keyword_beg
	// Keywords
	PACKAGE
	IMPORT
	CONST
	ENUM
	STRUCT
	PROTOCOL
	MAP
	VECTOR
	ARRAY
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",
	FILE:    "FILE",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",
	CHAR:   "CHAR",

	AT:  "@",
	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND:  "&",
	OR:   "|",
	XOR:  "^",
	SHL:  "<<",
	SHR:  ">>",
	LAND: "&&",
	LOR:  "||",

	EQL:     "==",
	LSS:     "<",
	GTR:     ">",
	ASSIGN:  "=",
	NOT:     "!",
	AND_NOT: "&^",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",

	PACKAGE:  "package",
	IMPORT:   "import",
	CONST:    "const",
	ENUM:     "enum",
	STRUCT:   "struct",
	PROTOCOL: "protocol",
	MAP:      "map",
	VECTOR:   "vector",
	ARRAY:    "array",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ, ASSIGN:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, AND, SHL, SHR:
		return 5
	}
	return LowestPrec
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keyword_end-(keyword_beg+1))
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

// IsExported reports whether name does not start with underscore "_".
func IsExported(name string) bool {
	return !strings.HasPrefix(name, "_")
}

// IsKeyword reports whether name is a Go keyword, such as "func" or "return".
func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

// IsIdentifier reports whether name is a Go identifier, that is, a non-empty
// string made up of letters, digits, and underscores, where the first character
// is not a digit. Keywords are not identifiers.
func IsIdentifier(name string) bool {
	if name == "" || IsKeyword(name) {
		return false
	}
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}
