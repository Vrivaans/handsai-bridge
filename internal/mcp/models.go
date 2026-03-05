package mcp

import "encoding/json"

// McpRequest represents an incoming JSON-RPC 2.0 request.
type McpRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// McpResponse represents an outgoing JSON-RPC 2.0 response.
type McpResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *McpError   `json:"error,omitempty"`
}

// McpError represents an error in an outgoing JSON-RPC 2.0 response.
type McpError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
