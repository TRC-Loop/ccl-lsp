package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentDefinition(_ *glsp.Context, params *protocol.DefinitionParams) (any, error) {
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

	sym, ok := doc.Symbols.Globals[word]
	if !ok {
		return nil, nil
	}

	// stdlib modules have no source location
	if sym.URI == "" {
		return nil, nil
	}

	defLine := uint32(sym.Line - 1)
	defChar := uint32(sym.Col - 1)
	if sym.Line < 1 {
		defLine = 0
	}
	if sym.Col < 1 {
		defChar = 0
	}

	return &protocol.Location{
		URI: protocol.DocumentUri(sym.URI),
		Range: protocol.Range{
			Start: protocol.Position{Line: defLine, Character: defChar},
			End:   protocol.Position{Line: defLine, Character: defChar + uint32(len(sym.Name))},
		},
	}, nil
}
