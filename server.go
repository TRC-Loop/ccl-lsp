package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspServer "github.com/tliron/glsp/server"
)

const serverName = "ccl-lsp"
const serverVersion = "0.1.0"

var handler protocol.Handler

func runServer() {
	handler.Initialize = initialize
	handler.Initialized = func(_ *glsp.Context, _ *protocol.InitializedParams) error { return nil }
	handler.Shutdown = func(_ *glsp.Context) error { return nil }

	handler.TextDocumentDidOpen = textDocumentDidOpen
	handler.TextDocumentDidChange = textDocumentDidChange
	handler.TextDocumentDidClose = textDocumentDidClose

	handler.TextDocumentCompletion = textDocumentCompletion
	handler.TextDocumentHover = textDocumentHover
	handler.TextDocumentDefinition = textDocumentDefinition
	handler.TextDocumentFormatting = textDocumentFormatting
	handler.TextDocumentSemanticTokensFull = textDocumentSemanticTokensFull

	srv := glspServer.NewServer(&handler, serverName, false)
	srv.RunStdio()
}

func initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	syncKind := protocol.TextDocumentSyncKindFull
	triggerChars := []string{"."}
	trueVal := true
	ver := serverVersion

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: &trueVal,
				Change:    &syncKind,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: triggerChars,
			},
			HoverProvider:              &trueVal,
			DefinitionProvider:         &trueVal,
			DocumentFormattingProvider: &trueVal,
			SemanticTokensProvider: &protocol.SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenTypes:     semanticTokenTypes,
					TokenModifiers: semanticTokenModifiers,
				},
				Full: true,
			},
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &ver,
		},
	}, nil
}

func textDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	doc := store.Update(uri, params.TextDocument.Text)
	publishDiagnostics(ctx, uri, doc)
	return nil
}

func textDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}
	uri := string(params.TextDocument.URI)
	change := params.ContentChanges[len(params.ContentChanges)-1]
	content, ok := change.(protocol.TextDocumentContentChangeEventWhole)
	if !ok {
		return nil
	}
	doc := store.Update(uri, content.Text)
	publishDiagnostics(ctx, uri, doc)
	return nil
}

func textDocumentDidClose(_ *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	store.Delete(string(params.TextDocument.URI))
	return nil
}

func boolPtr(b bool) *bool { return &b }
