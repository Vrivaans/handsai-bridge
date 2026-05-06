package main

import (
	"github.com/ivanv/invok-go-bridge/internal/config"
	"github.com/ivanv/invok-go-bridge/internal/mcp"
)

func main() {
	backendURL := config.LoadConfig()
	apiToken := config.GetAPIToken()

	server := mcp.NewServer(backendURL, apiToken)
	server.Run()
}
