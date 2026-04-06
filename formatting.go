package main

import (
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/ccolon/formatter"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentFormatting(_ *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	uri := string(params.TextDocument.URI)
	doc := store.Get(uri)
	if doc == nil {
		return nil, nil
	}

	// Load .ccolonfmt config from the file's directory if present.
	dir := filepath.Dir(strings.TrimPrefix(uri, "file://"))
	cfg := formatter.LoadConfig(dir)

	formatted, err := formatter.Format(doc.Content, cfg)
	if err != nil {
		// Don't surface formatter errors as LSP errors - just return no edits.
		return nil, nil
	}

	if formatted == doc.Content {
		return nil, nil
	}

	// Replace the entire document.
	lines := uint32(lineCount(doc.Content))
	lastLen := uint32(lastLineLen(doc.Content))

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: lines - 1, Character: lastLen},
			},
			NewText: formatted,
		},
	}, nil
}
