package protocol

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gopherd/core/text/document"
)

// PositionFor returns the line and column for a position in a document.
func PositionFor(pos int, doc *document.Document) (line, col int) {
	lineInfo, err := doc.OffsetToPosition(pos)
	if err != nil {
		slog.Error("Failed to calculate line and column", "error", err)
		text := doc.Content()[:pos]
		col = strings.LastIndex(text, "\n")
		if col == -1 {
			col = utf8.RuneCountInString(text)
		} else {
			col++ // After the newline.
			col = utf8.RuneCountInString(text[col:])
		}
		line = strings.Count(text, "\n")
		return line, col
	}
	return lineInfo.Line, lineInfo.Character
}

// DocumentManager is the interface that must be implemented by a document manager.
type DocumentManager interface {
	UpdateDocument(doc *document.Document)
	GetDocument(uri string) *document.Document
	RemoveDocument(uri string)
}

// Handler is the interface that must be implemented by a language server handler.
type Handler interface {
	// HandleSemanticTokens handles a "textDocument/semanticTokens/full" request.
	HandleSemanticTokens(context.Context, DocumentManager, *document.Document) (any, error)
}

// handlers maps file extension to handler.
var handlersMu sync.RWMutex
var handlers = make(map[string]Handler)

// Register registers a handler for a file extension.
func Register(extension string, handler Handler) {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	if _, dup := handlers[extension]; dup {
		panic("lsp: Register called twice for extension " + extension)
	}
	handlers[extension] = handler
}

// Lookup returns the handler for a file extension.
func Lookup(extension string) Handler {
	handlersMu.RLock()
	defer handlersMu.RUnlock()
	return handlers[extension]
}
