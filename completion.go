package main

import (
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var keywords = []string{
	"var", "const", "function", "return", "if", "else", "for", "in", "while",
	"import", "from", "as", "and", "or", "not", "true", "false",
	"break", "continue", "class", "extends", "self", "super",
	"public", "private", "try", "catch", "throw", "with",
	"range", "fixed",
}

var typeKeywords = []string{
	"string", "int", "sint", "float", "bool", "list", "array", "dict",
}

func textDocumentCompletion(_ *glsp.Context, params *protocol.CompletionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	doc := store.Get(uri)
	if doc == nil {
		return nil, nil
	}

	line := int(params.Position.Line)
	char := int(params.Position.Character)

	// Check if we are after a dot for member completion.
	if prefix, ok := dotPrefix(doc.Content, line, char); ok {
		return memberCompletion(doc, prefix), nil
	}

	word := wordAtPos(doc.Content, line, char)
	return generalCompletion(doc, word), nil
}

// dotPrefix checks if the cursor is right after "identifier." and returns the identifier.
func dotPrefix(content string, line, char int) (string, bool) {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return "", false
	}
	row := lines[line]
	pos := char
	if pos > len(row) {
		pos = len(row)
	}
	if pos == 0 || row[pos-1] != '.' {
		return "", false
	}
	// walk back over the identifier before the dot
	end := pos - 1
	start := end
	for start > 0 && isIdentChar(rune(row[start-1])) {
		start--
	}
	if start == end {
		return "", false
	}
	return row[start:end], true
}

func memberCompletion(doc *Document, objName string) []protocol.CompletionItem {
	if doc.Symbols == nil {
		return nil
	}

	// stdlib module methods
	if methods, ok := stdlibMethods[objName]; ok {
		return methodItems(methods)
	}

	// user-defined symbols
	sym, ok := doc.Symbols.Globals[objName]
	if !ok {
		return nil
	}

	switch sym.Kind {
	case KindModule:
		if methods, ok := stdlibMethods[sym.Name]; ok {
			return methodItems(methods)
		}
	case KindClass:
		var items []protocol.CompletionItem
		for _, f := range sym.Fields {
			kind := protocol.CompletionItemKindField
			items = append(items, protocol.CompletionItem{
				Label:  f.Name,
				Kind:   &kind,
				Detail: strPtr(f.TypeName),
			})
		}
		items = append(items, methodItems(sym.Methods)...)
		// inherit parent
		if sym.SuperName != "" {
			items = append(items, memberCompletion(doc, sym.SuperName)...)
		}
		return items
	default:
		// instance variable - use declared type for builtin method completion
		if methods, ok := builtinMethods[sym.TypeName]; ok {
			return methodItems(methods)
		}
	}

	return nil
}

func methodItems(methods []MethodInfo) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, len(methods))
	for i, m := range methods {
		kind := protocol.CompletionItemKindMethod
		items[i] = protocol.CompletionItem{
			Label:  m.Name,
			Kind:   &kind,
			Detail: strPtr(formatMethodSignature(m)),
		}
	}
	return items
}

func generalCompletion(doc *Document, word string) []protocol.CompletionItem {
	var items []protocol.CompletionItem

	// keywords
	for _, kw := range keywords {
		if strings.HasPrefix(kw, word) {
			kind := protocol.CompletionItemKindKeyword
			items = append(items, protocol.CompletionItem{
				Label: kw,
				Kind:  &kind,
			})
		}
	}

	// type keywords
	for _, t := range typeKeywords {
		if strings.HasPrefix(t, word) {
			kind := protocol.CompletionItemKindKeyword
			items = append(items, protocol.CompletionItem{
				Label:  t,
				Kind:   &kind,
				Detail: strPtr("type"),
			})
		}
	}

	if doc.Symbols == nil {
		return items
	}

	// user-defined globals
	for name, sym := range doc.Symbols.Globals {
		if !strings.HasPrefix(name, word) {
			continue
		}
		kind := symCompletionKind(sym.Kind)
		items = append(items, protocol.CompletionItem{
			Label:  name,
			Kind:   &kind,
			Detail: strPtr(formatSignature(sym)),
		})
	}

	// stdlib module names (not already added as globals)
	for name := range stdlibModules {
		if !strings.HasPrefix(name, word) {
			continue
		}
		if _, exists := doc.Symbols.Globals[name]; exists {
			continue
		}
		kind := protocol.CompletionItemKindModule
		items = append(items, protocol.CompletionItem{
			Label: name,
			Kind:  &kind,
		})
	}

	return items
}

func symCompletionKind(k SymbolKind) protocol.CompletionItemKind {
	switch k {
	case KindFunction:
		return protocol.CompletionItemKindFunction
	case KindClass:
		return protocol.CompletionItemKindClass
	case KindModule:
		return protocol.CompletionItemKindModule
	case KindConst:
		return protocol.CompletionItemKindConstant
	default:
		return protocol.CompletionItemKindVariable
	}
}
