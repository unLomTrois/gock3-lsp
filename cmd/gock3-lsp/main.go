package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/handler"
	lsp "github.com/sourcegraph/go-lsp"
)

// Server encapsulates the state and handlers for the language server.
type Server struct {
	jrpcServer *jrpc2.Server
	mutex      sync.RWMutex
	DiagFiles  map[string][]lsp.Diagnostic
	Documents  map[string]string
}

// NewServer initializes a new Server instance with handlers.
func NewServer() *Server {
	s := &Server{
		DiagFiles: make(map[string][]lsp.Diagnostic),
		Documents: make(map[string]string),
	}

	handlers := handler.Map{
		"initialize":              handler.New(s.Initialize),
		"textDocument/completion": handler.New(s.TextDocumentCompletion),
		"textDocument/didOpen":    handler.New(s.TextDocumentDidOpen),
		"textDocument/didClose":   handler.New(s.TextDocumentDidClose),
		"textDocument/didChange":  handler.New(s.TextDocumentDidChange),
		"textDocument/hover":      handler.New(s.TextDocumentHover),
	}

	s.jrpcServer = jrpc2.NewServer(handlers, nil)
	return s
}

// Initialize handles the LSP initialize request.
func (s *Server) Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error) {
	// No shared resources are accessed here, so no mutex is needed.
	capabilities := lsp.ServerCapabilities{
		TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
			Options: &lsp.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    lsp.TDSKIncremental,
			},
		},
		CompletionProvider: &lsp.CompletionOptions{
			ResolveProvider:   false,
			TriggerCharacters: []string{"."},
		},
		HoverProvider: true,
	}

	return lsp.InitializeResult{
		Capabilities: capabilities,
	}, nil
}

// TextDocumentCompletion provides completion items.
func (s *Server) TextDocumentCompletion(ctx context.Context, params lsp.CompletionParams) (lsp.CompletionList, error) {
	// No shared resources are accessed here, so no mutex is needed.
	// Example completion item; extend as needed.
	items := []lsp.CompletionItem{
		{
			Label:         "namespace",
			Kind:          lsp.CIKText,
			Detail:        "Namespace of events",
			Documentation: "https://ck3.paradoxwikis.com/Event_modding",
		},
	}

	return lsp.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// TextDocumentDidOpen handles the event when a text document is opened.
func (s *Server) TextDocumentDidOpen(ctx context.Context, params lsp.DidOpenTextDocumentParams) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s': %v", uri, err)
		return err
	}

	// Store the document content in memory.
	s.Documents[filePath] = params.TextDocument.Text

	// Get diagnostics for the opened file.
	diagnostics := s.GetDiagnostics(filePath)
	s.DiagFiles[filePath] = diagnostics

	// Publish diagnostics to the client.
	return s.publishDiagnostics(ctx, uri, diagnostics)
}

// TextDocumentDidChange handles the event when a text document is changed.
func (s *Server) TextDocumentDidChange(ctx context.Context, params lsp.DidChangeTextDocumentParams) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s': %v", uri, err)
		return err
	}

	// Apply changes to the document content in memory.
	if len(params.ContentChanges) == 0 {
		return nil // No changes to apply.
	}

	change := params.ContentChanges[0]
	s.Documents[filePath] = change.Text

	// Get updated diagnostics.
	diagnostics := s.GetDiagnostics(filePath)
	s.DiagFiles[filePath] = diagnostics

	// Publish updated diagnostics.
	return s.publishDiagnostics(ctx, uri, diagnostics)
}

// TextDocumentDidClose handles the event when a text document is closed.
func (s *Server) TextDocumentDidClose(ctx context.Context, params lsp.DidCloseTextDocumentParams) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s': %v", uri, err)
		return err
	}

	// Remove diagnostics and document content.
	delete(s.DiagFiles, filePath)
	delete(s.Documents, filePath)

	return nil
}

// TextDocumentHover provides hover information at a given position.
func (s *Server) TextDocumentHover(ctx context.Context, params lsp.TextDocumentPositionParams) (lsp.Hover, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s': %v", uri, err)
		return lsp.Hover{}, err
	}

	content, exists := s.Documents[filePath]
	if !exists {
		errMsg := "Document does not exist for URI: " + string(uri)
		log.Println(errMsg)
		return lsp.Hover{}, errors.New(errMsg)
	}

	// Get the specific line.
	lines := strings.Split(content, "\n")
	if params.Position.Line >= len(lines) {
		return lsp.Hover{}, nil // Line out of range.
	}
	lineContent := lines[params.Position.Line]

	// Extract the word at the given character position.
	word, err := extractWord(lineContent, params.Position.Character)
	if err != nil {
		return lsp.Hover{}, nil // No word found.
	}

	// Example hover information; extend as needed.
	hoverContent := "Information about: " + word

	// Define the range for the hover.
	hoverRange := &lsp.Range{
		Start: lsp.Position{
			Line:      params.Position.Line,
			Character: params.Position.Character - len(word),
		},
		End: lsp.Position{
			Line:      params.Position.Line,
			Character: params.Position.Character,
		},
	}

	return lsp.Hover{
		Contents: []lsp.MarkedString{{
			Language: "plaintext",
			Value:    hoverContent,
		}},
		Range: hoverRange,
	}, nil
}

// Start runs the language server.
func (s *Server) Start() error {
	s.jrpcServer.Start(channel.Header("")(os.Stdin, os.Stdout))
	return s.jrpcServer.Wait()
}

// publishDiagnostics sends diagnostics to the client.
func (s *Server) publishDiagnostics(ctx context.Context, uri lsp.DocumentURI, diagnostics []lsp.Diagnostic) error {
	// No shared resources are accessed here, so no mutex is needed.
	params := lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}
	if err := s.jrpcServer.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
		log.Printf("Failed to publish diagnostics: %v", err)
		return err
	}
	return nil
}

// GetDiagnostics generates diagnostics for a given file.
// TODO: Implement actual diagnostic logic.
func (s *Server) GetDiagnostics(filePath string) []lsp.Diagnostic {
	// Placeholder: Return no diagnostics.
	return []lsp.Diagnostic{}
}

// uriToFilePath converts a file URI to a local file path.
func uriToFilePath(uri lsp.DocumentURI) (string, error) {
	if !strings.HasPrefix(string(uri), "file://") {
		return "", errors.New("unsupported URI scheme")
	}
	return strings.TrimPrefix(string(uri), "file://"), nil
}

// extractWord extracts the word at the given character position.
func extractWord(line string, character int) (string, error) {
	if character > len(line) {
		return "", errors.New("character position out of range")
	}

	// Find the start and end of the word at the given position.
	start := character
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	end := character
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return "", errors.New("no word found at position")
	}

	return line[start:end], nil
}

// isWordChar checks if a byte is part of a word.
func isWordChar(b byte) bool {
	return ('a' <= b && b <= 'z') ||
		('A' <= b && b <= 'Z') ||
		('0' <= b && b <= '9') ||
		b == '_'
}

func main() {
	server := NewServer()
	log.Println("Starting Language Server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
