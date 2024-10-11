package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/gopherd/core/text/document"
	"github.com/mkideal/next/cmd/nextls/internal/protocol"
	"github.com/sourcegraph/jsonrpc2"
)

// Server is the language server.
type Server struct {
	documents sync.Map
}

// UpdateDocument updates the document in the server.
func (s *Server) UpdateDocument(doc *document.Document) {
	s.documents.Store(doc.Path(), doc)
}

// GetDocument returns the document from the server.
func (s *Server) GetDocument(uri string) *document.Document {
	v, ok := s.documents.Load(document.URIToPath(uri))
	if !ok {
		return nil
	}
	return v.(*document.Document)
}

// RemoveDocument removes the document from the server.
func (s *Server) RemoveDocument(uri string) {
	s.documents.Delete(document.URIToPath(uri))
}

func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic", "error", r, "stack", debug.Stack())
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	slog.Debug("received request", "method", req.Method)
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req)
	case "textDocument/didOpen":
		return s.handleDidOpen(ctx, req)
	case "textDocument/didClose":
		return s.handleDidClose(ctx, req)
	case "textDocument/didChange":
		return s.handleDidChange(ctx, req)
	case "textDocument/semanticTokens/full":
		return s.handleSemanticTokens(ctx, req)
	default:
		return nil, nil
	}
}

func (s *Server) handleInitialize(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	return map[string]any{
		"capabilities": &protocol.ServerCapabilities{
			SemanticTokensProvider: &protocol.SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenTypes:     protocol.TokenTypes(),
					TokenModifiers: protocol.TokenModifiers(),
				},
				Full: true,
			},
			TextDocumentSync: 1,
		},
	}, nil
}

func (s *Server) handleDidOpen(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params struct {
		TextDocument protocol.TextDocumentItem
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	s.UpdateDocument(document.NewDocument(params.TextDocument.URI, params.TextDocument.Text))
	return nil, nil
}

func (s *Server) handleDidClose(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params struct {
		TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	s.RemoveDocument(params.TextDocument.URI)
	return nil, nil
}

func (s *Server) handleDidChange(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params struct {
		TextDocument   protocol.TextDocumentIdentifier           `json:"textDocument"`
		ContentChanges []protocol.TextDocumentContentChangeEvent `json:"contentChanges"`
	}

	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	doc := s.GetDocument(params.TextDocument.URI)
	if doc == nil {
		return nil, fmt.Errorf("document not found in memory store")
	}

	for _, change := range params.ContentChanges {
		if change.Range == nil {
			doc.ApplyChanges([]document.ChangeEvent{
				{Text: change.Text},
			})
			s.UpdateDocument(doc)
		} else {
			doc.ApplyChanges([]document.ChangeEvent{
				{
					Range: &document.Range{
						Start: document.Position{
							Line:      change.Range.Start.Line,
							Character: change.Range.Start.Character,
						},
						End: document.Position{
							Line:      change.Range.End.Line,
							Character: change.Range.End.Character,
						},
					},
					Text: change.Text,
				},
			})
			s.UpdateDocument(doc)
		}
	}

	return nil, nil
}

func (s *Server) handleSemanticTokens(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params struct {
		TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	}

	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	doc := s.GetDocument(params.TextDocument.URI)
	if doc == nil {
		return nil, fmt.Errorf("document not found in memory store")
	}

	handler := protocol.Lookup(doc.Extension())
	if handler == nil {
		return nil, fmt.Errorf("handler not found for %s", doc.Filename())
	}

	return handler.HandleSemanticTokens(ctx, s, doc)
}
