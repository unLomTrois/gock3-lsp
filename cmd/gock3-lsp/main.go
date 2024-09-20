package main

import (
	"context"
	"log"
	"os"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/handler"
	lsp "github.com/sourcegraph/go-lsp"
)

func Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error) {
	capabilities := lsp.ServerCapabilities{
		TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
			Options: &lsp.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1,
			}},
		CompletionProvider: &lsp.CompletionOptions{
			ResolveProvider:   false,
			TriggerCharacters: []string{"."},
		},
	}

	return lsp.InitializeResult{
		Capabilities: capabilities,
	}, nil
}

func TextDocumentCompletion(ctx context.Context, params lsp.CompletionParams) (lsp.CompletionList, error) {
	return lsp.CompletionList{
		IsIncomplete: false,
		Items: []lsp.CompletionItem{
			{
				Label:         "namespace",
				Data:          1,
				Kind:          lsp.CIKText,
				Detail:        "namespace of events",
				Documentation: "https://ck3.paradoxwikis.com/Event_modding",
			},
		},
	}, nil
}

func main() {
	handlers := handler.Map{
		"initialize":              handler.New(Initialize),
		"textDocument/completion": handler.New(TextDocumentCompletion),
	}

	server := jrpc2.NewServer(handlers, nil)

	server.Start(channel.Header("")(os.Stdin, os.Stdout))

	if err := server.Wait(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
