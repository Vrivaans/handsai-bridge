package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// BridgeConfig holds the configurable settings for the bridge.
// It is read from a config.json file in the same directory as the binary.
// Example config.json:
//
//	{ "handsaiUrl": "http://localhost:8080" }
type BridgeConfig struct {
	HandsaiUrl string `json:"handsaiUrl"`
}

// LoadConfig reads the config.json file and returns the HandsAI Backend URL.
func LoadConfig() string {
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

// GetAPIToken retrieves the API token for HandsAI from the environment.
func GetAPIToken() string {
	return os.Getenv("HANDSAI_TOKEN")
}
