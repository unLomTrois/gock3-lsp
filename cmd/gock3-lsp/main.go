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
	log.Println("Initialize request received.")

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

	log.Println("Initialization complete. Server capabilities set.")
	return lsp.InitializeResult{
		Capabilities: capabilities,
	}, nil
}

// TextDocumentCompletion provides completion items.
func (s *Server) TextDocumentCompletion(ctx context.Context, params lsp.CompletionParams) (lsp.CompletionList, error) {
	log.Printf("Completion request received for URI: %s at position Line %d, Character %d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Example completion item; extend as needed.
	items := []lsp.CompletionItem{
		{
			Label:         "namespace",
			Kind:          lsp.CIKText,
			Detail:        "Namespace of events",
			Documentation: "https://ck3.paradoxwikis.com/Event_modding",
		},
	}

	log.Printf("Returning %d completion items.", len(items))
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
		log.Printf("Invalid URI '%s' in DidOpen: %v", uri, err)
		return err
	}

	log.Printf("Opening document: %s", filePath)

	// Store the document content in memory.
	s.Documents[filePath] = params.TextDocument.Text
	log.Printf("Stored content for document: %s (Length: %d characters)", filePath, len(params.TextDocument.Text))

	// Get diagnostics for the opened file.
	diagnostics := s.GetDiagnostics(filePath)
	s.DiagFiles[filePath] = diagnostics
	log.Printf("Generated %d diagnostics for document: %s", len(diagnostics), filePath)

	// Publish diagnostics to the client.
	if err := s.publishDiagnostics(ctx, uri, diagnostics); err != nil {
		log.Printf("Failed to publish diagnostics for document: %s", filePath)
		return err
	}
	log.Printf("Published diagnostics for document: %s", filePath)
	return nil
}

// TextDocumentDidChange handles the event when a text document is changed.
func (s *Server) TextDocumentDidChange(ctx context.Context, params lsp.DidChangeTextDocumentParams) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s' in DidChange: %v", uri, err)
		return err
	}

	log.Printf("Changing document: %s", filePath)

	// Apply changes to the document content in memory.
	if len(params.ContentChanges) == 0 {
		log.Printf("No content changes provided for document: %s", filePath)
		return nil // No changes to apply.
	}

	change := params.ContentChanges[0]
	previousLength := len(s.Documents[filePath])
	s.Documents[filePath] = change.Text
	newLength := len(change.Text)
	log.Printf("Applied change to document: %s (Previous Length: %d, New Length: %d)", filePath, previousLength, newLength)

	// Get updated diagnostics.
	diagnostics := s.GetDiagnostics(filePath)
	s.DiagFiles[filePath] = diagnostics
	log.Printf("Generated %d updated diagnostics for document: %s", len(diagnostics), filePath)

	// Publish updated diagnostics.
	if err := s.publishDiagnostics(ctx, uri, diagnostics); err != nil {
		log.Printf("Failed to publish updated diagnostics for document: %s", filePath)
		return err
	}
	log.Printf("Published updated diagnostics for document: %s", filePath)
	return nil
}

// TextDocumentDidClose handles the event when a text document is closed.
func (s *Server) TextDocumentDidClose(ctx context.Context, params lsp.DidCloseTextDocumentParams) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s' in DidClose: %v", uri, err)
		return err
	}

	log.Printf("Closing document: %s", filePath)

	// Remove diagnostics and document content.
	delete(s.DiagFiles, filePath)
	delete(s.Documents, filePath)
	log.Printf("Removed diagnostics and content for document: %s", filePath)

	return nil
}

// TextDocumentHover provides hover information at a given position.
func (s *Server) TextDocumentHover(ctx context.Context, params lsp.TextDocumentPositionParams) (lsp.Hover, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	uri := params.TextDocument.URI
	filePath, err := uriToFilePath(uri)
	if err != nil {
		log.Printf("Invalid URI '%s' in Hover: %v", uri, err)
		return lsp.Hover{}, err
	}

	log.Printf("Hover request for document: %s at Line %d, Character %d", uri, params.Position.Line, params.Position.Character)

	content, exists := s.Documents[filePath]
	if !exists {
		errMsg := "Document does not exist for URI: " + string(uri)
		log.Println(errMsg)
		return lsp.Hover{}, errors.New(errMsg)
	}

	// Get the specific line.
	lines := strings.Split(content, "\n")
	if params.Position.Line >= len(lines) {
		log.Printf("Hover position out of range in document: %s", filePath)
		return lsp.Hover{}, nil // Line out of range.
	}
	lineContent := lines[params.Position.Line]

	// Extract the word at the given character position.
	word, err := extractWord(lineContent, params.Position.Character)
	if err != nil {
		log.Printf("No word found at hover position in document: %s", filePath)
		return lsp.Hover{}, nil // No word found.
	}

	log.Printf("Extracted word for hover: '%s' in document: %s", word, filePath)

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

	log.Printf("Providing hover information for word: '%s' in document: %s", word, filePath)

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
	log.Println("Starting Language Server...")
	s.jrpcServer.Start(channel.Header("")(os.Stdin, os.Stdout))
	log.Println("Language Server started successfully.")
	return s.jrpcServer.Wait()
}

// publishDiagnostics sends diagnostics to the client.
func (s *Server) publishDiagnostics(ctx context.Context, uri lsp.DocumentURI, diagnostics []lsp.Diagnostic) error {
	// No shared resources are accessed here, so no mutex is needed.
	log.Printf("Publishing %d diagnostics for URI: %s", len(diagnostics), uri)
	params := lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}
	if err := s.jrpcServer.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
		log.Printf("Failed to publish diagnostics for URI: %s - Error: %v", uri, err)
		return err
	}
	log.Printf("Diagnostics published successfully for URI: %s", uri)
	return nil
}

// GetDiagnostics generates diagnostics for a given file.
// TODO: Implement actual diagnostic logic.
func (s *Server) GetDiagnostics(filePath string) []lsp.Diagnostic {
	// Placeholder: Return no diagnostics.
	log.Printf("Generating diagnostics for document: %s (Placeholder implementation)", filePath)
	return []lsp.Diagnostic{}
}

// uriToFilePath converts a file URI to a local file path.
func uriToFilePath(uri lsp.DocumentURI) (string, error) {
	if !strings.HasPrefix(string(uri), "file://") {
		return "", errors.New("unsupported URI scheme")
	}
	filePath := strings.TrimPrefix(string(uri), "file://")
	log.Printf("Converted URI '%s' to file path '%s'", uri, filePath)
	return filePath, nil
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
	// Set up logging to include date and time.
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server := NewServer()
	log.Println("Initializing Language Server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
