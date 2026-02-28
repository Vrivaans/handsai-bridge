package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Config holds the configurable settings for the bridge.
// It is read from a config.json file in the same directory as the binary.
// Example config.json:
//
//	{ "handsaiUrl": "http://localhost:8080" }
type BridgeConfig struct {
	HandsaiUrl string `json:"handsaiUrl"`
}

func loadConfig() string {
	// Look for config.json next to the binary
	exe, err := os.Executable()
	if err != nil {
		return "http://localhost:8080"
	}
	configPath := filepath.Join(filepath.Dir(exe), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config.json found — use default
		return "http://localhost:8080"
	}

	var cfg BridgeConfig
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.HandsaiUrl == "" {
		return "http://localhost:8080"
	}
	return cfg.HandsaiUrl
}

var backendURL = loadConfig()

type McpRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type McpResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *McpError   `json:"error,omitempty"`
}

type McpError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Accept optional 'mcp' subcommand to match how Antigravity IDE registers servers
	// e.g.: handsai-mcp mcp  (argument is accepted but ignored)
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		handleLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v\n", err)
		os.Exit(1)
	}
}

func handleLine(line []byte) {
	var req McpRequest
	if err := json.Unmarshal(line, &req); err != nil {
		sendError(nil, -32700, "Parse error", err.Error())
		return
	}

	if req.Jsonrpc != "2.0" {
		sendError(req.Id, -32600, "Invalid Request: jsonrpc must be '2.0'", nil)
		return
	}

	switch req.Method {
	case "initialize":
		sendResponse(req.Id, map[string]interface{}{
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
		handleToolsList(req.Id)
	case "tools/call":
		handleToolsCall(req.Id, req.Params)
	default:
		sendError(req.Id, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

func handleToolsList(id interface{}) {
	resp, err := http.Get(backendURL + "/mcp/tools/list")
	if err != nil {
		sendError(id, -32603, "Internal error", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		sendError(id, -32603, fmt.Sprintf("HTTP %d", resp.StatusCode), nil)
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

	sendResponse(id, map[string]interface{}{
		"tools": tools,
	})
}

func handleToolsCall(id interface{}, params json.RawMessage) {
	if params == nil {
		sendError(id, -32602, "Invalid params", nil)
		return
	}

	var pMap map[string]interface{}
	json.Unmarshal(params, &pMap)

	name, _ := pMap["name"].(string)
	if name == "" {
		sendError(id, -32602, "Invalid params: tool name is required", nil)
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

		sendResponse(id, map[string]interface{}{
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
	resp, err := http.Post(backendURL+"/mcp/tools/call", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		sendError(id, -32603, "Execution error", err.Error())
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

	sendResponse(id, result)
}

func sendResponse(id interface{}, result interface{}) {
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

func sendError(id interface{}, code int, message string, data interface{}) {
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
