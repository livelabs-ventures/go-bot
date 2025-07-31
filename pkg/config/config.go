package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Repository    string `json:"repository"`
	Name          string `json:"name"`
	LocalRepoPath string `json:"localRepoPath"`
}

// Manager handles configuration operations
type Manager struct {
	configDir  string
	configFile string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".standup-bot")
	configFile := filepath.Join(configDir, "config.json")

	return &Manager{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

// ConfigDir returns the configuration directory path
func (m *Manager) ConfigDir() string {
	return m.configDir
}

// Load reads the configuration from disk
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Config doesn't exist yet
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand tilde in local repo path
	if cfg.LocalRepoPath != "" && cfg.LocalRepoPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.LocalRepoPath = filepath.Join(homeDir, cfg.LocalRepoPath[1:])
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (m *Manager) Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert absolute path to tilde notation for config file
	saveCfg := *cfg
	homeDir, err := os.UserHomeDir()
	if err == nil && filepath.HasPrefix(saveCfg.LocalRepoPath, homeDir) {
		saveCfg.LocalRepoPath = "~" + saveCfg.LocalRepoPath[len(homeDir):]
	}

	data, err := json.MarshalIndent(saveCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if configuration file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configFile)
	return err == nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LocalRepoPath: "~/.standup-bot/repo",
	}
}