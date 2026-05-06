package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultInvokURL = "https://useinvok.run/"

// BridgeConfig holds the configurable settings for the bridge.
// It is read from a config.json file in the same directory as the binary.
// Example config.json:
//
//	{ "invokUrl": "https://useinvok.run/" }
type BridgeConfig struct {
	InvokUrl string `json:"invokUrl"`
}

// LoadConfig reads the config.json file and returns the Invok Backend URL.
func LoadConfig() string {
	// Look for config.json next to the binary
	exe, err := os.Executable()
	if err != nil {
		return defaultInvokURL
	}
	configPath := filepath.Join(filepath.Dir(exe), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config.json found — use default
		return defaultInvokURL
	}

	var cfg BridgeConfig
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.InvokUrl == "" {
		return defaultInvokURL
	}
	return cfg.InvokUrl
}

// GetAPIToken retrieves the API token for Invok from the environment.
func GetAPIToken() string {
	return os.Getenv("INVOK_TOKEN")
}
