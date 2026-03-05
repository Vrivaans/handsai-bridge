package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Server encapsulates the MCP bridge functionality
type Server struct {
	BackendURL string
	APIToken   string
}

// NewServer creates a new MCP bridge server
func NewServer(backendURL, apiToken string) *Server {
	return &Server{
		BackendURL: backendURL,
		APIToken:   apiToken,
	}
}

// Run starts the STDIN scanning loop
func (s *Server) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		s.handleLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v\n", err)
		os.Exit(1)
	}
}

func (s *Server) handleLine(line []byte) {
	var req McpRequest
	if err := json.Unmarshal(line, &req); err != nil {
		s.sendError(nil, -32700, "Parse error", err.Error())
		return
	}

	if req.Jsonrpc != "2.0" {
		s.sendError(req.Id, -32600, "Invalid Request: jsonrpc must be '2.0'", nil)
		return
	}

	switch req.Method {
	case "initialize":
		s.sendResponse(req.Id, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "HandsAI Go MCP Bridge",
				"version": "1.0.0",
			},
		})
	case "notifications/initialized":
		// Do nothing
	case "tools/list":
		s.handleToolsList(req.Id)
	case "tools/call":
		s.handleToolsCall(req.Id, req.Params)
	default:
		s.sendError(req.Id, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

func (s *Server) handleToolsList(id interface{}) {
	req, err := http.NewRequest("GET", s.BackendURL+"/mcp/tools/list", nil)
	if err != nil {
		s.sendError(id, -32603, "Internal error", err.Error())
		return
	}
	if s.APIToken != "" {
		req.Header.Set("X-HandsAI-Token", s.APIToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.sendError(id, -32603, "Internal error", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.sendError(id, -32603, fmt.Sprintf("HTTP %d", resp.StatusCode), nil)
		return
	}

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	// Default empty result
	var tools interface{} = []interface{}{}

	if res, ok := data["result"].(map[string]interface{}); ok {
		if t, ok := res["tools"]; ok {
			tools = t
		}
	} else if t, ok := data["tools"]; ok {
		tools = t
	}

	// Inject the virtual sync tool
	if toolsSlice, ok := tools.([]interface{}); ok {
		tools = append(toolsSlice, map[string]interface{}{
			"name":        "handsai_sync_tools",
			"description": "Fuerza una actualización de las Herramientas y Proveedores cacheados enviando una notificación MCP, sin necesidad de reiniciar el servidor.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		})
	}

	s.sendResponse(id, map[string]interface{}{
		"tools": tools,
	})
}

func (s *Server) handleToolsCall(id interface{}, params json.RawMessage) {
	if params == nil {
		s.sendError(id, -32602, "Invalid params", nil)
		return
	}

	var pMap map[string]interface{}
	json.Unmarshal(params, &pMap)

	name, _ := pMap["name"].(string)
	if name == "" {
		s.sendError(id, -32602, "Invalid params: tool name is required", nil)
		return
	}

	if name == "handsai_sync_tools" {
		// Send the list_changed notification to the MCP client
		notification := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notifications/tools/list_changed",
		}
		out, _ := json.Marshal(notification)
		fmt.Println(string(out))

		s.sendResponse(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "¡Caché de MCP invalidado exitosamente! El cliente de Inteligencia Artificial acaba de ser notificado para que descargue la nueva lista de herramientas de HandsAI automáticamente."},
			},
		})
		return
	}

	args, _ := pMap["arguments"]
	if args == nil {
		args = map[string]interface{}{}
	}

	mcpCall := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "internal-go-1",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	reqBody, _ := json.Marshal(mcpCall)

	req, err := http.NewRequest("POST", s.BackendURL+"/mcp/tools/call", bytes.NewBuffer(reqBody))
	if err != nil {
		s.sendError(id, -32603, "Internal error", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.APIToken != "" {
		req.Header.Set("X-HandsAI-Token", s.APIToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.sendError(id, -32603, "Execution error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	var result interface{}
	if r, ok := data["result"]; ok {
		result = r
	} else if c, ok := data["content"]; ok {
		result = map[string]interface{}{"content": c}
	} else {
		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(body)},
			},
		}
	}

	s.sendResponse(id, result)
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	if id == nil {
		return // Notification, no reply
	}
	resp := McpResponse{
		Jsonrpc: "2.0",
		Id:      id,
		Result:  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}

func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := McpResponse{
		Jsonrpc: "2.0",
		Id:      id,
		Error: &McpError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}
