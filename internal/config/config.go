package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultHandsaiURL = "http://invok-invok-bbthdh-2b633b-157-254-174-169.traefik.me/"

// BridgeConfig holds the configurable settings for the bridge.
// It is read from a config.json file in the same directory as the binary.
// Example config.json:
//
//	{ "handsaiUrl": "http://invok-invok-tikua9-2431f7-157-254-174-169.traefik.me/" }
type BridgeConfig struct {
	HandsaiUrl string `json:"handsaiUrl"`
}

// LoadConfig reads the config.json file and returns the HandsAI Backend URL.
func LoadConfig() string {
	// Look for config.json next to the binary
	exe, err := os.Executable()
	if err != nil {
		return defaultHandsaiURL
	}
	configPath := filepath.Join(filepath.Dir(exe), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config.json found — use default
		return defaultHandsaiURL
	}

	var cfg BridgeConfig
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.HandsaiUrl == "" {
		return defaultHandsaiURL
	}
	return cfg.HandsaiUrl
}

// GetAPIToken retrieves the API token for HandsAI from the environment.
func GetAPIToken() string {
	return os.Getenv("HANDSAI_TOKEN")
}
