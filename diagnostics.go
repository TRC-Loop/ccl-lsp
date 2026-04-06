package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func publishDiagnostics(ctx *glsp.Context, uri string, doc *Document) {
	diags := buildDiagnostics(doc)
	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentUri(uri),
		Diagnostics: diags,
	})
}

func buildDiagnostics(doc *Document) []protocol.Diagnostic {
	if doc.Parsed == nil || len(doc.Parsed.Errors) == 0 {
		return []protocol.Diagnostic{}
	}

	sev := protocol.DiagnosticSeverityError
	diags := make([]protocol.Diagnostic, 0, len(doc.Parsed.Errors))

	for _, e := range doc.Parsed.Errors {
		line := uint32(e.Line - 1)
		col := uint32(e.Col - 1)
		if e.Line < 1 {
			line = 0
		}
		if e.Col < 1 {
			col = 0
		}
		diags = append(diags, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + 1},
			},
			Severity: &sev,
			Source:   strPtr("ccl-lsp"),
			Message:  e.Message,
		})
	}
	return diags
}

func strPtr(s string) *string { return &s }
