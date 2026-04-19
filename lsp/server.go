package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

const serverName = "noname-lsp"

var handler protocol.Handler

func main() {
	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		TextDocumentDidOpen:    textDocumentDidOpen,
		TextDocumentDidChange:  textDocumentDidChange,
		TextDocumentDidClose:   textDocumentDidClose,
		TextDocumentHover:      textDocumentHover,
		TextDocumentCompletion: textDocumentCompletion,
		TextDocumentDefinition: textDocumentDefinition,
	}

	s := server.NewServer(&handler, serverName, false)
	s.RunStdio()
}

// initialize is called once when the editor first connects.
// We tell the editor what features our server supports.
func initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	// Full sync: Neovim sends the entire file content on every change.
	// Without this it sends only the edited range, which our server doesn't handle.
	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull

	// Tell the editor we support hover and go to definition.
	capabilities.HoverProvider = true
	capabilities.DefinitionProvider = true

	// Tell the editor we support completion.
	// TriggerCharacters defines which characters cause the list to pop up automatically.
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{".", " "},
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name: serverName,
		},
	}, nil
}

// initialized is a notification from the editor — it means "I got your
// capabilities, we're ready to go". No response needed.
func initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(ctx *glsp.Context) error {
	return nil
}
