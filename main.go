package main

import (
	"github.com/ivanv/handsai-go-bridge/internal/config"
	"github.com/ivanv/handsai-go-bridge/internal/mcp"
)

func main() {
	backendURL := config.LoadConfig()
	apiToken := config.GetAPIToken()

	server := mcp.NewServer(backendURL, apiToken)
	server.Run()
}
