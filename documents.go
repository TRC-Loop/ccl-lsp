package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
)

type ParseError struct {
	Message string
	Line    int
	Col     int
}

type ParseResult struct {
	Tokens []lexer.Token
	AST    *parser.Program
	Errors []ParseError
}

type Document struct {
	URI     string
	Content string
	Parsed  *ParseResult
	Symbols *Analysis
}

type DocumentStore struct {
	mu   sync.RWMutex
	docs map[string]*Document
}

var store = &DocumentStore{docs: make(map[string]*Document)}

func (s *DocumentStore) Update(uri, content string) *Document {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc := &Document{URI: uri, Content: content}
	doc.Parsed = parseDoc(content)
	doc.Symbols = analyzeAST(uri, doc.Parsed.AST)
	s.docs[uri] = doc
	return doc
}

func (s *DocumentStore) Get(uri string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs[uri]
}

func (s *DocumentStore) Delete(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.docs, uri)
}

func parseDoc(content string) *ParseResult {
	result := &ParseResult{}

	l := lexer.New(content)
	tokens, err := l.Tokenize()
	if err != nil {
		result.Errors = append(result.Errors, errorFromMsg(err.Error()))
		return result
	}
	result.Tokens = tokens

	p := parser.New(tokens)
	ast, err := p.Parse()
	if err != nil {
		result.Errors = append(result.Errors, errorFromMsg(err.Error()))
		return result
	}
	result.AST = ast
	return result
}

// errorFromMsg parses "line L:C: message" from ccolon error strings.
func errorFromMsg(msg string) ParseError {
	pe := ParseError{Message: msg, Line: 1, Col: 1}
	// format: "line L:C: text"
	if !strings.HasPrefix(msg, "line ") {
		return pe
	}
	rest := msg[5:]
	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return pe
	}
	lineStr := rest[:colonIdx]
	rest = rest[colonIdx+1:]
	colonIdx2 := strings.Index(rest, ":")
	if colonIdx2 < 0 {
		return pe
	}
	colStr := rest[:colonIdx2]
	pe.Message = strings.TrimSpace(rest[colonIdx2+1:])

	if l, err := strconv.Atoi(strings.TrimSpace(lineStr)); err == nil {
		pe.Line = l
	}
	if c, err := strconv.Atoi(strings.TrimSpace(colStr)); err == nil {
		pe.Col = c
	}
	return pe
}

// wordAtPos returns the identifier word at the given 0-indexed line/character position.
func wordAtPos(content string, line, char int) string {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return ""
	}
	row := lines[line]
	if char > len(row) {
		char = len(row)
	}
	start := char
	for start > 0 && isIdentChar(rune(row[start-1])) {
		start--
	}
	end := char
	for end < len(row) && isIdentChar(rune(row[end])) {
		end++
	}
	return row[start:end]
}

func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// tokenAtPos finds the token covering 0-indexed line/char.
func tokenAtPos(tokens []lexer.Token, line, char int) *lexer.Token {
	lsp0line := line + 1
	lsp0char := char + 1
	for i := range tokens {
		t := &tokens[i]
		if t.Line == lsp0line && t.Col <= lsp0char && lsp0char <= t.Col+len(t.Literal) {
			return t
		}
	}
	return nil
}

// lineCount returns the number of lines in content.
func lineCount(content string) int {
	return strings.Count(content, "\n") + 1
}

// lastLineLen returns the length of the last line.
func lastLineLen(content string) int {
	idx := strings.LastIndex(content, "\n")
	if idx < 0 {
		return len(content)
	}
	return len(content) - idx - 1
}

func sprintf(f string, args ...any) string {
	return fmt.Sprintf(f, args...)
}
