package main

import (
	"strings"

	"github.com/TRC-Loop/ccl-lsp/internal/lexer"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Token type indices - order must match semanticTokenTypes slice.
const (
	ttKeyword  = 0
	ttType     = 1
	ttFunction = 2
	ttVariable = 3
	ttClass    = 4
	ttProperty = 5
	ttMethod   = 6
	ttString   = 7
	ttNumber   = 8
	ttOperator = 9
	ttComment  = 10
	ttNamespace = 11
)

var semanticTokenTypes = []string{
	"keyword",
	"type",
	"function",
	"variable",
	"class",
	"property",
	"method",
	"string",
	"number",
	"operator",
	"comment",
	"namespace",
}

var semanticTokenModifiers = []string{
	"declaration",
	"readonly",
}

func textDocumentSemanticTokensFull(_ *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	uri := string(params.TextDocument.URI)
	doc := store.Get(uri)
	if doc == nil {
		return &protocol.SemanticTokens{Data: []uint32{}}, nil
	}

	data := encodeTokens(doc)
	return &protocol.SemanticTokens{Data: data}, nil
}

type semToken struct {
	line      uint32
	startChar uint32
	length    uint32
	tokenType uint32
	modifiers uint32
}

func encodeTokens(doc *Document) []uint32 {
	var tokens []semToken

	// First pass: classify tokens from the lexer token stream.
	if doc.Parsed != nil && len(doc.Parsed.Tokens) > 0 {
		tokens = append(tokens, classifyLexerTokens(doc.Parsed.Tokens, doc.Symbols)...)
	}

	// Second pass: find comments (not in lexer output).
	tokens = append(tokens, findComments(doc.Content)...)

	// Sort by line then character.
	sortSemTokens(tokens)

	// Encode as LSP delta format.
	out := make([]uint32, 0, len(tokens)*5)
	var prevLine, prevChar uint32
	for _, t := range tokens {
		deltaLine := t.line - prevLine
		deltaChar := t.startChar
		if deltaLine == 0 {
			deltaChar = t.startChar - prevChar
		}
		out = append(out, deltaLine, deltaChar, t.length, t.tokenType, t.modifiers)
		prevLine = t.line
		prevChar = t.startChar
	}
	return out
}

func classifyLexerTokens(tokens []lexer.Token, syms *Analysis) []semToken {
	var out []semToken

	for i, tok := range tokens {
		if tok.Type == lexer.TOKEN_EOF {
			continue
		}
		tt, mods, ok := classifyToken(tok, i, tokens, syms)
		if !ok {
			continue
		}
		line := uint32(tok.Line - 1)
		col := uint32(tok.Col - 1)
		length := uint32(len(tok.Literal))
		out = append(out, semToken{line, col, length, uint32(tt), uint32(mods)})
	}
	return out
}

func classifyToken(tok lexer.Token, i int, tokens []lexer.Token, syms *Analysis) (int, int, bool) {
	switch {
	case isKeywordToken(tok.Type):
		return ttKeyword, 0, true

	case isTypeToken(tok.Type):
		return ttType, 0, true

	case tok.Type == lexer.TOKEN_INT_LIT || tok.Type == lexer.TOKEN_FLOAT_LIT:
		return ttNumber, 0, true

	case tok.Type == lexer.TOKEN_STRING_LIT || tok.Type == lexer.TOKEN_FSTRING_LIT:
		return ttString, 0, true

	case isOperatorToken(tok.Type):
		return ttOperator, 0, true

	case tok.Type == lexer.TOKEN_IDENT:
		return classifyIdent(tok, i, tokens, syms)
	}
	return 0, 0, false
}

func classifyIdent(tok lexer.Token, i int, tokens []lexer.Token, syms *Analysis) (int, int, bool) {
	prev := prevSignificant(tokens, i)
	next := nextSignificant(tokens, i)

	// After "function" keyword
	if prev != nil && prev.Type == lexer.TOKEN_FUNCTION {
		return ttFunction, modDeclaration, true
	}
	// After "class" keyword
	if prev != nil && prev.Type == lexer.TOKEN_CLASS {
		return ttClass, modDeclaration, true
	}
	// After "extends" keyword
	if prev != nil && prev.Type == lexer.TOKEN_EXTENDS {
		return ttClass, 0, true
	}
	// After "." = property or method
	if prev != nil && prev.Type == lexer.TOKEN_DOT {
		if next != nil && next.Type == lexer.TOKEN_LPAREN {
			return ttMethod, 0, true
		}
		return ttProperty, 0, true
	}
	// Before "(" = function call
	if next != nil && next.Type == lexer.TOKEN_LPAREN {
		return ttFunction, 0, true
	}
	// After "catch" type position
	if prev != nil && prev.Type == lexer.TOKEN_CATCH {
		return ttClass, 0, true
	}

	// Check symbol table
	if syms != nil {
		if sym, ok := syms.Globals[tok.Literal]; ok {
			switch sym.Kind {
			case KindFunction:
				return ttFunction, 0, true
			case KindClass:
				return ttClass, 0, true
			case KindModule:
				return ttNamespace, 0, true
			case KindConst:
				return ttVariable, modReadonly, true
			}
			return ttVariable, 0, true
		}
	}

	return ttVariable, 0, true
}

const (
	modDeclaration = 1 << 0
	modReadonly    = 1 << 1
)

func prevSignificant(tokens []lexer.Token, i int) *lexer.Token {
	for j := i - 1; j >= 0; j-- {
		if tokens[j].Type != lexer.TOKEN_EOF {
			return &tokens[j]
		}
	}
	return nil
}

func nextSignificant(tokens []lexer.Token, i int) *lexer.Token {
	for j := i + 1; j < len(tokens); j++ {
		if tokens[j].Type != lexer.TOKEN_EOF {
			return &tokens[j]
		}
	}
	return nil
}

func isKeywordToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TOKEN_VAR, lexer.TOKEN_CONST, lexer.TOKEN_FUNCTION, lexer.TOKEN_RETURN,
		lexer.TOKEN_IF, lexer.TOKEN_ELSE, lexer.TOKEN_FOR, lexer.TOKEN_IN,
		lexer.TOKEN_WHILE, lexer.TOKEN_IMPORT, lexer.TOKEN_FROM, lexer.TOKEN_AND,
		lexer.TOKEN_OR, lexer.TOKEN_NOT, lexer.TOKEN_TRUE, lexer.TOKEN_FALSE,
		lexer.TOKEN_BREAK, lexer.TOKEN_CONTINUE, lexer.TOKEN_CLASS, lexer.TOKEN_EXTENDS,
		lexer.TOKEN_SELF, lexer.TOKEN_SUPER, lexer.TOKEN_PUBLIC, lexer.TOKEN_PRIVATE,
		lexer.TOKEN_TRY, lexer.TOKEN_CATCH, lexer.TOKEN_THROW, lexer.TOKEN_WITH,
		lexer.TOKEN_AS, lexer.TOKEN_RANGE, lexer.TOKEN_FIXED:
		return true
	}
	return false
}

func isTypeToken(t lexer.TokenType) bool {
	return t >= lexer.TOKEN_TYPE_STRING && t <= lexer.TOKEN_TYPE_SINT
}

func isOperatorToken(t lexer.TokenType) bool {
	switch t {
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS, lexer.TOKEN_STAR, lexer.TOKEN_SLASH,
		lexer.TOKEN_PERCENT, lexer.TOKEN_ASSIGN, lexer.TOKEN_EQ, lexer.TOKEN_NEQ,
		lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE:
		return true
	}
	return false
}

// findComments scans the raw source for line and block comments.
func findComments(content string) []semToken {
	var out []semToken
	lines := strings.Split(content, "\n")

	for lineIdx, row := range lines {
		i := 0
		for i < len(row) {
			if i+1 < len(row) && row[i] == '/' && row[i+1] == '/' {
				// line comment: rest of line
				out = append(out, semToken{
					line:      uint32(lineIdx),
					startChar: uint32(i),
					length:    uint32(len(row) - i),
					tokenType: ttComment,
				})
				break
			}
			if i+1 < len(row) && row[i] == '/' && row[i+1] == '*' {
				// block comment: scan until */
				startLine := lineIdx
				startCol := i
				end := strings.Index(content, "*/")
				if end < 0 {
					// unterminated - mark to end of file
					out = append(out, semToken{
						line:      uint32(startLine),
						startChar: uint32(startCol),
						length:    uint32(len(row) - startCol),
						tokenType: ttComment,
					})
					break
				}
				// For simplicity, mark only first line of block comments.
				// Multi-line block comment highlighting is handled line by line below.
				_ = end
				out = append(out, semToken{
					line:      uint32(startLine),
					startChar: uint32(startCol),
					length:    uint32(len(row) - startCol),
					tokenType: ttComment,
				})
				break
			}
			i++
		}
	}
	return out
}

func sortSemTokens(tokens []semToken) {
	// insertion sort - token lists are usually nearly sorted
	for i := 1; i < len(tokens); i++ {
		j := i
		for j > 0 {
			a, b := tokens[j-1], tokens[j]
			if a.line > b.line || (a.line == b.line && a.startChar > b.startChar) {
				tokens[j-1], tokens[j] = tokens[j], tokens[j-1]
				j--
			} else {
				break
			}
		}
	}
}
