package main

import (
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentHover(_ *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := string(params.TextDocument.URI)
	doc := store.Get(uri)
	if doc == nil || doc.Symbols == nil {
		return nil, nil
	}

	line := int(params.Position.Line)
	char := int(params.Position.Character)
	word := wordAtPos(doc.Content, line, char)
	if word == "" {
		return nil, nil
	}

	// Check for module.method hover: if the token before the word is ".", resolve the object.
	if obj, method, ok := dotContext(doc.Content, line, char, word); ok {
		return hoverForMember(doc, obj, method), nil
	}

	// Keyword hover
	if info, ok := keywordDocs[word]; ok {
		return markdownHover(info), nil
	}

	// Symbol hover
	sym, ok := doc.Symbols.Globals[word]
	if !ok {
		return nil, nil
	}
	return markdownHover(sprintf("```ccl\n%s\n```", formatSignature(sym))), nil
}

// dotContext checks if the word is after a dot and returns (object, method, true).
func dotContext(content string, line, char int, word string) (string, string, bool) {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return "", "", false
	}
	row := lines[line]
	// find the start of word in row near char
	start := char - len(word)
	if start < 0 {
		return "", "", false
	}
	if start == 0 || row[start-1] != '.' {
		return "", "", false
	}
	objEnd := start - 1
	objStart := objEnd
	for objStart > 0 && isIdentChar(rune(row[objStart-1])) {
		objStart--
	}
	if objStart == objEnd {
		return "", "", false
	}
	return row[objStart:objEnd], word, true
}

func hoverForMember(doc *Document, obj, method string) *protocol.Hover {
	// stdlib
	if methods, ok := stdlibMethods[obj]; ok {
		for _, m := range methods {
			if m.Name == method {
				return markdownHover(sprintf("```ccl\n%s.%s\n```", obj, formatMethodSignature(m)))
			}
		}
	}

	if doc.Symbols == nil {
		return nil
	}
	sym, ok := doc.Symbols.Globals[obj]
	if !ok {
		return nil
	}

	switch sym.Kind {
	case KindModule:
		if methods, ok := stdlibMethods[sym.Name]; ok {
			for _, m := range methods {
				if m.Name == method {
					return markdownHover(sprintf("```ccl\n%s.%s\n```", obj, formatMethodSignature(m)))
				}
			}
		}
	case KindClass:
		for _, f := range sym.Fields {
			if f.Name == method {
				return markdownHover(sprintf("```ccl\n%s %s\n```", f.TypeName, f.Name))
			}
		}
		for _, m := range sym.Methods {
			if m.Name == method {
				return markdownHover(sprintf("```ccl\n%s\n```", formatMethodSignature(m)))
			}
		}
	default:
		if methods, ok := builtinMethods[sym.TypeName]; ok {
			for _, m := range methods {
				if m.Name == method {
					return markdownHover(sprintf("```ccl\n%s\n```", formatMethodSignature(m)))
				}
			}
		}
	}
	return nil
}

func markdownHover(content string) *protocol.Hover {
	kind := protocol.MarkupKindMarkdown
	return &protocol.Hover{
		Contents: protocol.MarkupContent{Kind: kind, Value: content},
	}
}

var keywordDocs = map[string]string{
	"var":      "`var type name = value` - mutable variable declaration",
	"const":    "`const type name = value` - immutable constant declaration",
	"function": "`function name(params) returnType { ... }` - function declaration",
	"return":   "`return [expr]` - return from a function",
	"if":       "`if (cond) { ... } else { ... }` - conditional",
	"else":     "else branch of an if statement",
	"for":      "`for x in iterable { ... }` - iterate over list, array, dict or range",
	"in":       "used with for-in loops",
	"while":    "`while (cond) { ... }` - loop while condition is true",
	"import":   "`import module` or `import module as alias` - import a stdlib module",
	"from":     "`from module import name1, name2` - selective import",
	"as":       "alias in import or resource binding in with",
	"and":      "logical AND (short-circuits)",
	"or":       "logical OR (short-circuits)",
	"not":      "logical NOT",
	"true":     "boolean literal true",
	"false":    "boolean literal false",
	"break":    "exit the innermost loop",
	"continue": "skip to the next loop iteration",
	"class":    "`class Name [extends Parent] { ... }` - class declaration",
	"extends":  "inherit from a parent class",
	"self":     "reference to the current class instance",
	"super":    "`super.method(args)` - call a parent class method",
	"public":   "public field or method visibility modifier",
	"private":  "private field or method visibility modifier",
	"try":      "`try { ... } catch (Type e) { ... }` - exception handling",
	"catch":    "catch block for exceptions",
	"throw":    "`throw expr` - raise an exception",
	"with":     "`with expr as name { ... }` - resource management (calls .close() on exit)",
	"range":    "`range(n)` or `range(start, end)` - integer range for for-in loops",
	"fixed":    "`fixed([...])` - create a fixed-size array",
	"string":   "built-in string type",
	"int":      "built-in 64-bit signed integer type",
	"sint":     "built-in arbitrary precision integer type",
	"float":    "built-in 64-bit float type",
	"bool":     "built-in boolean type",
	"list":     "built-in dynamic list type",
	"array":    "built-in fixed-size array type",
	"dict":     "built-in dictionary type (string keys)",
}
