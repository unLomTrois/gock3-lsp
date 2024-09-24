package main

import (
	"context"
	"log"
	"os"
	"strings"

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

var tempFile *os.File
var DiagsFiles = make(map[string][]lsp.Diagnostic)
var Server *jrpc2.Server

func TextDocumentDidOpen(ctx context.Context, vs lsp.DidOpenTextDocumentParams) error {
	fileURL := strings.Replace(string(vs.TextDocument.URI), "file://", "", 1)
	DiagsFiles[fileURL] = GetDiagnostics(fileURL, fileURL)

	TextDocumentPublishDiagnostics(Server, ctx, lsp.PublishDiagnosticsParams{
		URI:         vs.TextDocument.URI,
		Diagnostics: DiagsFiles[fileURL],
	})
	tempFile.Write([]byte(vs.TextDocument.Text))
	return nil
}

func TextDocumentPublishDiagnostics(server *jrpc2.Server, ctx context.Context, vs lsp.PublishDiagnosticsParams) error {
	return server.Notify(ctx, "textDocument/publishDiagnostics", vs)
}

func GetDiagnostics(fileURL string, file string) []lsp.Diagnostic {
	var diags []lsp.Diagnostic
	return diags
}

func TextDocumentDidChange(ctx context.Context, vs lsp.DidChangeTextDocumentParams) error {
	tempFile.Truncate(0)
	tempFile.Seek(0, 0)
	tempFile.Write([]byte(vs.ContentChanges[0].Text))

	fileURL := strings.Replace(string(vs.TextDocument.URI), "file://", "", 1)
	DiagsFiles[fileURL] = GetDiagnostics(tempFile.Name(), fileURL)
	TextDocumentPublishDiagnostics(Server, ctx, lsp.PublishDiagnosticsParams{
		URI:         vs.TextDocument.URI,
		Diagnostics: DiagsFiles[fileURL],
	})
	return nil
}

func TextDocumentDidClose(ctx context.Context, vs lsp.DidCloseTextDocumentParams) error {
	return nil
}

func TextDocumentHover(ctx context.Context, vs lsp.TextDocumentPositionParams) (lsp.Hover, error) {
	pos := vs.Position
	character := pos.Character
	line := pos.Line

	// read temp file
	tempFile.Seek(0, 0)
	contents := make([]byte, 0)
	tempFile.Read(contents)
	contents = contents[:len(contents)-1]

	// now get it by pos line
	// get the line
	lineContents := strings.Split(string(contents), "\n")[line]
	// get the character
	characterContents := strings.Split(lineContents, " ")[character:3]

	rang := &lsp.Range{
		Start: lsp.Position{Line: line, Character: character},
		End: lsp.Position{
			Line:      line,
			Character: character + 1,
		},
	}

	return lsp.Hover{
		Contents: []lsp.MarkedString{{
			Language: "ck3",
			Value:    strings.Join(characterContents, " "),
		}},
		Range: rang,
	}, nil
}

func main() {
	handlers := handler.Map{
		"initialize":              handler.New(Initialize),
		"textDocument/completion": handler.New(TextDocumentCompletion),
		"textDocument/didOpen":    handler.New(TextDocumentDidOpen),
		"textDocument/didClose":   handler.New(TextDocumentDidClose),
		"textDocument/didChange":  handler.New(TextDocumentDidChange),

		"textDocument/hover": handler.New(TextDocumentHover),
	}

	server := jrpc2.NewServer(handlers, nil)

	server.Start(channel.Header("")(os.Stdin, os.Stdout))

	if err := server.Wait(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
