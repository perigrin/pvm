// ABOUTME: Type information query functionality for LSP
// ABOUTME: Provides detailed type information and symbol analysis

package lsp

import (
	"fmt"
)

// TypeQuery represents a query for type information
type TypeQuery struct {
	URI      string   `json:"uri"`
	Position Position `json:"position"`
	Symbol   string   `json:"symbol,omitempty"`
}

// TypeInfo represents detailed type information
type TypeInfo struct {
	Symbol        string         `json:"symbol"`
	Type          string         `json:"type"`
	Kind          string         `json:"kind"` // variable, function, class, etc.
	Documentation string         `json:"documentation,omitempty"`
	Location      *Location      `json:"location,omitempty"`
	Signature     *FunctionSig   `json:"signature,omitempty"`
	Properties    []PropertyInfo `json:"properties,omitempty"`
	Methods       []MethodInfo   `json:"methods,omitempty"`
	Examples      []string       `json:"examples,omitempty"`
}

// FunctionSig represents a function signature
type FunctionSig struct {
	Parameters []ParameterInfo `json:"parameters"`
	ReturnType string          `json:"returnType,omitempty"`
}

// ParameterInfo represents function parameter information
type ParameterInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PropertyInfo represents object property information
type PropertyInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MethodInfo represents method information
type MethodInfo struct {
	Name      string       `json:"name"`
	Signature *FunctionSig `json:"signature,omitempty"`
}

// TypeQueryService provides type query functionality
type TypeQueryService struct {
	server *Server
}

// NewTypeQueryService creates a new type query service
func NewTypeQueryService(server *Server) *TypeQueryService {
	return &TypeQueryService{
		server: server,
	}
}

// QueryTypeAtPosition queries type information at a specific position
func (s *TypeQueryService) QueryTypeAtPosition(query TypeQuery) (*TypeInfo, error) {
	// TODO: Implement using language service
	return nil, fmt.Errorf("type queries not yet implemented with language service")
}

// QuerySymbol queries type information for a specific symbol
func (s *TypeQueryService) QuerySymbol(uri, symbol string) (*TypeInfo, error) {
	// TODO: Implement using language service
	return nil, fmt.Errorf("symbol queries not yet implemented with language service")
}

// GetAvailableSymbols returns all available symbols in a document
func (s *TypeQueryService) GetAvailableSymbols(uri string) ([]TypeInfo, error) {
	// TODO: Implement using language service
	return nil, fmt.Errorf("symbol listing not yet implemented with language service")
}
