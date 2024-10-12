// Package protocol provides definitions for the Language Server Protocol (LSP).
// It includes structures and constants for common LSP operations and types.
package protocol

import (
	"encoding/json"
)

func ToJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`      // Line position in a document (zero-based).
	Character int `json:"character"` // Character offset on a line in a document (zero-based).
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"` // The range's start position.
	End   Position `json:"end"`   // The range's end position.
}

// Location represents a location inside a resource, such as a line inside a text file.
type Location struct {
	URI   string `json:"uri"`   // The text document's URI.
	Range Range  `json:"range"` // The location's range.
}

// TextDocumentIdentifier identifies a text document using its URI.
type TextDocumentIdentifier struct {
	URI string `json:"uri"` // The text document's URI.
}

// TextDocumentItem represents an open text document.
type TextDocumentItem struct {
	URI        string `json:"uri"`        // The text document's URI.
	LanguageID string `json:"languageId"` // The text document's language identifier.
	Version    int    `json:"version"`    // The version number of this document (it will increase after each change, including undo/redo).
	Text       string `json:"text"`       // The content of the opened text document.
}

// VersionedTextDocumentIdentifier identifies a specific version of a text document.
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"` // The version number of this document.
}

// TextDocumentContentChangeEvent represents a change to a text document.
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`       // The range of the document that changed.
	RangeLength int    `json:"rangeLength,omitempty"` // The length of the range that got replaced.
	Text        string `json:"text"`                  // The new text of the range/document.
}

// The following constants define method names for LSP requests and notifications.
const (
	Initialize                     = "initialize"
	Initialized                    = "initialized"
	Shutdown                       = "shutdown"
	Exit                           = "exit"
	TextDocumentDidOpen            = "textDocument/didOpen"
	TextDocumentDidChange          = "textDocument/didChange"
	TextDocumentDidSave            = "textDocument/didSave"
	TextDocumentDidClose           = "textDocument/didClose"
	TextDocumentCompletion         = "textDocument/completion"
	TextDocumentHover              = "textDocument/hover"
	TextDocumentDefinition         = "textDocument/definition"
	TextDocumentReferences         = "textDocument/references"
	TextDocumentDocumentSymbol     = "textDocument/documentSymbol"
	TextDocumentFormatting         = "textDocument/formatting"
	TextDocumentRename             = "textDocument/rename"
	TextDocumentCodeAction         = "textDocument/codeAction"
	TextDocumentSemanticTokensFull = "textDocument/semanticTokens/full"
	TextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
)

// InitializeParams represents the parameters sent in an initialize request.
type InitializeParams struct {
	ProcessID             int                `json:"processId"`             // The process ID of the parent process that started the server.
	RootURI               string             `json:"rootUri"`               // The rootUri of the workspace.
	InitializationOptions any                `json:"initializationOptions"` // User provided initialization options.
	Capabilities          ClientCapabilities `json:"capabilities"`          // The capabilities provided by the client (editor or tool).
}

// ClientCapabilities represents the capabilities provided by the client.
type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Window       WindowClientCapabilities       `json:"window,omitempty"`
}

// WorkspaceClientCapabilities represents the workspace-specific client capabilities.
type WorkspaceClientCapabilities struct {
	ApplyEdit bool `json:"applyEdit,omitempty"`
	// Add more workspace capabilities as needed
}

// TextDocumentClientCapabilities represents the text document-specific client capabilities.
type TextDocumentClientCapabilities struct {
	Synchronization TextDocumentSyncClientCapabilities `json:"synchronization,omitempty"`
	Completion      CompletionClientCapabilities       `json:"completion,omitempty"`
	Hover           HoverClientCapabilities            `json:"hover,omitempty"`
	// Add more text document capabilities as needed
}

// TextDocumentSyncClientCapabilities represents the client capabilities for document synchronization.
type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

// CompletionClientCapabilities represents the client capabilities for completion.
type CompletionClientCapabilities struct {
	DynamicRegistration bool                             `json:"dynamicRegistration,omitempty"`
	CompletionItem      CompletionItemClientCapabilities `json:"completionItem,omitempty"`
}

// CompletionItemClientCapabilities represents the client capabilities for completion items.
type CompletionItemClientCapabilities struct {
	SnippetSupport          bool     `json:"snippetSupport,omitempty"`
	CommitCharactersSupport bool     `json:"commitCharactersSupport,omitempty"`
	DocumentationFormat     []string `json:"documentationFormat,omitempty"`
	DeprecatedSupport       bool     `json:"deprecatedSupport,omitempty"`
	PreselectSupport        bool     `json:"preselectSupport,omitempty"`
}

// HoverClientCapabilities represents the client capabilities for hover.
type HoverClientCapabilities struct {
	DynamicRegistration bool     `json:"dynamicRegistration,omitempty"`
	ContentFormat       []string `json:"contentFormat,omitempty"`
}

// WindowClientCapabilities represents the window-specific client capabilities.
type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
}

// ServerCapabilities represents the capabilities provided by the language server.
type ServerCapabilities struct {
	TextDocumentSync           int                    `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionOptions     `json:"completionProvider,omitempty"`
	HoverProvider              bool                   `json:"hoverProvider,omitempty"`
	DefinitionProvider         bool                   `json:"definitionProvider,omitempty"`
	ReferencesProvider         bool                   `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider     bool                   `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider    bool                   `json:"workspaceSymbolProvider,omitempty"`
	CodeActionProvider         bool                   `json:"codeActionProvider,omitempty"`
	DocumentFormattingProvider bool                   `json:"documentFormattingProvider,omitempty"`
	RenameProvider             bool                   `json:"renameProvider,omitempty"`
	SemanticTokensProvider     *SemanticTokensOptions `json:"semanticTokensProvider,omitempty"`
}

// CompletionOptions represents the options for completion support.
type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// SemanticTokensOptions represents the options for semantic tokens support.
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Range  bool                 `json:"range,omitempty"`
	Full   bool                 `json:"full,omitempty"`
}

// SemanticTokensLegend represents the legend for semantic tokens.
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// InitializeResult represents the result of an initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"` // The capabilities the language server provides.
}

// CompletionParams represents the parameters of a textDocument/completion request.
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext represents additional information about the context in which a completion request is triggered.
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind represents how a completion was triggered.
type CompletionTriggerKind int

const (
	CompletionTriggerInvoked                         CompletionTriggerKind = 1
	CompletionTriggerTriggerCharacter                CompletionTriggerKind = 2
	CompletionTriggerTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionList represents the result of a completion request.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion item.
type CompletionItem struct {
	Label               string             `json:"label"`
	Kind                CompletionItemKind `json:"kind,omitempty"`
	Detail              string             `json:"detail,omitempty"`
	Documentation       json.RawMessage    `json:"documentation,omitempty"`
	Deprecated          bool               `json:"deprecated,omitempty"`
	Preselect           bool               `json:"preselect,omitempty"`
	SortText            string             `json:"sortText,omitempty"`
	FilterText          string             `json:"filterText,omitempty"`
	InsertText          string             `json:"insertText,omitempty"`
	InsertTextFormat    InsertTextFormat   `json:"insertTextFormat,omitempty"`
	TextEdit            *TextEdit          `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit         `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string           `json:"commitCharacters,omitempty"`
	Command             *Command           `json:"command,omitempty"`
	Data                any                `json:"data,omitempty"`
}

// CompletionItemKind represents the kind of a completion item.
type CompletionItemKind int

// InsertTextFormat represents the format of the insert text.
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// TextEdit represents a text edit applicable to a text document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Command represents a reference to a command.
type Command struct {
	Title     string `json:"title"`
	Command   string `json:"command"`
	Arguments []any  `json:"arguments,omitempty"`
}

// TextDocumentPositionParams represents a text document and a position inside that document.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               any                            `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	Tags               []int                          `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

// DiagnosticRelatedInformation represents a related piece of information for a diagnostic.
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// PublishDiagnosticsParams represents the parameters of a textDocument/publishDiagnostics notification.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// MarkupContent represents a string value with a specific content type.
type MarkupContent struct {
	// Kind is the type of the MarkupContent.
	// It can be "plaintext" or "markdown".
	Kind string `json:"kind"`

	// Value is the actual content value.
	// This will be the text shown to the user, with the format depending on the "kind".
	Value string `json:"value"`
}

// HoverParams represents the parameters of a textDocument/hover request.
type HoverParams struct {
	TextDocumentPositionParams
}

// Hover represents the result of a hover request.
type Hover struct {
	Contents json.RawMessage `json:"contents"`
	Range    *Range          `json:"range,omitempty"`
}

// DefinitionParams represents the parameters of a textDocument/definition request.
type DefinitionParams struct {
	TextDocumentPositionParams
}

// ReferenceParams represents the parameters of a textDocument/references request.
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext represents the context of a reference request.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// DocumentSymbolParams represents the parameters of a textDocument/documentSymbol request.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentFormattingParams represents the parameters of a textDocument/formatting request.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions represents options for formatting.
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// RenameParams represents the parameters of a textDocument/rename request.
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

// CodeActionParams represents the parameters of a textDocument/codeAction request.
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext represents the context of a code action request.
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Only        []string     `json:"only,omitempty"`
}

// SemanticTokensParams represents the parameters of a textDocument/semanticTokens/full request.
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens represents the result of a textDocument/semantic
// SemanticTokens represents the result of a textDocument/semanticTokens/full request.
type SemanticTokensResult struct {
	// The result id. If provided and clients support delta updating, this id
	// can be sent in a semantic token request to request delta updates.
	ResultID string `json:"resultId,omitempty"`

	// The actual tokens. For a detailed description of the format please refer
	// to the LSP spec at:
	// https://microsoft.github.io/language-server-protocol/specifications/specification-3-17/#textDocument_semanticTokens
	Data []uint32 `json:"data"`
}
